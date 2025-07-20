package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateWebhook(channelID, name, createdBy string) (string, string, error) {
	webhookID := GenerateSessionID("webhook")
	token, err := generateSessionToken()
	if err != nil {
		return "", "", err
	}

	err = ExecuteQuery("CreateWebhook",
		`INSERT INTO webhooks (webhook_id, channel_id, name, token, created_by) 
		 VALUES ($1, $2, $3, $4, $5)`,
		webhookID, channelID, name, token, createdBy)

	if err != nil {
		return "", "", err
	}

	return webhookID, token, nil
}

func GetChannelWebhooks(channelID string) ([]map[string]interface{}, error) {
	rows, err := GetDB().Query(`
		SELECT webhook_id, name, token, created_by, created_at, is_active, use_count, last_used, COALESCE(profile_picture, '') as profile_picture
		FROM webhooks 
		WHERE channel_id = $1 AND is_active = true
		ORDER BY created_at DESC`,
		channelID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []map[string]interface{}
	for rows.Next() {
		var webhookID, name, token, createdBy, profilePicture string
		var createdAt time.Time
		var isActive bool
		var useCount int
		var lastUsed sql.NullTime

		err := rows.Scan(&webhookID, &name, &token, &createdBy, &createdAt, &isActive, &useCount, &lastUsed, &profilePicture)
		if err != nil {
			continue
		}

		webhook := map[string]interface{}{
			"webhook_id":      webhookID,
			"name":            name,
			"token":           token,
			"created_by":      createdBy,
			"created_at":      createdAt,
			"is_active":       isActive,
			"use_count":       useCount,
			"profile_picture": profilePicture,
		}

		if lastUsed.Valid {
			webhook["last_used"] = lastUsed.Time
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

func ValidateWebhookToken(webhookID, token string) (string, bool, error) {
	var channelID string
	var isActive bool

	err := GetDB().QueryRow(`SELECT channel_id, is_active FROM webhooks 
		WHERE webhook_id = $1 AND token = $2`, webhookID, token).Scan(&channelID, &isActive)

	if err != nil {
		return "", false, err
	}

	if !isActive {
		return "", false, nil
	}

	return channelID, true, nil
}

func SaveWebhookProfilePicture(webhookName string, imageData []byte, outputFormat string) (string, error) {
	os.MkdirAll("public/webhooks", 0750)

	cleanName := strings.ReplaceAll(webhookName, "/", "_")
	cleanName = strings.ReplaceAll(cleanName, "\\", "_")
	cleanName = strings.ReplaceAll(cleanName, "..", "_")

	filename := fmt.Sprintf("%s_%d.%s",
		cleanName,
		time.Now().Unix(),
		outputFormat)

	dst := filepath.Join("public/webhooks", filename)

	if !strings.HasPrefix(filepath.Clean(dst), "public/webhooks/") {
		return "", fmt.Errorf("invalid file path")
	}

	err := os.WriteFile(dst, imageData, 0644)
	if err != nil {
		return "", err
	}

	return "/public/webhooks/" + filename, nil
}

func CreateWebhookWithProfilePicture(channelID, name, createdBy, profilePicture string) (string, string, error) {
	webhookID := GenerateSessionID("webhook")
	token, err := generateSessionToken()
	if err != nil {
		return "", "", err
	}

	var query string
	var args []interface{}

	if profilePicture != "" {
		query = `INSERT INTO webhooks (webhook_id, channel_id, name, token, created_by, profile_picture) 
				 VALUES ($1, $2, $3, $4, $5, $6)`
		args = []interface{}{webhookID, channelID, name, token, createdBy, profilePicture}
	} else {
		query = `INSERT INTO webhooks (webhook_id, channel_id, name, token, created_by) 
				 VALUES ($1, $2, $3, $4, $5)`
		args = []interface{}{webhookID, channelID, name, token, createdBy}
	}

	err = ExecuteQuery("CreateWebhook", query, args...)
	if err != nil {
		return "", "", err
	}

	return webhookID, token, nil
}
