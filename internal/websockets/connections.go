package websockets

import (
	"fmt"
	"time"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

func RegisterConnection(conn *websocket.Conn, userID, sessionID, httpSessionToken string) *WebSocketConnection {
	wsConn := &WebSocketConnection{
		Conn:      conn,
		UserID:    userID,
		SessionID: sessionID,
	}

	Manager.Mu.Lock()
	Manager.Connections[sessionID] = wsConn
	Manager.Mu.Unlock()

	cache.Provider.AddWebSocketConnection(userID, sessionID, map[string]string{
		"user_id":            userID,
		"session_id":         sessionID,
		"connected":          fmt.Sprintf("%d", time.Now().Unix()),
		"http_session_token": httpSessionToken,
	}, 24*time.Hour)

	return wsConn
}

func RemoveConnection(sessionID string) (string, []string) {
	Manager.Mu.RLock()
	conn := Manager.Connections[sessionID]
	Manager.Mu.RUnlock()

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

	Manager.Mu.Lock()
	delete(Manager.Connections, sessionID)
	Manager.Mu.Unlock()

	if userID != "" {
		cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		channelIDs, _ := cache.Provider.RemoveUserFromAllTypingIndicators(userID)
		return userID, channelIDs
	}
	
	return "", nil
}