package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"auth.com/v4/cache"
	"auth.com/v4/internal/websockets"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)



func UpgradeAndRegister(c echo.Context, userID string) (*websockets.WebSocketConnection, string, error) {
	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, "", err
	}

	ws.SetReadLimit(int64(AppConfig.MaxWSMessageSize))
	ws.SetReadDeadline(time.Now().Add(5 * time.Minute))
	
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(5 * time.Minute))
		return nil
	})

	var httpSessionToken string
	if cookie, err := c.Cookie("session"); err == nil {
		httpSessionToken = cookie.Value
	}

	sessionID := websockets.GenerateWebSocketSessionID(userID)
	wsConn := websockets.RegisterConnection(ws, userID, sessionID, httpSessionToken)
	
	logrus.WithFields(logrus.Fields{
		"module":  "websocket",
		"action":  "user_connected",
		"user_id": userID,
	}).Info("User connected via WebSocket")
	BroadcastUserStatusChange(userID, true)
	go func() {
		SendInitialStatusesToUser(userID)
	}()
	
	return wsConn, sessionID, nil
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		host := r.Host

		if origin == "http://"+host || origin == "https://"+host {
			return true
		}

		for _, allowedOrigin := range AppConfig.AllowedOrigins {
			if origin == allowedOrigin {
				return true
			}
		}
		return false
	},
}



func RemoveConnection(sessionID string) {
	userID, channelIDs := websockets.RemoveConnection(sessionID)
	
	if userID != "" {
		if len(channelIDs) > 0 {
			for _, channelID := range channelIDs {
				BroadcastTypingStatus(channelID)
			}
		}
		BroadcastUserStatusChange(userID, false)
	}
}


func HandleMessageEvent(userID string, data map[string]interface{}) error {
	channelID, ok := data["channel_id"].(string)
	if !ok {
		return fmt.Errorf("invalid channel_id")
	}

	content, ok := data["content"].(string)
	if !ok {
		return fmt.Errorf("invalid content")
	}

	var username string
	var profilePicture string
	var isWebhook bool

	if len(userID) > 3 && userID[:3] == "wh_" {
		webhookID := userID[3:]
		err := GetDB().QueryRow("SELECT name, COALESCE(profile_picture, '') FROM webhooks WHERE webhook_id = $1", webhookID).Scan(&username, &profilePicture)
		if err != nil {
			username = "Unknown Webhook"
			profilePicture = ""
		}
		isWebhook = true
	} else {
		username, _ = GetUsernameByID(userID)
		profilePicture = ""
		isWebhook = false
	}

	messageID, err := CreateMessage(channelID, userID, content)
	if err != nil {
		return err
	}

	currentTime := time.Now().UTC()
	messageData := map[string]interface{}{
		"type":            "new_message",
		"message_id":      messageID,
		"channel_id":      channelID,
		"user_id":         userID,
		"username":        username,
		"content":         content,
		"created_at":      currentTime.Format(time.RFC3339),
		"profile_picture": profilePicture,
		"is_webhook":      isWebhook,
	}

	return BroadcastToChannel(channelID, messageData)
}

func BroadcastToChannel(channelID string, data map[string]interface{}) error {
	var guildID string
	found, err := QueryRow("GetGuildFromChannel", &guildID,
		"SELECT guild_id FROM channels WHERE channel_id = $1", channelID)

	if !found || err != nil {
		return err
	}

	return BroadcastToGuildMembers(guildID, data)
}

func BroadcastToGuildMembers(guildID string, data map[string]interface{}) error {
	members, _, err := GetGuildMembersPaginated(guildID, 1, AppConfig.AllMembers)
	if err != nil {
		return err
	}

	broadcastData, _ := json.Marshal(data)

	for _, member := range members {
		isStillInGuild, err := IsUserInGuild(guildID, member.UserID)
		if err == nil && isStillInGuild {
			websockets.SendToUser(member.UserID, websocket.TextMessage, broadcastData)
		}
	}

	return nil
}

type MessageEvent struct {
	Type      string `json:"type"`
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

type MessageResult struct {
	ID        string                 `json:"message_id"`
	ChannelID string                 `json:"channel_id"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Content   string                 `json:"content"`
	CreatedAt time.Time              `json:"created_at"`
	Data      map[string]interface{} `json:"broadcast_data"`
}


func HandleWebSocketMessage(userID string, rawMessage []byte) error {
	var jsonMsg map[string]interface{}
	if json.Unmarshal(rawMessage, &jsonMsg) != nil {
		return fmt.Errorf("invalid_message_format")
	}

	if jsonMsg["type"] == "message" {
		channelID, ok1 := jsonMsg["channel_id"].(string)
		content, ok2 := jsonMsg["content"].(string)

		if !ok1 || !ok2 || channelID == "" || content == "" {
			return fmt.Errorf("invalid_message_data")
		}

		cache.Provider.RemoveTypingUser(channelID, userID)

		username, _ := GetUsernameByID(userID)

		var profilePicture sql.NullString
		QueryRow("GetUserProfilePicture", &profilePicture,
			"SELECT profile_picture FROM users WHERE user_id = $1", userID)

		profilePictureValue := ""
		if profilePicture.Valid {
			profilePictureValue = profilePicture.String
		}

		messageID, err := CreateMessage(channelID, userID, content)
		if err != nil {
			return err
		}

		currentTime := time.Now().UTC()
		result := &MessageResult{
			ID:        messageID,
			ChannelID: channelID,
			UserID:    userID,
			Username:  username,
			Content:   content,
			CreatedAt: currentTime,
			Data: map[string]interface{}{
				"type":            "new_message",
				"message_id":      messageID,
				"channel_id":      channelID,
				"user_id":         userID,
				"username":        username,
				"content":         content,
				"created_at":      currentTime.Format(time.RFC3339),
				"profile_picture": profilePictureValue,
			},
		}

		broadcastJSON, _ := json.Marshal(result.Data)
		websockets.BroadcastWithRedis(1, broadcastJSON)

		BroadcastTypingStatus(channelID)

		return nil
	}

	if jsonMsg["type"] == "status_update" {
		// Frontend explicitly requesting status update broadcast
		logrus.WithFields(logrus.Fields{
			"component": "WebSocket",
			"action":    "status_update_request",
			"user_id":   userID,
		}).Info("Frontend requested status update")
		
		BroadcastUserStatusChange(userID, true)
		return nil
	}

	if jsonMsg["type"] == "typing_start" {
		return HandleTypingStart(userID, jsonMsg)
	}

	if jsonMsg["type"] == "typing_stop" {
		return HandleTypingStop(userID, jsonMsg)
	}

	if jsonMsg["type"] == "delete_message" {
		messageID, ok := jsonMsg["message_id"].(string)
		if !ok || messageID == "" {
			return fmt.Errorf("invalid_message_id")
		}

		var channelID string
		found, _ := QueryRow("GetMessageChannel", &channelID,
			"SELECT channel_id FROM messages WHERE message_id = $1", messageID)

		if !found {
			return fmt.Errorf("message not found")
		}

		err := DeleteMessage(messageID, userID)
		if err != nil {
			return err
		}

		deleteData := map[string]interface{}{
			"type":       "message_deleted",
			"message_id": messageID,
			"channel_id": channelID,
		}
		return BroadcastToChannel(channelID, deleteData)
	}

	if jsonMsg["type"] == "request_typing_state" {
		return HandleRequestTypingState(userID, jsonMsg)
	}

	return fmt.Errorf("unsupported_message_type")

}

