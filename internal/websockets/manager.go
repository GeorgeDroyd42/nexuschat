package websockets

import (
	"fmt"
	"sync"
	"time"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	Conn      *websocket.Conn
	UserID    string
	SessionID string
	WriteMu   sync.Mutex
}

type ConnectionManager struct {
	Connections map[string]*WebSocketConnection
	Mu          sync.RWMutex
}

var Manager = &ConnectionManager{
	Connections: make(map[string]*WebSocketConnection),
}

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

