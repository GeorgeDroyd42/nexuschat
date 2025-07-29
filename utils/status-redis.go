package utils

import (
	"encoding/json"
	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"time"
)

func BroadcastUserStatusChange(userID string, isOnline bool) {
	statusText := "offline"
	if isOnline {
		statusText = "online"
	}
	Log.Info("status", "broadcast_start", "Broadcasting "+statusText+" status for user "+userID, map[string]interface{}{"user_id": userID})

	username, err := GetUsernameByID(userID)
	if err != nil {
		Log.Error("status", "get_username_error", "Failed to get username", err, map[string]interface{}{"user_id": userID})
		return
	}

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		Log.Error("status", "get_guilds_error", "Failed to get user guilds", err, map[string]interface{}{"user_id": userID})
		return
	}

	broadcastCount := 0
	for _, guild := range userGuilds {
		if guildID, ok := guild["guild_id"].(string); ok {
			if isOnline {
				cache.Provider.AddUserToGuildOnline(guildID, userID, username)
			} else {
				cache.Provider.AddUserToGuildOffline(guildID, userID, username)
			}

			statusData := map[string]interface{}{
				"type":      "user_status_changed",
				"user_id":   userID,
				"is_online": isOnline,
				"guild_id":  guildID,
			}

			statusData["guild_id"] = guildID
			messageBytes, _ := json.Marshal(statusData)
			err := cache.Provider.PublishMessage("broadcast", map[string]interface{}{
				"type":       1,
				"data":       messageBytes,
				"channel_id": "",
				"secure":     true,
			})
			if err != nil {
				Log.Error("status", "broadcast_error", "Failed broadcast to guild", err, map[string]interface{}{"user_id": userID, "guild_id": guildID})
			} else {
				broadcastCount++
			}
		}
	}

	Log.Info("status", "broadcast_complete", "Status broadcast completed", map[string]interface{}{"user_id": userID, "status": statusText, "broadcast_count": broadcastCount})
}

func SendInitialStatusesToUser(userID string) {
	Log.Info("status", "send_initial_statuses", "Sending initial statuses to user", map[string]interface{}{"user_id": userID})

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		Log.Error("status", "get_guilds_error", "Failed to get user guilds", err, map[string]interface{}{"user_id": userID})
		return
	}

	totalSent := 0
	for _, guild := range userGuilds {
		if guildID, ok := guild["guild_id"].(string); ok {
			members, _, err := GetGuildMembersPaginated(guildID, 1, AppConfig.MembersPerPage)
			if err == nil {
				for i := range members {
					members[i].IsOnline = IsUserOnlineRedis(members[i].UserID, guildID)
				}
			}
			if err != nil {
				Log.Error("status", "get_members_error", "Failed to get guild members", err, map[string]interface{}{"user_id": userID, "guild_id": guildID})
				continue
			}

			guildSent := 0
			for _, member := range members {
				statusData := map[string]interface{}{
					"type":      "user_status_changed",
					"user_id":   member.UserID,
					"is_online": member.IsOnline,
					"guild_id":  guildID,
				}

				statusJSON, err := json.Marshal(statusData)
				if err != nil {
					Log.Error("status", "marshal_error", "Failed to marshal status data", err, map[string]interface{}{"user_id": userID})
					continue
				}

				SendToUser(userID, websocket.TextMessage, statusJSON)
				guildSent++
				totalSent++
			}
			Log.Debug("status", "guild_statuses_sent", "Guild statuses sent", map[string]interface{}{"user_id": userID, "guild_id": guildID, "statuses_sent": guildSent})
		}
	}

	Log.Info("status", "initial_statuses_complete", "Initial statuses sending completed", map[string]interface{}{"user_id": userID, "total_sent": totalSent})
}

func IsUserOnlineRedis(userID, guildID string) bool {
	onlineUsers, err := cache.Provider.GetGuildOnlineUsers(guildID, 0, 10000)
	if err != nil {
		return IsUserOnline(userID)
	}
	
	for _, onlineUserID := range onlineUsers {
		if onlineUserID == userID {
			return true
		}
	}
	return false
}


func GetOnlineUsersInGuildRedis(guildID string) ([]string, error) {
	return cache.Provider.GetGuildOnlineUsers(guildID, 0, 10000)
}

func EnsureGuildMembersInRedis(guildID string) error {
	members, _, err := GetGuildMembersPaginated(guildID, 1, AppConfig.AllMembers)
	if err != nil {
		return err
	}

	for _, member := range members {
		if IsUserOnline(member.UserID) {
			cache.Provider.AddUserToGuildOnline(guildID, member.UserID, member.Username)
		} else {
			cache.Provider.AddUserToGuildOffline(guildID, member.UserID, member.Username)
		}
	}
	return nil
}

func PopulateRedisOnStartup() error {
	// Check if already populated
	var dummy string
	exists, _ := cache.Provider.Get("redis:populated:v1", &dummy)
	if exists {
		Log.Info("redis", "startup", "Redis already populated, skipping")
		return nil
	}
	
	Log.Info("redis", "startup", "Populating Redis with guild members...")
	
	// Get all guild IDs
	rows, err := GetDB().Query("SELECT DISTINCT guild_id FROM guild_members")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	guildCount := 0
	for rows.Next() {
		var guildID string
		if err := rows.Scan(&guildID); err != nil {
			continue
		}
		
		// Use existing function to populate each guild
		EnsureGuildMembersInRedis(guildID)
		guildCount++
	}
	
	// Mark as populated
	cache.Provider.Set("redis:populated:v1", "true", 24*time.Hour)
	Log.Info("redis", "startup", "Redis population complete", map[string]interface{}{"guilds_populated": guildCount})
	return nil
}