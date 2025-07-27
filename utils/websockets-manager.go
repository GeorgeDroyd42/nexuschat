package utils

import (
	"encoding/json"
	"time"
	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)


func SendToUser(userID string, messageType int, data []byte) {
	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()

	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				Log.Error("WebSocket", "send_to_user", "Failed to send message to user", err, map[string]interface{}{"user_id": userID})
			}
		}
	}
}

func BroadcastToAll(messageType int, data []byte) {
	WebSocketManager.Mu.RLock()
	var failedSessions []string

	for sessionID, conn := range WebSocketManager.Connections {
		if conn != nil && conn.Conn != nil {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				failedSessions = append(failedSessions, sessionID)
			}
		}
	}
	WebSocketManager.Mu.RUnlock()

	if len(failedSessions) > 0 {
		WebSocketManager.Mu.Lock()
		for _, sessionID := range failedSessions {
			delete(WebSocketManager.Connections, sessionID)
		}
		WebSocketManager.Mu.Unlock()
	}
}

func SendEventToSpecificSession(userID, sessionToken, eventType, message string) {
	eventData := map[string]string{"type": eventType}
	if message != "" {
		eventData["message"] = message
	}
	jsonData, _ := json.Marshal(eventData)

	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()

	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(conn.SessionID)
			if found {
				if token, exists := connectionData["http_session_token"]; exists && token == sessionToken {
					conn.WriteMu.Lock()
					err := conn.Conn.WriteMessage(websocket.TextMessage, jsonData)
					conn.WriteMu.Unlock()
					if err != nil {
						Log.Error("WebSocket", "send_to_specific_session", "Failed to send to specific session", err)
					}
				}
			}
		}
	}
}

func StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			WebSocketManager.Mu.RLock()
			for _, conn := range WebSocketManager.Connections {
				conn.WriteMu.Lock()
				err := conn.Conn.WriteMessage(websocket.PingMessage, []byte{})
				conn.WriteMu.Unlock()
				if err != nil {
					Log.Debug("WebSocket", "heartbeat", "Ping failed, connection may be stale")
					go func(sessionID string) {
						RemoveWebSocketConnection(sessionID)
					}(conn.SessionID)
				}
			}
			WebSocketManager.Mu.RUnlock()
		}
	}()
}

// IsUserOnline checks if user has any active WebSocket connections
func IsUserOnline(userID string) bool {
	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()
	
	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			return true
		}
	}
	return false
}

// GetOnlineUsersInGuild returns list of online userIDs for a specific guild
func GetOnlineUsersInGuild(guildID string, allGuildMembers []string) []string {
	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()
	
	onlineUsers := make([]string, 0)
	onlineUserMap := make(map[string]bool)
	
	// Build map of online users
	for _, conn := range WebSocketManager.Connections {
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