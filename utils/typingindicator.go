package utils

import (
    "errors"
	"encoding/json"
	"github.com/gorilla/websocket"
)



func HandleTypingStart(userID string, sessionID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}

	WebSocketManager.Mu.Lock()
	if conn, exists := WebSocketManager.Connections[sessionID]; exists {
		conn.IsTyping = true
		conn.TypingChannel = channelID
	}
	WebSocketManager.Mu.Unlock()

	return broadcastTypingUpdate(channelID)
}

func HandleTypingStop(userID string, sessionID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}

	WebSocketManager.Mu.Lock()
	if conn, exists := WebSocketManager.Connections[sessionID]; exists && conn.TypingChannel == channelID {
		conn.IsTyping = false
		conn.TypingChannel = ""
	}
	WebSocketManager.Mu.Unlock()

	return broadcastTypingUpdate(channelID)
}

func broadcastTypingUpdate(channelID string) error {
	WebSocketManager.Mu.RLock()
	typingUserIDs := make(map[string]bool)
	for _, conn := range WebSocketManager.Connections {
		if conn.IsTyping && conn.TypingChannel == channelID && conn.Conn != nil {
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

	var guildID string
found, err := QueryRow("GetGuildFromChannel", &guildID,
    "SELECT guild_id FROM channels WHERE channel_id = $1", channelID)
if !found || err != nil {
    return err
}

members, _, err := GetGuildMembersPaginated(guildID, 1, AppConfig.AllMembers)
if err != nil {
    return err
}

broadcastData, _ := json.Marshal(typingData)
for _, member := range members {
    if !typingUserIDs[member.UserID] {
        SendToUser(member.UserID, websocket.TextMessage, broadcastData)
    }
}
return nil
}



func HandleRequestTypingState(userID string, sessionID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return errors.New("invalid_channel_id")
	}
	
	return broadcastTypingUpdate(channelID)
}