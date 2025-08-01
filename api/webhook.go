package api

import (
	"encoding/json"
	"auth.com/v4/utils"
	"auth.com/v4/internal/perms"
	"auth.com/v4/internal/webhook"
	"github.com/labstack/echo/v4"
)

func CreateWebhookHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	channelID, err := utils.RequireParam(c, "channelid")
	if err != nil {
		return err
	}

	name := c.FormValue("name")
	if name == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}
	_, err = perms.Service.RequireChannelPermission(c, userID, channelID, perms.CREATE_WEBHOOKS)
	if err != nil {
		return err
	}
	var profilePicturePath string
	file, err := c.FormFile("profile_picture")
	if err == nil {
		imageData, outputFormat, valid, errCode := utils.ValidateProfilePicture(file)
		if !valid {
			return utils.SendErrorResponse(c, errCode)
		}

		profilePicturePath, err = webhook.Service.SaveWebhookProfilePicture(name, imageData, outputFormat)
		if err != nil {
			return utils.SendErrorResponse(c, utils.ErrDatabaseError)
		}
	}

	webhookID, token, err := webhook.Service.CreateWebhookWithProfilePicture(channelID, name, userID, profilePicturePath)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	var guildID string
	utils.QueryRow("GetGuildFromChannel", &guildID,
		"SELECT guild_id FROM channels WHERE channel_id = $1", channelID)

	webhookData := map[string]interface{}{
		"type":       "webhook_created",
		"channel_id": channelID,
		"guild_id":   guildID,
		"webhook_id": webhookID,
		"name":       name,
	}
	broadcastData, _ := json.Marshal(webhookData)
	utils.BroadcastWithRedis(1, broadcastData)

	return c.JSON(200, map[string]interface{}{
		"success":    true,
		"message":    "Webhook created successfully",
		"webhook_id": webhookID,
		"token":      token,
	})
}

func ListWebhooksHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	channelID, err := utils.RequireParam(c, "channelid")
	if err != nil {
		return err
	}
	_, err = perms.Service.RequireChannelPermission(c, userID, channelID, perms.VIEW_WEBHOOKS)
	if err != nil {
		return err
	}

	webhooks, err := webhook.Service.GetChannelWebhooks(channelID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success":  true,
		"webhooks": webhooks,
	})
}

func ExecuteWebhookHandler(c echo.Context) error {
	webhookID, err := utils.RequireParam(c, "webhookid")
	if err != nil {
		return err
	}

	token, err := utils.RequireParam(c, "token")
	if err != nil {
		return err
	}

	content := c.FormValue("content")
	if content == "" {
		return utils.SendErrorResponse(c, utils.ErrEmptyMessage)
	}

	channelID, valid, err := webhook.Service.ValidateWebhookToken(webhookID, token)
	if !valid || err != nil {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	valid, errCode := utils.ValidateMessageContent(content)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	// Use existing WebSocket message system for real-time broadcasting
	messageData := map[string]interface{}{
		"channel_id": channelID,
		"content":    content,
	}

	err = utils.HandleMessageEvent("wh_"+webhookID, messageData)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
	})
}

func DeleteWebhookHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	webhookID, err := utils.RequireParam(c, "webhookid")
	if err != nil {
		return err
	}

	channelID, guildID, err := perms.Service.RequireWebhookPermission(c, userID, webhookID, perms.DELETE_WEBHOOKS)
	if err != nil {
		return err
	}

	err = utils.ExecuteQuery("DeleteWebhook",
		"DELETE FROM webhooks WHERE webhook_id = $1", webhookID)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	webhookData := map[string]interface{}{
		"type":       "webhook_deleted",
		"channel_id": channelID,
		"guild_id":   guildID,
	}
	broadcastData, _ := json.Marshal(webhookData)
	utils.BroadcastWithRedis(1, broadcastData)

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"message": "Webhook deleted successfully",
	})
}
