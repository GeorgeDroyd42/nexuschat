package websockets

import (
	"encoding/json"
	"time"
	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)


func SendToUser(userID string, messageType int, data []byte) {
	Manager.Mu.RLock()
	defer Manager.Mu.RUnlock()

	for _, conn := range Manager.Connections {
		if conn.UserID == userID {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "WebSocket",
					"action":    "send_to_user",
					"user_id":   userID,
				}).Error("Failed to send message to user: ", err)
			}
		}
	}
}

func BroadcastToAll(messageType int, data []byte) {
	Manager.Mu.RLock()
	var failedSessions []string

	for sessionID, conn := range Manager.Connections {
		if conn != nil && conn.Conn != nil {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				failedSessions = append(failedSessions, sessionID)
			}
		}
	}
	Manager.Mu.RUnlock()

	if len(failedSessions) > 0 {
		Manager.Mu.Lock()
		for _, sessionID := range failedSessions {
			delete(Manager.Connections, sessionID)
		}
		Manager.Mu.Unlock()
	}
}

func SendEventToSpecificSession(userID, sessionToken, eventType, message string) {
	eventData := map[string]string{"type": eventType}
	if message != "" {
		eventData["message"] = message
	}
	jsonData, _ := json.Marshal(eventData)

	Manager.Mu.RLock()
	defer Manager.Mu.RUnlock()

	for _, conn := range Manager.Connections {
		if conn.UserID == userID {
			connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(conn.SessionID)
			if found {
				if token, exists := connectionData["http_session_token"]; exists && token == sessionToken {
					conn.WriteMu.Lock()
					err := conn.Conn.WriteMessage(websocket.TextMessage, jsonData)
					conn.WriteMu.Unlock()
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"component": "WebSocket",
							"action":    "send_to_specific_session",
						}).Error("Failed to send to specific session: ", err)
					}
				}
			}
		}
	}
}

func SendEventToUser(userID, eventType, message string) {
	eventData := map[string]string{"type": eventType}
	if message != "" {
		eventData["message"] = message
	}
	jsonData, _ := json.Marshal(eventData)
	SendToUser(userID, websocket.TextMessage, jsonData)
}

func StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			Manager.Mu.RLock()
			for _, conn := range Manager.Connections {
				conn.WriteMu.Lock()
				err := conn.Conn.WriteMessage(websocket.PingMessage, []byte{})
				conn.WriteMu.Unlock()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"component": "WebSocket",
						"action":    "heartbeat",
					}).Debug("Ping failed, connection may be stale: ", err)
					go func(sessionID string) {
						RemoveConnection(sessionID)
					}(conn.SessionID)
				}
			}
			Manager.Mu.RUnlock()
		}
	}()
}

// IsUserOnline checks if user has any active WebSocket connections
func IsUserOnline(userID string) bool {
	Manager.Mu.RLock()
	defer Manager.Mu.RUnlock()
	
	for _, conn := range Manager.Connections {
		if conn.UserID == userID {
			return true
		}
	}
	return false
}

// GetOnlineUsersInGuild returns list of online userIDs for a specific guild
func GetOnlineUsersInGuild(guildID string, allGuildMembers []string) []string {
	Manager.Mu.RLock()
	defer Manager.Mu.RUnlock()
	
	onlineUsers := make([]string, 0)
	onlineUserMap := make(map[string]bool)
	
	// Build map of online users
	for _, conn := range Manager.Connections {
		onlineUserMap[conn.UserID] = true
	}
	
	// Filter guild members who are online
	for _, memberID := range allGuildMembers {
		if onlineUserMap[memberID] {
			onlineUsers = append(onlineUsers, memberID)
		}
	}
	
	return onlineUsers
}