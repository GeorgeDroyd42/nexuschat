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



func BroadcastToAll(messageType int, data []byte) {
	websockets.Manager.Mu.RLock()
	var failedSessions []string

	for sessionID, conn := range websockets.Manager.Connections {
		if conn != nil && conn.Conn != nil {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				failedSessions = append(failedSessions, sessionID)
			}
		}
	}
	websockets.Manager.Mu.RUnlock()

	if len(failedSessions) > 0 {
		websockets.Manager.Mu.Lock()
		for _, sessionID := range failedSessions {
			delete(websockets.Manager.Connections, sessionID)
		}
		websockets.Manager.Mu.Unlock()
	}
}

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

	sessionID := GenerateWebSocketSessionID(userID)
	wsConn := websockets.RegisterConnection(ws, userID, sessionID, httpSessionToken)
	
	log(logrus.InfoLevel, "WebSocket", "user_connected", userID, nil)
	HandleUserConnect(userID)
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
	websockets.Manager.Mu.RLock()
	conn := websockets.Manager.Connections[sessionID]
	websockets.Manager.Mu.RUnlock()

	var userID string
	if conn != nil {
		userID = conn.UserID
	} else {
		connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(sessionID)
		if found {
			if uid, exists := connectionData["user_id"]; exists {
				userID = uid
			}
		}
	}

	websockets.Manager.Mu.Lock()
	delete(websockets.Manager.Connections, sessionID)
	websockets.Manager.Mu.Unlock()

	if userID != "" {
		cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		channelIDs, _ := cache.Provider.RemoveUserFromAllTypingIndicators(userID)
		if len(channelIDs) > 0 {
			BroadcastTypingStatusForChannels(channelIDs)
		}
		HandleUserDisconnect(userID)
	}
}

func SendToUser(userID string, messageType int, data []byte) {

	websockets.Manager.Mu.RLock()
	defer websockets.Manager.Mu.RUnlock()

	for _, conn := range websockets.Manager.Connections {
		if conn.UserID == userID {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				log(logrus.ErrorLevel, "WebSocket", "send_to_user", "", err)
			}
		}
	}
}
func StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			websockets.Manager.Mu.RLock()
			for _, conn := range websockets.Manager.Connections {
				conn.WriteMu.Lock()
				err := conn.Conn.WriteMessage(websocket.PingMessage, []byte{})
				conn.WriteMu.Unlock()
				if err != nil {
					log(logrus.DebugLevel, "WebSocket", "heartbeat", "Ping failed, connection may be stale", err)
					go RemoveConnection(conn.SessionID)
				}
			}
			websockets.Manager.Mu.RUnlock()
		}
	}()
}

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
		log(logrus.ErrorLevel, "WebSocket", "broadcast_publish", "", err)
	}
}

func SendToUserWithRedis(userID string, messageType int, data []byte) {
	msg := struct {
		Type   int    `json:"type"`
		Data   []byte `json:"data"`
		UserID string `json:"user_id"`
	}{
		Type:   messageType,
		Data:   data,
		UserID: userID,
	}

	err := cache.Provider.PublishMessage("user_messages", msg)
	if err != nil {
		log(logrus.ErrorLevel, "WebSocket", "send_user_publish", "", err)
	}
}

func CleanupUserWebSocketConnections(userID string) {
	connections, found, _ := cache.Provider.GetWebSocketConnections(userID)

	websockets.Manager.Mu.Lock()
	for sessionID, conn := range websockets.Manager.Connections {
		if conn.UserID == userID {
			conn.Conn.Close()
			delete(websockets.Manager.Connections, sessionID)
		}
	}
	websockets.Manager.Mu.Unlock()

	if found {
		for _, sessionID := range connections {
			cache.Provider.RemoveWebSocketConnection(userID, sessionID)
		}
	}
}

func SendErrorToUser(userID string, errorCode int) {
	errorMessage := ErrorMessages[errorCode]
	eventData := map[string]interface{}{
		"type":       "error",
		"error_code": errorCode,
		"message":    errorMessage,
	}
	jsonData, _ := json.Marshal(eventData)
	SendToUser(userID, websocket.TextMessage, jsonData)
}

func SendEventToUser(userID, eventType, message string) {
	eventData := map[string]string{"type": eventType}
	if message != "" {
		eventData["message"] = message
	}
	jsonData, _ := json.Marshal(eventData)
	SendToUser(userID, websocket.TextMessage, jsonData)
}

func GenerateWebSocketSessionID(userID string) string {
	return fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
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
	members, _, err := GetGuildMembersPaginated(guildID, 1, 0)
	if err != nil {
		return err
	}

	broadcastData, _ := json.Marshal(data)

	for _, member := range members {
		isStillInGuild, err := IsUserInGuild(guildID, member.UserID)
		if err == nil && isStillInGuild {
			SendToUser(member.UserID, websocket.TextMessage, broadcastData)
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
		BroadcastWithRedis(1, broadcastJSON)

		BroadcastTypingStatus(channelID)

		return nil
	}

	if jsonMsg["type"] == "status_update" {
		statusData := map[string]interface{}{
			"type":      "user_status_changed",
			"user_id":   userID,
			"is_online": true,
		}

		userGuilds, err := GetUserGuilds(userID)
		if err == nil {
			for _, guild := range userGuilds {
				if guildID, ok := guild["guild_id"].(string); ok {
					statusData["guild_id"] = guildID
					BroadcastToGuildMembers(guildID, statusData)
				}
			}
		}
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

func SendEventToSpecificSession(userID, sessionToken, eventType, message string) {
	eventData := map[string]string{"type": eventType}
	if message != "" {
		eventData["message"] = message
	}
	jsonData, _ := json.Marshal(eventData)

	websockets.Manager.Mu.RLock()
	defer websockets.Manager.Mu.RUnlock()

	for _, conn := range websockets.Manager.Connections {
		if conn.UserID == userID {
			connectionData, found, _ := cache.Provider.GetWebSocketConnectionData(conn.SessionID)
			if found {
				if token, exists := connectionData["http_session_token"]; exists && token == sessionToken {
					conn.WriteMu.Lock()
					err := conn.Conn.WriteMessage(websocket.TextMessage, jsonData)
					conn.WriteMu.Unlock()
					if err != nil {
						log(logrus.ErrorLevel, "WebSocket", "send_to_specific_session", "", err)
					}
				}
			}
		}
	}
}
