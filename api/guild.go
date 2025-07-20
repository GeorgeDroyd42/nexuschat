package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"auth.com/v4/cache"
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

func CreateGuildHandler(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	owner, err := utils.RequireUserID(c)

	if err != nil {
		return err
	}

	valid, errCode := utils.ValidateGuildData(name, description)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	guilds, err := utils.GetUserGuilds(owner)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if len(guilds) >= utils.AppConfig.MaxGuildsPerUser {
		return utils.SendErrorResponse(c, utils.ErrMaxGuildsReached)
	}

	guildID := utils.GenerateSessionID("guild")

	// Handle server picture upload
	var profilePicturePath sql.NullString
	file, err := c.FormFile("server_picture")
	if err == nil {
		// Server picture provided
		imageData, outputFormat, valid, errCode := utils.ValidateProfilePicture(file)
		if !valid {
			return utils.SendErrorResponse(c, errCode)
		}

		os.MkdirAll("public/uploads/guilds", 0750)
		filename := fmt.Sprintf("%s_%d.%s", guildID, time.Now().Unix(), outputFormat)
		dst := filepath.Join("public/uploads/guilds", filename)

		err = os.WriteFile(dst, imageData, 0644)
		if err != nil {
			return utils.SendErrorResponse(c, utils.ErrDatabaseError)
		}

		profilePicturePath = sql.NullString{String: "/public/uploads/guilds/" + filename, Valid: true}
	}

	// Create the guild with profile picture
	err = utils.InsertGuild(guildID, name, description, owner, profilePicturePath)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	utils.NotifyUserGuildAdded(owner, guildID)

	// Automatically add the creator as a member
	err = utils.AddGuildMember(guildID, owner)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success":  true,
		"message":  "Guild created successfully",
		"guild_id": guildID,
	})
}

func EditChannelHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	channelID := c.Param("channelid")
	name := c.FormValue("name")
	description := c.FormValue("description")

	guildID, valid, errCode := utils.ValidateChannelOperation(userID, channelID, "", name, description, "edit")
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	// Update the channel
	_, err = utils.GetDB().Exec(
		`UPDATE channels SET name = $1, description = $2 WHERE channel_id = $3`,
		name, description, channelID)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	// Broadcast the channel update event
	utils.BroadcastChannelEvent(guildID, "channel_updated", channelID, name, description)

	return utils.SendSuccessResponse(c, "Channel updated successfully")
}

func GetGuildHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "guildid")
	if err != nil {
		return err
	}

	if err := utils.RequireGuildMembership(c, userID, guildID); err != nil {
		return err
	}

	guild, found, err := utils.GetGuildByID(guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if !found {
		return utils.SendErrorResponse(c, utils.ErrGuildNotFound)
	}

	return c.JSON(200, map[string]interface{}{
		"success":             true,
		"guild_id":            guild["guild_id"],
		"name":                guild["name"],
		"description":         guild["description"],
		"owner_id":            guild["owner_id"],
		"created_at":          guild["created_at"],
		"profile_picture_url": guild["profile_picture_url"],
	})
}

func JoinGuildHandler(c echo.Context) error {
	user, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	id, err := utils.RequireParam(c, "tag")
	if err != nil {
		return err
	}

	guildID := id

	validUser, err := utils.GetValidUserID(user)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	guilds, err := utils.GetUserGuilds(validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if len(guilds) >= utils.AppConfig.MaxGuildsPerUser {
		return utils.SendErrorResponse(c, utils.ErrMaxGuildsReached)
	}

	isInGuild, err := utils.IsUserInGuild(guildID, validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if isInGuild {
		return utils.SendSuccessResponse(c, "User is already a member of this guild")
	}

	err = utils.AddGuildMember(guildID, validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	utils.NotifyUserGuildAdded(validUser, guildID)
	username, err := utils.GetUsernameByID(validUser)
	if err == nil {
		utils.BroadcastMemberEvent(guildID, "member_joined", validUser, username)
	}
	return utils.SendSuccessResponse(c, "Joined guild successfully")
}

func ViewGuildHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "id")
	if err != nil {
		return err
	}

	isInGuild, err := utils.IsUserInGuild(guildID, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if !isInGuild {
		return c.Redirect(302, "/i/"+guildID)
	}

	guild, found, err := utils.GetGuildByID(guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if !found {
		return utils.HandleGuildNotFound(c, true)
	}

	guild["guild_id"] = guildID
	return c.Render(200, "guild.html", guild)
}

func ViewMainHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	username, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	data := map[string]string{
		"name":        "Welcome, " + username + "!",
		"description": "Select a guild from the sidebar or create a new one to get started.",
	}
	return c.Render(200, "guild.html", data)
}

func LeaveGuildHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "id")
	if err != nil {
		return err
	}

	// Use CacheFirstQuery to check if user is actually in the guild
	isInGuild, err := utils.CacheFirstQuery(
		fmt.Sprintf("user_in_guild:%s:%s", userID, guildID),
		5*time.Minute,
		func() (bool, bool, error) {
			exists, err := utils.IsUserInGuild(guildID, userID)
			return exists, true, err // Always "found" since EXISTS returns a boolean
		})

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	if !isInGuild {
		return utils.SendSuccessResponse(c, "You are not a member of this guild")
	}

	tx, err := utils.GetDB().Begin()
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM guild_members WHERE guild_id = $1 AND user_id = $2", guildID, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	_, err = tx.Exec("UPDATE guilds SET member_count = GREATEST(member_count - 1, 0) WHERE guild_id = $1", guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	err = tx.Commit()

	username, _ := utils.GetUsernameByID(userID)
	if err == nil {
		utils.BroadcastMemberEvent(guildID, "member_left", userID, username)
		utils.NotifyUserGuildRemoved(userID, guildID)
	}

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	// Invalidate cache entries
	cache.Provider.Delete(fmt.Sprintf("user_in_guild:%s:%s", userID, guildID))
	cache.Provider.Delete(fmt.Sprintf("guild:%s", guildID))

	return utils.SendSuccessResponse(c, "Left guild successfully")
}

func GetGuildMembersHandler(c echo.Context) error {
	guildID, err := utils.RequireParam(c, "id")
	if err != nil {
		return err
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	if err := utils.RequireGuildMembership(c, userID, guildID); err != nil {
		return err
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 25
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	members, totalCount, err := utils.GetGuildMembersWithStatus(guildID, page, limit)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"members":     members,
		"count":       len(members),
		"total_count": totalCount,
	})
}

func CreateChannelHandler(c echo.Context) error {
	guildID := c.FormValue("guild_id")
	name := c.FormValue("name")
	description := c.FormValue("description")

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, valid, errCode := utils.ValidateChannelOperation(userID, "", guildID, name, description, "create")
	if !valid {

		return utils.SendErrorResponse(c, errCode)
	}

	channelID := utils.GenerateSessionID("channel")

	_, err = utils.GetDB().Exec(
		`INSERT INTO channels (channel_id, guild_id, name, description, created_by) VALUES ($1, $2, $3, $4, $5)`,
		channelID, guildID, name, description, userID)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	utils.BroadcastChannelEvent(guildID, "channel_created", channelID, name, description)

	return utils.SendSuccessResponse(c, "Channel created successfully")
}

func DeleteChannelHandler(c echo.Context) error {
	var req struct {
		ChannelID string `json:"channel_id"`
	}

	if err := c.Bind(&req); err != nil {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	channelID := req.ChannelID

	guildID, valid, errCode := utils.ValidateChannelOperation(userID, channelID, "", "", "", "delete")
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	var name, description string
	err = utils.GetDB().QueryRow(`SELECT name, description FROM channels WHERE channel_id = $1`, channelID).Scan(&name, &description)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrChannelNotFound)
	}

	_, err = utils.GetDB().Exec(`DELETE FROM channels WHERE channel_id = $1`, channelID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	utils.BroadcastChannelEvent(guildID, "channel_deleted", channelID, name, description)

	return utils.SendSuccessResponse(c, "Channel deleted successfully")
}

func GetChannelsHandler(c echo.Context) error {
	guildID := c.QueryParam("guild_id")

	if _, err := utils.RequireUserID(c); err != nil {
		return err
	}

	if guildID == "" {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	rows, err := utils.GetDB().Query(
		`SELECT channel_id, name, description FROM channels WHERE guild_id = $1 ORDER BY created_at`,
		guildID)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var channelID, name, description string
		rows.Scan(&channelID, &name, &description)

		channels = append(channels, map[string]interface{}{
			"channel_id":  channelID,
			"name":        name,
			"description": description,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"channels": channels})
}

func ViewChannelHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "guildid")
	if err != nil {
		return err
	}

	channelID, err := utils.RequireParam(c, "channelid")
	if err != nil {
		return err
	}

	var channel struct {
		ChannelID   string `db:"channel_id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		GuildID     string `db:"guild_id"`
		CreatedAt   string `db:"created_at"`
	}

	err = utils.GetDB().QueryRow(`
		SELECT c.channel_id, c.name, c.description, c.guild_id, c.created_at 
		FROM channels c 
		WHERE c.channel_id = $1`, channelID).Scan(
		&channel.ChannelID, &channel.Name, &channel.Description, &channel.GuildID, &channel.CreatedAt)

	if err != nil {
		return c.Redirect(http.StatusFound, "/v/main")
	}

	isInGuild, err := utils.IsUserInGuild(guildID, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	if !isInGuild {
		return c.Redirect(302, "/i/"+guildID)
	}

	return c.Render(200, "guild.html", map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
		"channel_id":  channel.ChannelID,
		"guild_id":    guildID,
		"created_at":  channel.CreatedAt,
		"is_channel":  true,
	})
}

func GetChannelInfoHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}
	channelID, err := utils.RequireParam(c, "channelid")
	if err != nil {
		return err
	}

	valid, errCode := utils.ValidateChannelAccess(userID, channelID)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	var channelName, guildID, description string
	err = utils.GetDB().QueryRow("SELECT name, guild_id, COALESCE(description, '') FROM channels WHERE channel_id = $1", channelID).Scan(&channelName, &guildID, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendErrorResponse(c, utils.ErrChannelNotFound)
		}
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success":     true,
		"channel_id":  channelID,
		"name":        channelName,
		"guild_id":    guildID,
		"description": description,
	})
}
