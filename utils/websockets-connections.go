package utils

import (
	"fmt"
	"time"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

func RegisterWebSocketConnection(conn *websocket.Conn, userID, sessionID, httpSessionToken string) *WebSocketConnection {
	wsConn := &WebSocketConnection{
		Conn:      conn,
		UserID:    userID,
		SessionID: sessionID,
	}

	WebSocketManager.Mu.Lock()
	WebSocketManager.Connections[sessionID] = wsConn
	WebSocketManager.Mu.Unlock()

	cache.Provider.AddWebSocketConnection(userID, sessionID, map[string]string{
		"user_id":            userID,
		"session_id":         sessionID,
		"connected":          fmt.Sprintf("%d", time.Now().Unix()),
		"http_session_token": httpSessionToken,
	}, 24*time.Hour)

	return wsConn
}

func RemoveWebSocketConnection(sessionID string) {
	WebSocketManager.Mu.RLock()
	conn := WebSocketManager.Connections[sessionID]
	WebSocketManager.Mu.RUnlock()

	var userID string
	if conn != nil {
		userID = conn.UserID
	} else {
		connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(sessionID)
		if found {
			if uid, exists := connectionData["user_id"]; exists {
				userID = uid
			}
		}
	}

	WebSocketManager.Mu.Lock()
	delete(WebSocketManager.Connections, sessionID)
	WebSocketManager.Mu.Unlock()

	if userID != "" {
		cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		channelIDs, _ := cache.Provider.RemoveUserFromAllTypingIndicators(userID)
		
		if len(channelIDs) > 0 {
			BroadcastTypingStatusForChannels(channelIDs)
		}
		BroadcastUserStatusChange(userID, false)
	}
}

func CleanupUserWebSocketConnections(userID string) {
	connections, found, _ := cache.Provider.GetWebSocketConnections(userID)

	WebSocketManager.Mu.Lock()
	for sessionID, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			conn.Conn.Close()
			delete(WebSocketManager.Connections, sessionID)
		}
	}
	WebSocketManager.Mu.Unlock()

	if found {
		for _, sessionID := range connections {
			cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		}
	}
}

func GenerateWebSocketSessionID(userID string) string {
	return fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
}