package utils

import (
	"fmt"
	"sync"
	"time"

	"auth.com/v4/cache"
	"github.com/labstack/echo/v4"
)

type SessionManager struct {
	sessionMutex      sync.RWMutex
	lastExtensionTime map[string]time.Time
}

var GlobalSessionManager = &SessionManager{
	lastExtensionTime: make(map[string]time.Time),
}

func (sm *SessionManager) RefreshSession(c echo.Context, userID string) (time.Time, error) {
	cookie, err := c.Cookie("session")
	if err != nil {
		return time.Time{}, fmt.Errorf("no session cookie")
	}
	oldToken := cookie.Value

	// Validate current session
	if err := sm.validateSessionOwnership(oldToken, userID); err != nil {
		return time.Time{}, err
	}

	// Create new session
	newToken, err := sm.createNewSession(userID)
	if err != nil {
		return time.Time{}, err
	}

	// Set new cookie
	newExpiresAt := time.Now().Add(AppConfig.SessionExpiryDuration)
	newCookie := createSessionCookie(newToken, newExpiresAt)
	c.SetCookie(newCookie)

	// Update WebSocket connections with new token
	sm.UpdateWebSocketTokens(userID, oldToken, newToken)

	// Wait 100ms then cleanup old session
	go func() {
		time.Sleep(100 * time.Millisecond)
		sm.cleanupOldSession(oldToken)
	}()

	return newExpiresAt, nil
}

func (sm *SessionManager) validateSessionOwnership(token, userID string) error {
	cacheKey := cache.DefaultKeys.Session(token)
	var sessionUserID string
	found, err := cache.Provider.Get(cacheKey, &sessionUserID)
	if err != nil {
		return err
	}

	if !found {
		var dbUserID string
		found, err = QueryRow("GetSessionUserID", &dbUserID,
			"SELECT user_id FROM sessions WHERE token = $1", token)
		if !found || err != nil {
			return fmt.Errorf("session not found")
		}
		sessionUserID = dbUserID
	}

	if sessionUserID != userID {
		return fmt.Errorf("session user mismatch")
	}
	return nil
}

func (sm *SessionManager) createNewSession(userID string) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	sessionID := GenerateSessionID("session")
	expiresAt := time.Now().Add(AppConfig.SessionExpiryDuration)

	err = cache.Provider.SetSession(sessionID, userID, AppConfig.SessionExpiryDuration)
	if err != nil {
		return "", err
	}

	err = ExecuteQuery("CreateSession",
		"INSERT INTO sessions (token, user_id, session_id, expires_at) VALUES ($1, $2, $3, $4)",
		token, userID, sessionID, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (sm *SessionManager) cleanupOldSession(token string) {
	sessionID, found, _ := GetSessionIDByToken(token)
	if found {
		TerminateSession(sessionID)
	}
}

func (sm *SessionManager) UpdateWebSocketTokens(userID, oldToken, newToken string) {
	// Get all websocket connections for this user
	connections, found, _ := cache.Provider.GetWebSocketConnections(userID)
	if !found {
		return
	}

	// Update each connection's stored token
	for _, wsSessionID := range connections {
		connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(wsSessionID)
		if found {
			if storedToken, exists := connectionData["http_session_token"]; exists && storedToken == oldToken {
				connectionData["http_session_token"] = newToken
				cache.Provider.AddWebSocketConnection(userID, wsSessionID, connectionData, 24*time.Hour)
				fmt.Printf("🔄 Updated WebSocket token for session %s\n", wsSessionID)
			}
		}
	}
}

func (sm *SessionManager) ExtendSessionSafe(token string) error {
	// Check if session was extended recently (throttling)
	sm.sessionMutex.RLock()
	lastExtended, exists := sm.lastExtensionTime[token]
	sm.sessionMutex.RUnlock()

	if exists && time.Since(lastExtended) < 1*time.Second {
		return nil // Skip if extended within last second
	}

	// Record extension time and cleanup old entries
	sm.sessionMutex.Lock()
	sm.lastExtensionTime[token] = time.Now()

	// Cleanup old entries (prevent memory leak)
	cutoff := time.Now().Add(-5 * time.Minute)
	for t, lastTime := range sm.lastExtensionTime {
		if lastTime.Before(cutoff) {
			delete(sm.lastExtensionTime, t)
		}
	}
	sm.sessionMutex.Unlock()

	return ExtendSession(token, AppConfig.SessionExpiryDuration)
}
