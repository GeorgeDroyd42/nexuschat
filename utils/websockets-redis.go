package utils

import (
	"encoding/json"

	"auth.com/v4/cache"
)

func BroadcastWithRedis(messageType int, data []byte) {
	var messageData map[string]interface{}
	json.Unmarshal(data, &messageData)

	channelID, hasChannel := messageData["channel_id"].(string)

	msg := struct {
		Type      int    `json:"type"`
		Data      []byte `json:"data"`
		ChannelID string `json:"channel_id,omitempty"`
		Secure    bool   `json:"secure"`
	}{
		Type:      messageType,
		Data:      data,
		ChannelID: channelID,
		Secure:    hasChannel,
	}

	err := cache.Provider.PublishMessage("broadcast", msg)
	if err != nil {
		Log.Error("WebSocket", "broadcast_publish", "Failed to publish broadcast message", err)
	}
}