package websockets

import (
	"sync"
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