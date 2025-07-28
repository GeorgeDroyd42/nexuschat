package utils

import (
	"encoding/json"
	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

// BroadcastUserStatusChange sends status change to all guilds the user is in
func BroadcastUserStatusChange(userID string, isOnline bool) {
	statusText := "offline"
	if isOnline {
		statusText = "online"
	}
	Log.Info("status", "broadcast_start", "Broadcasting "+statusText+" status for user "+userID, map[string]interface{}{"user_id": userID})

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		Log.Error("status", "get_guilds_error", "Failed to get user guilds", err, map[string]interface{}{"user_id": userID})
		return
	}

	broadcastCount := 0
	for _, guild := range userGuilds {
		if guildID, ok := guild["guild_id"].(string); ok {
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

// SendInitialStatusesToUser sends current online status of all guild members to a newly connected user
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
					members[i].IsOnline = IsUserOnline(members[i].UserID)
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