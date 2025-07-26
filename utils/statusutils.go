// X code: Status logic scattered across multiple files

// Y code: Centralized status management
package utils

import (
	"encoding/json"
	"fmt"
	"auth.com/v4/internal/websockets"
	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)



// BroadcastUserStatusChange sends status change to all guilds the user is in
func BroadcastUserStatusChange(userID string, isOnline bool) {
	statusText := "offline"
	if isOnline {
		statusText = "online"
	}
	log(logrus.InfoLevel, "Status", "broadcast_start", userID, fmt.Errorf("broadcasting %s status for user %s", statusText, userID))

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "get_guilds_error", userID, err)
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

			err := BroadcastToGuildMembers(guildID, statusData)
			if err != nil {
				log(logrus.ErrorLevel, "Status", "broadcast_error", userID, fmt.Errorf("failed broadcast to guild %s: %v", guildID, err))
			} else {
				broadcastCount++
			}
		}
	}

	log(logrus.InfoLevel, "Status", "broadcast_complete", userID, fmt.Errorf("broadcasted %s status to %d guilds for user %s", statusText, broadcastCount, userID))
}

// SendInitialStatusesToUser sends current online status of all guild members to a newly connected user
func SendInitialStatusesToUser(userID string) {
	log(logrus.InfoLevel, "Status", "send_initial_statuses", userID, fmt.Errorf("sending initial statuses to user %s", userID))

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "get_guilds_error", userID, err)
		return
	}

	totalSent := 0
	for _, guild := range userGuilds {
		if guildID, ok := guild["guild_id"].(string); ok {
			members, _, err := GetGuildMembersPaginated(guildID, 1, AppConfig.MembersPerPage)
			if err == nil {
				for i := range members {
					members[i].IsOnline = websockets.IsUserOnline(members[i].UserID)
				}
			}
			if err != nil {
				log(logrus.ErrorLevel, "Status", "get_members_error", userID, fmt.Errorf("guild %s: %v", guildID, err))
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
					log(logrus.ErrorLevel, "Status", "marshal_error", userID, err)
					continue
				}

				websockets.SendToUser(userID, websocket.TextMessage, statusJSON)
				guildSent++
				totalSent++
			}
			log(logrus.DebugLevel, "Status", "guild_statuses_sent", userID, fmt.Errorf("sent %d statuses for guild %s", guildSent, guildID))
		}
	}

	log(logrus.InfoLevel, "Status", "initial_statuses_complete", userID, fmt.Errorf("sent %d total statuses to user %s", totalSent, userID))
}


func ValidateWebSocketSession(userID, sessionID string) (bool, error) {
	connectionData, found, err := cache.Provider.GetWebSocketConnectionData(sessionID)
	if err != nil || !found {
		return false, fmt.Errorf("websocket connection data not found")
	}

	storedToken, exists := connectionData["http_session_token"]
	if !exists || storedToken == "" {
		return false, fmt.Errorf("no session token stored for websocket")
	}

	validatedUserID, isValid, err := ValidateSessionToken(storedToken)
	if err != nil || !isValid {
		return false, err
	}

	if validatedUserID != userID {
		return false, fmt.Errorf("user ID mismatch")
	}

	return true, nil
}


