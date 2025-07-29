package utils

import (
    "errors"
)



func HandleTypingStart(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}

	WebSocketManager.Mu.Lock()
	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			conn.IsTyping = true
			conn.TypingChannel = channelID
		}
	}
	WebSocketManager.Mu.Unlock()

	return broadcastTypingUpdate(channelID)
}

func HandleTypingStop(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}

	WebSocketManager.Mu.Lock()
	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID && conn.TypingChannel == channelID {
			conn.IsTyping = false
			conn.TypingChannel = ""
		}
	}
	WebSocketManager.Mu.Unlock()

	return broadcastTypingUpdate(channelID)
}

func broadcastTypingUpdate(channelID string) error {
	WebSocketManager.Mu.RLock()
	typingUserIDs := make(map[string]bool)
	for _, conn := range WebSocketManager.Connections {
		if conn.IsTyping && conn.TypingChannel == channelID {
			typingUserIDs[conn.UserID] = true
		}
	}
	WebSocketManager.Mu.RUnlock()

	usernames := make([]string, 0)
	for userID := range typingUserIDs {
		if username, err := GetUsernameByID(userID); err == nil {
			usernames = append(usernames, username)
		}
	}

	typingData := map[string]interface{}{
		"type":         "typing_update", 
		"channel_id":   channelID,
		"typing_users": usernames,
	}

	return BroadcastToChannel(channelID, typingData)
}



func HandleRequestTypingState(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}
	
	return broadcastTypingUpdate(channelID)
}