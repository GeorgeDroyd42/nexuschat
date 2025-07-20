package utils

import (
	"fmt"
	"time"
)

func CreateMessage(channelID, userID, content string) (string, error) {
	valid, errCode := ValidateMessageContent(content)
	if !valid {
		return "", fmt.Errorf("validation failed: %d", errCode)
	}

	// Skip channel access validation for webhook messages
	if !isWebhookUser(userID) {
		valid, errCode = ValidateChannelAccess(userID, channelID)
		if !valid {
			return "", fmt.Errorf("access denied: %d", errCode)
		}
	}

	messageID := GenerateSessionID("message")

	query := `INSERT INTO messages (message_id, channel_id, user_id, content, created_at) 
          VALUES ($1, $2, $3, $4, $5)`

	currentTime := time.Now().UTC()
	err := ExecuteQuery("CreateMessage", query, messageID, channelID, userID, content, currentTime)
	if err != nil {
		return "", err
	}

	return messageID, nil
}
func isWebhookUser(userID string) bool {
	return len(userID) > 3 && userID[:3] == "wh_"
}

func GetChannelMessages(channelID string, limit int, beforeMessageID string) ([]map[string]interface{}, error) {
	baseQuery := `SELECT m.message_id, m.user_id, 
      CASE 
          WHEN m.user_id LIKE 'wh_%' THEN COALESCE(w.name, 'Unknown Webhook')
          ELSE COALESCE(u.username, 'Unknown User')
      END as username,
      m.content, m.created_at, 
      CASE 
          WHEN m.user_id LIKE 'wh_%' THEN COALESCE(w.profile_picture, '')
          ELSE COALESCE(u.profile_picture, '')
      END as profile_picture,
      CASE 
          WHEN m.user_id LIKE 'wh_%' THEN true
          ELSE false
      END as is_webhook
  FROM messages m 
  LEFT JOIN users u ON m.user_id = u.user_id AND m.user_id NOT LIKE 'wh_%'
  LEFT JOIN webhooks w ON SUBSTRING(m.user_id FROM 4) = w.webhook_id AND m.user_id LIKE 'wh_%'
  WHERE m.channel_id = $1`

	var query string
	var args []interface{}

	if beforeMessageID != "" {
		query = baseQuery + ` AND m.created_at < (
            SELECT created_at FROM messages WHERE message_id = $2
        ) ORDER BY m.created_at DESC LIMIT $3`
		args = []interface{}{channelID, beforeMessageID, limit}
	} else {
		query = baseQuery + ` ORDER BY m.created_at DESC LIMIT $2`
		args = []interface{}{channelID, limit}
	}

	rows, err := GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var messageID, userID, username, content, profilePicture string
		var createdAt time.Time
		var isWebhook bool

		err := rows.Scan(&messageID, &userID, &username, &content, &createdAt, &profilePicture, &isWebhook)
		if err != nil {
			continue
		}

		message := map[string]interface{}{
			"message_id":      messageID,
			"user_id":         userID,
			"username":        username,
			"content":         content,
			"created_at":      createdAt,
			"profile_picture": profilePicture,
			"is_webhook":      isWebhook,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func DeleteMessage(messageID, userID string) error {
	// First check if message exists and get its owner
	var messageUserID string
	var channelID string
	found, _ := QueryRow("GetMessageOwner", &messageUserID,
		"SELECT user_id FROM messages WHERE message_id = $1", messageID)

	if found {
		QueryRow("GetMessageChannel", &channelID,
			"SELECT channel_id FROM messages WHERE message_id = $1", messageID)
	}

	if !found {
		return fmt.Errorf("message not found")
	}

	// Store channel_id for later use
	QueryRow("GetMessageChannel", &channelID,
		"SELECT channel_id FROM messages WHERE message_id = $1", messageID)

	// Check if user can delete (either message owner or guild owner)
	canDelete := false
	if messageUserID == userID {
		canDelete = true
	} else {
		// Check if user has DELETE_MESSAGE permission
		hasPermission, err := HasChannelPermission(userID, channelID, DELETE_MESSAGE)
		if err == nil && hasPermission {
			canDelete = true
		}
	}

	if !canDelete {
		return fmt.Errorf("permission denied")
	}

	// Delete the message
	query := "DELETE FROM messages WHERE message_id = $1"
	err := ExecuteQuery("DeleteMessage", query, messageID)
	if err != nil {
		return err
	}

	return nil
}
