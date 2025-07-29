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
	var typingChannel string
	if conn != nil {
		userID = conn.UserID
		if conn.IsTyping {
			typingChannel = conn.TypingChannel
		}
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

	if typingChannel != "" {
		broadcastTypingUpdate(typingChannel)
	}

	if userID != "" {
		cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		
		// Only broadcast offline if no other connections exist
		if !IsUserOnline(userID) {
			BroadcastUserStatusChange(userID, false)
		}
		
	}
}

func CleanupUserWebSocketConnections(userID string) {
	connections, found, _ := cache.Provider.GetWebSocketConnections(userID)
	if !found {
		return
	}

	for _, sessionID := range connections {
		RemoveWebSocketConnection(sessionID)
	}
}

func GenerateWebSocketSessionID(userID string) string {
	return fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
}

func DisconnectWebSocketsByToken(userID, sessionToken string) {
	connections, found, _ := cache.Provider.GetWebSocketConnections(userID)
	if !found {
		return
	}

	var sessionsToDisconnect []string

	for _, wsSessionID := range connections {
		connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(wsSessionID)
		if found {
			if storedToken, exists := connectionData["http_session_token"]; exists && storedToken == sessionToken {
				sessionsToDisconnect = append(sessionsToDisconnect, wsSessionID)
			}
		}
	}

	WebSocketManager.Mu.Lock()
	for _, wsSessionID := range sessionsToDisconnect {
		if conn, exists := WebSocketManager.Connections[wsSessionID]; exists {
			conn.Conn.Close()
			delete(WebSocketManager.Connections, wsSessionID)
		}
	}
	WebSocketManager.Mu.Unlock()

	for _, wsSessionID := range sessionsToDisconnect {
		cache.Provider.RemoveWebSocketConnection(userID, wsSessionID)
	}

	if len(sessionsToDisconnect) > 0 {
		
		// Only broadcast offline if no other connections exist
		if !IsUserOnline(userID) {
			BroadcastUserStatusChange(userID, false)
		}
	}
}