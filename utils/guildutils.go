// File: codebase/utils/guildutils.go
package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type MemberData struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
	JoinedAt       string `json:"joined_at"`
	IsOnline       bool   `json:"is_online"`
}

func GetGuildIDByTag(tag string) (string, bool, error) {
	var guildID string
	found, err := QueryRow("GetGuildIDByTag", &guildID,
		"SELECT guild_id FROM guilds WHERE tag = $1", tag)
	return guildID, found, err
}

func IsUserInGuild(guildID, userID string) (bool, error) {
	var exists bool
	_, err := QueryRow("CheckGuildMembership", &exists,
		"SELECT EXISTS(SELECT 1 FROM guild_members WHERE guild_id = $1 AND user_id = $2)",
		guildID, userID)
	return exists, err
}

func HandleGuildNotFound(c echo.Context, preferHTML bool) error {
	if preferHTML {
		return c.Render(200, "guild_not_found.html", nil)
	}
	return SendErrorResponse(c, ErrGuildNotFound)
}

func ValidateGuildExists(c echo.Context, guildID string, preferHTML bool) (map[string]string, error) {
	guild, found, err := GetGuildByID(guildID)
	if err != nil {
		return nil, SendErrorResponse(c, ErrDatabaseError)
	}
	if !found {
		return nil, HandleGuildNotFound(c, preferHTML)
	}
	return guild, nil
}

func GetValidUserID(userID string) (string, error) {
	var exists bool
	found, err := QueryRow("CheckUserExists", &exists,
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userID)

	if err != nil {
		return "", err
	}

	if !found || !exists {
		return "", fmt.Errorf("user ID %s does not exist", userID)
	}

	return userID, nil
}
func AddGuildMember(guildID, userID string) error {
	tx, err := GetDB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO guild_members (guild_id, user_id) VALUES ($1, $2)", guildID, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE guilds SET member_count = member_count + 1 WHERE guild_id = $1", guildID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err == nil {
		cache.Provider.Delete(fmt.Sprintf("guild:%s", guildID))
	}
	return err
}

func GetGuildByID(guildID string) (map[string]string, bool, error) {
	guild, err := CacheFirstQuery(
		fmt.Sprintf("guild:%s", guildID),
		cache.DefaultConfig.DefaultTTL,
		func() (map[string]string, bool, error) {
			rows, err := GetDB().Query("SELECT guild_id, name, description, owner_id, created_at, profile_picture_url, member_count FROM guilds WHERE guild_id = $1", guildID)
			if err != nil {
				return nil, false, err
			}
			defer rows.Close()

			if !rows.Next() {
				return nil, false, nil
			}

			var id, name, description, ownerID, createdAt string
			var memberCount int
			var profilePictureURL sql.NullString
			err = rows.Scan(&id, &name, &description, &ownerID, &createdAt, &profilePictureURL, &memberCount)
			if err != nil {
				return nil, false, err
			}

			guild := map[string]string{
				"guild_id":            id,
				"name":                name,
				"description":         description,
				"owner_id":            ownerID,
				"created_at":          createdAt,
				"profile_picture_url": profilePictureURL.String,
				"member_count":        strconv.Itoa(memberCount),
			}

			return guild, true, nil
		})

	if err != nil {
		return nil, false, err
	}
	return guild, guild != nil, nil
}

func GetUserGuilds(userID string) ([]map[string]interface{}, error) {
	guilds := []map[string]interface{}{}

	query := `
	SELECT g.guild_id, g.name, g.description, g.owner_id, g.created_at, g.profile_picture_url, gm.joined_at
	FROM guilds g
	JOIN guild_members gm ON g.guild_id = gm.guild_id
	WHERE gm.user_id = $1
	ORDER BY g.created_at DESC
	`

	rows, err := GetDB().Query(query, userID)
	if err != nil {
		return guilds, err
	}
	defer rows.Close()

	for rows.Next() {
		var guildID, name, description, ownerID, createdAt, joinedAt string
		var profilePictureURL sql.NullString
		err := rows.Scan(&guildID, &name, &description, &ownerID, &createdAt, &profilePictureURL, &joinedAt)
		if err != nil {
			continue
		}

		guild := map[string]interface{}{
			"guild_id":    guildID,
			"name":        name,
			"description": description,
			"owner_id":    ownerID,
			"created_at":  createdAt,
			"joined_at":   joinedAt,
		}

		if profilePictureURL.Valid {
			guild["profile_picture_url"] = profilePictureURL.String
		}

		guilds = append(guilds, guild)
	}

	return guilds, rows.Err()
}

func InsertGuild(guildID, name, description, ownerID string, profilePicturePath sql.NullString) error {
	var query string
	var args []interface{}

	if profilePicturePath.Valid {
		query = "INSERT INTO guilds (guild_id, name, description, owner_id, profile_picture_url, member_count) VALUES ($1, $2, $3, $4, $5, $6)"
		args = []interface{}{guildID, name, description, ownerID, profilePicturePath.String, 0}
	} else {
		query = "INSERT INTO guilds (guild_id, name, description, owner_id, member_count) VALUES ($1, $2, $3, $4, $5)"
		args = []interface{}{guildID, name, description, ownerID, 0}
	}

	return ExecuteQuery("CreateGuild", query, args...)
}

func GetGuildMembersPaginated(guildID string, page, limit int) ([]MemberData, int, error) {
	if page < 1 {
		page = 1
	}
	if page > 10000 {
		page = 10000
	}
	// No backend limits - respect frontend request or use unlimited for internal calls
	if limit > 1000 {
		limit = 1000  // Reasonable safety limit only
	}

	offset := (page - 1) * limit

	tx, err := GetDB().Begin()
	if err != nil {
		return []MemberData{}, 0, err
	}
	defer tx.Rollback()

	var totalCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM guild_members WHERE guild_id = $1", guildID).Scan(&totalCount)
	if err != nil {
		return []MemberData{}, 0, err
	}

	members := []MemberData{}
	var rows *sql.Rows
	if limit == AppConfig.AllMembers {
		rows, err = tx.Query("SELECT gm.user_id, u.username, COALESCE(u.profile_picture, '') as profile_picture, gm.joined_at FROM guild_members gm JOIN users u ON gm.user_id = u.user_id WHERE gm.guild_id = $1 ORDER BY u.username ASC", guildID)
	} else {
		rows, err = tx.Query("SELECT gm.user_id, u.username, COALESCE(u.profile_picture, '') as profile_picture, gm.joined_at FROM guild_members gm JOIN users u ON gm.user_id = u.user_id WHERE gm.guild_id = $1 ORDER BY u.username ASC LIMIT $2 OFFSET $3", guildID, limit, offset)
	}
	if err != nil {
		return members, totalCount, err
	}
	defer rows.Close()

	var onlineMembers []MemberData
	var offlineMembers []MemberData

	for rows.Next() {
		var member MemberData
		if err := rows.Scan(&member.UserID, &member.Username, &member.ProfilePicture, &member.JoinedAt); err != nil {
			continue
		}
		member.IsOnline = IsUserOnline(member.UserID)
		
		if member.IsOnline {
			onlineMembers = append(onlineMembers, member)
		} else {
			offlineMembers = append(offlineMembers, member)
		}
	}

	// Combine: online first, then offline (both already alphabetically sorted from SQL)
	members = append(onlineMembers, offlineMembers...)

	tx.Commit()
	return members, totalCount, nil
}

func BroadcastMemberEvent(guildID, eventType, userID, username string) {
	var profilePicture sql.NullString
	QueryRow("GetUserProfilePicture", &profilePicture,
		"SELECT profile_picture FROM users WHERE user_id = $1", userID)

	profilePictureValue := ""
	if profilePicture.Valid {
		profilePictureValue = profilePicture.String
	}

memberData := map[string]interface{}{
    "type":            eventType,
    "guild_id":        guildID,
    "user_id":         userID,
    "username":        username,
    "profile_picture": profilePictureValue,
}
broadcastData, _ := json.Marshal(memberData)
BroadcastWithRedis(1, broadcastData)
}

func NotifyUserGuildAdded(userID, guildID string) {
	guild, found, err := GetGuildByID(guildID)
	if !found || err != nil {
		return
	}

	guildData := map[string]interface{}{
		"type":  "guild_created",
		"guild": guild,
	}
	broadcastData, _ := json.Marshal(guildData)
	SendToUser(userID, websocket.TextMessage, broadcastData)
}

func NotifyUserGuildRemoved(userID, guildID string) {
	guildData := map[string]interface{}{
		"type":     "guild_removed",
		"guild_id": guildID,
	}
	broadcastData, _ := json.Marshal(guildData)
	SendToUser(userID, websocket.TextMessage, broadcastData)
}

func RequireGuildMembership(c echo.Context, userID, guildID string) error {
	isInGuild, err := IsUserInGuild(guildID, userID)
	if err != nil {
		return SendErrorResponse(c, ErrDatabaseError)
	}
	if !isInGuild {
		return SendErrorResponse(c, ErrUserNotInGuild)
	}
	return nil
}
