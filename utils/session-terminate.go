package utils

import (
	"encoding/json"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

func TerminateSessionWithNotification(sessionID string, sendNotification bool) (string, bool, error) {
	userID, found, _ := GetUserBySessionID(sessionID)

	var token string
	QueryRow("GetTokenBySessionID", &token, "SELECT token FROM sessions WHERE session_id = $1", sessionID)

	cache.Provider.DeleteSession(sessionID)
	cache.Provider.DeleteSessionToken(token)

	ExecuteQuery("DeleteSession", "DELETE FROM sessions WHERE session_id = $1", sessionID)

	if found && sendNotification {
		SendEventToSpecificSession(userID, token, "session_terminated", "Your session was terminated. Please log in again.")
	}
	return userID, true, nil
}

func TerminateAllUserSessions(userID string) (bool, error) {
	sessions, err := GetSessionsByUserID(userID)
	if err != nil {
		return false, err
	}

	for _, sessionID := range sessions {
		var token string
		QueryRow("GetTokenBySessionID", &token, "SELECT token FROM sessions WHERE session_id = $1", sessionID)

		cache.Provider.DeleteSession(sessionID)
		cache.Provider.DeleteSessionToken(token)
	}

	eventData := map[string]interface{}{"type": "all_sessions_terminated"}
	broadcastData, _ := json.Marshal(eventData)
	SendToUser(userID, websocket.TextMessage, broadcastData)
	return true, nil
}