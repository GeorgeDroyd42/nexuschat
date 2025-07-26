package utils

import (
	"encoding/json"
	"fmt"
	"time"
	"auth.com/v4/internal/websockets"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
)

func createTypingData(channelID string) (map[string]interface{}, error) {
	typingUserIDs, err := cache.Provider.GetTypingUsers(channelID)
	if err != nil {
		return nil, err
	}

	typingUsernames := make([]string, 0, len(typingUserIDs))
	for _, userID := range typingUserIDs {
		username, err := GetUsernameByID(userID)
		if err == nil {
			typingUsernames = append(typingUsernames, username)
		}
	}

	return map[string]interface{}{
		"type":         "typing_update",
		"channel_id":   channelID,
		"typing_users": typingUsernames,
	}, nil
}

func HandleTypingStart(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return fmt.Errorf("invalid_channel_id")
	}

	err := cache.Provider.AddTypingUser(channelID, userID, 15*time.Second)
	if err != nil {
		return err
	}

	return BroadcastTypingStatus(channelID)
}

func HandleTypingStop(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return fmt.Errorf("invalid_channel_id")
	}

	err := cache.Provider.RemoveTypingUser(channelID, userID)
	if err != nil {
		return err
	}

	return BroadcastTypingStatus(channelID)
}

func BroadcastTypingStatus(channelID string) error {
	typingData, err := createTypingData(channelID)
	if err != nil {
		return err
	}

	var guildID string
	found, err := QueryRow("GetGuildFromChannel", &guildID,
		"SELECT guild_id FROM channels WHERE channel_id = $1", channelID)

	if !found || err != nil {
		return err
	}

	return BroadcastToGuildMembers(guildID, typingData)
}

func HandleRequestTypingState(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok || channelID == "" {
		return fmt.Errorf("invalid_channel_id")
	}

	typingData, err := createTypingData(channelID)
	if err != nil {
		return err
	}

	jsonData, _ := json.Marshal(typingData)
	websockets.SendToUser(userID, websocket.TextMessage, jsonData)

	return nil
}

func BroadcastTypingStatusForChannels(channelIDs []string) {
	for _, channelID := range channelIDs {
		BroadcastTypingStatus(channelID)
	}
}