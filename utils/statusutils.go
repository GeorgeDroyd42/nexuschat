// X code: Status logic scattered across multiple files

// Y code: Centralized status management
package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// IsUserOnline checks if a user has any active websocket connections
func IsUserOnline(userID string) bool {
	var isOnline bool
	var connectionCount int

	err := GetDB().QueryRow("SELECT is_online, connection_count FROM users WHERE user_id = $1", userID).Scan(&isOnline, &connectionCount)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "check_online_error", userID, err)
		return false
	}

	actuallyOnline := isOnline && connectionCount > 0
	log(logrus.DebugLevel, "Status", "check_online", userID, fmt.Errorf("user %s online status: %t, connection_count: %d, actually_online: %t", userID, isOnline, connectionCount, actuallyOnline))
	return actuallyOnline
}

func HandleUserConnect(userID string) {
	log(logrus.InfoLevel, "Status", "user_connect", userID, fmt.Errorf("user %s connected", userID))

	var connectionCount int
	var wasOnline bool

	tx, err := GetDB().Begin()
	if err != nil {
		log(logrus.ErrorLevel, "Status", "begin_tx_failed", userID, err)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow("SELECT connection_count, is_online FROM users WHERE user_id = $1", userID).Scan(&connectionCount, &wasOnline)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "get_connection_count_failed", userID, err)
		return
	}

	newConnectionCount := connectionCount + 1
	newIsOnline := newConnectionCount > 0

	_, err = tx.Exec("UPDATE users SET connection_count = $1, is_online = $2, last_seen = CURRENT_TIMESTAMP WHERE user_id = $3",
		newConnectionCount, newIsOnline, userID)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "update_connection_count_failed", userID, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log(logrus.ErrorLevel, "Status", "commit_failed", userID, err)
		return
	}

	log(logrus.InfoLevel, "Status", "connect_status", userID, fmt.Errorf("user %s connection count: %d -> %d, wasOnline: %t, nowOnline: %t",
		userID, connectionCount, newConnectionCount, wasOnline, newIsOnline))

	if !wasOnline && newIsOnline {
		log(logrus.InfoLevel, "Status", "broadcasting_online", userID, fmt.Errorf("user %s going online", userID))
		BroadcastUserStatusChange(userID, true)
	}
}

func HandleUserDisconnect(userID string) {
	log(logrus.InfoLevel, "Status", "user_disconnect", userID, fmt.Errorf("user %s disconnected", userID))

	var connectionCount int
	var wasOnline bool

	tx, err := GetDB().Begin()
	if err != nil {
		log(logrus.ErrorLevel, "Status", "begin_tx_failed", userID, err)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow("SELECT connection_count, is_online FROM users WHERE user_id = $1", userID).Scan(&connectionCount, &wasOnline)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "get_connection_count_failed", userID, err)
		return
	}

	newConnectionCount := connectionCount - 1
	if newConnectionCount < 0 {
		newConnectionCount = 0
	}
	newIsOnline := newConnectionCount > 0

	_, err = tx.Exec("UPDATE users SET connection_count = $1, is_online = $2, last_seen = CURRENT_TIMESTAMP WHERE user_id = $3",
		newConnectionCount, newIsOnline, userID)
	if err != nil {
		log(logrus.ErrorLevel, "Status", "update_connection_count_failed", userID, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log(logrus.ErrorLevel, "Status", "commit_failed", userID, err)
		return
	}

	log(logrus.InfoLevel, "Status", "disconnect_status", userID, fmt.Errorf("user %s connection count: %d -> %d, wasOnline: %t, nowOnline: %t",
		userID, connectionCount, newConnectionCount, wasOnline, newIsOnline))

	if wasOnline && !newIsOnline {
		log(logrus.InfoLevel, "Status", "scheduling_offline", userID, fmt.Errorf("user %s scheduling offline in 1s", userID))
		go func(uid string) {
			time.Sleep(1 * time.Second)
			
			var currentConnectionCount int
			var currentIsOnline bool
			err := GetDB().QueryRow("SELECT connection_count, is_online FROM users WHERE user_id = $1", uid).Scan(&currentConnectionCount, &currentIsOnline)
			if err != nil {
				log(logrus.ErrorLevel, "Status", "check_delayed_status_error", uid, err)
				return
			}
			
			actuallyOnline := currentIsOnline && currentConnectionCount > 0
			if !actuallyOnline {
				log(logrus.InfoLevel, "Status", "broadcasting_offline", uid, fmt.Errorf("user %s going offline after delay (count: %d, online: %t)", uid, currentConnectionCount, currentIsOnline))
				BroadcastUserStatusChange(uid, false)
			} else {
				log(logrus.InfoLevel, "Status", "cancelled_offline", uid, fmt.Errorf("user %s reconnected, cancelled offline (count: %d, online: %t)", uid, currentConnectionCount, currentIsOnline))
			}
		}(userID)
	}
}

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
			members, _, err := GetGuildMembersWithStatus(guildID, 1, 0)
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

				SendToUser(userID, websocket.TextMessage, statusJSON)
				guildSent++
				totalSent++
			}
			log(logrus.DebugLevel, "Status", "guild_statuses_sent", userID, fmt.Errorf("sent %d statuses for guild %s", guildSent, guildID))
		}
	}

	log(logrus.InfoLevel, "Status", "initial_statuses_complete", userID, fmt.Errorf("sent %d total statuses to user %s", totalSent, userID))
}

// GetGuildMembersWithStatus returns guild members with their online status
func GetGuildMembersWithStatus(guildID string, page, limit int) ([]MemberData, int, error) {
	members, totalCount, err := GetGuildMembersPaginated(guildID, page, limit)
	if err != nil {
		return members, totalCount, err
	}

	// Add online status to each member
	for i := range members {
		members[i].IsOnline = IsUserOnline(members[i].UserID)
	}

	log(logrus.InfoLevel, "Status", "members_with_status", "", fmt.Errorf("loaded %d members for guild %s with status", len(members), guildID))

	return members, totalCount, nil
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

func CleanupUserStatusesOnServerStart() {
	log(logrus.InfoLevel, "Status", "server_startup_cleanup", "", fmt.Errorf("starting user status cleanup on server restart"))

	result, err := GetDB().Exec("UPDATE users SET connection_count = 0, is_online = false WHERE connection_count > 0 OR is_online = true")
	if err != nil {
		log(logrus.ErrorLevel, "Status", "cleanup_failed", "", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log(logrus.InfoLevel, "Status", "cleanup_complete", "", fmt.Errorf("reset %d users to offline status on server startup", rowsAffected))
}
