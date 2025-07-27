package utils

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

var WebSocketManager = &ConnectionManager{
	Connections: make(map[string]*WebSocketConnection),
}