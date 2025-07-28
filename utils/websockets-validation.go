package utils

import (
	"errors"
	"auth.com/v4/cache"
)

func ValidateWebSocketSession(userID, sessionID string) (bool, error) {
	connectionData, found, err := cache.Provider.GetWebSocketConnectionData(sessionID)
	if err != nil || !found {
		Log.Error("websocket", "validate_session", "WebSocket connection data not found", nil, map[string]interface{}{"user_id": userID, "session_id": sessionID})
return false, errors.New("websocket connection data not found")
	}

	storedToken, exists := connectionData["http_session_token"]
	if !exists || storedToken == "" {
		Log.Error("websocket", "validate_session", "No session token stored for websocket", nil, map[string]interface{}{"session_id": sessionID})
return false, errors.New("no session token stored for websocket")
	}

	validatedUserID, isValid, err := ValidateSessionToken(storedToken)
	if err != nil || !isValid {
		return false, err
	}

	if validatedUserID != userID {
		Log.Error("websocket", "validate_session", "User ID mismatch in websocket session", nil, map[string]interface{}{"expected_user": userID, "actual_user": validatedUserID})
return false, errors.New("user ID mismatch")
	}

	return true, nil
}