package utils

import (
	"sync"
	"time"
)

type SessionManager struct {
	sessionMutex      sync.RWMutex
	lastExtensionTime map[string]time.Time
}

var GlobalSessionManager = &SessionManager{
	lastExtensionTime: make(map[string]time.Time),
}