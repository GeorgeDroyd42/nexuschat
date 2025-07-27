// X code: Status logic scattered across multiple files

// Y code: Centralized status management
package utils

import (
	"encoding/json"
	"fmt"
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
	logrus.WithFields(logrus.Fields{
	"module":  "status",
	"action":  "broadcast_start", 
	"user_id": userID,
	}).Infof("Broadcasting %s status for user %s", statusText, userID)

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
		"module":  "status",
		"action":  "get_guilds_error",
		"user_id": userID,
	}).WithError(err).Error("Failed to get user guilds")
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
				logrus.WithFields(logrus.Fields{
				"module":   "status",
				"action":   "broadcast_error",
				"user_id":  userID,
				"guild_id": guildID,
			}).WithError(err).Error("Failed broadcast to guild")
			} else {
				broadcastCount++
			}
		}
	}

	logrus.WithFields(logrus.Fields{
	"module":         "status",
	"action":         "broadcast_complete",
	"user_id":        userID,
	"status":         statusText,
	"broadcast_count": broadcastCount,
}).Info("Status broadcast completed")
}

// SendInitialStatusesToUser sends current online status of all guild members to a newly connected user
func SendInitialStatusesToUser(userID string) {
	logrus.WithFields(logrus.Fields{
	"module":  "status",
	"action":  "send_initial_statuses",
	"user_id": userID,
}).Info("Sending initial statuses to user")

	userGuilds, err := GetUserGuilds(userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
		"module":  "status",
		"action":  "get_guilds_error",
		"user_id": userID,
	}).WithError(err).Error("Failed to get user guilds")
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
				logrus.WithFields(logrus.Fields{
				"module":   "status",
				"action":   "get_members_error", 
				"user_id":  userID,
				"guild_id": guildID,
			}).WithError(err).Error("Failed to get guild members")
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
					logrus.WithFields(logrus.Fields{
					"module":  "status",
					"action":  "marshal_error",
					"user_id": userID,
				}).WithError(err).Error("Failed to marshal status data")
					continue
				}

				SendToUser(userID, websocket.TextMessage, statusJSON)
				guildSent++
				totalSent++
			}
			logrus.WithFields(logrus.Fields{
				"module":       "status",
				"action":       "guild_statuses_sent",
				"user_id":      userID,
				"guild_id":     guildID,
				"statuses_sent": guildSent,
			}).Debug("Guild statuses sent")
		}
	}

	logrus.WithFields(logrus.Fields{
		"module":      "status",
		"action":      "initial_statuses_complete",
		"user_id":     userID,
		"total_sent":  totalSent,
	}).Info("Initial statuses sending completed")
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


