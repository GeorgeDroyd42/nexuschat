package api

import (
	"auth.com/v4/utils"
	"auth.com/v4/internal/perms"
	"github.com/labstack/echo/v4"
)

func GetContextMenuHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	contextType := c.Param("type")
	guildID := c.QueryParam("guild_id")
	messageID := c.QueryParam("message_id")

	buttons := []map[string]interface{}{}

	switch contextType {
	case "guild":
		buttons = append(buttons,
			map[string]interface{}{"text": "Invite", "action": "invite", "color": "#4CAF50"},
			map[string]interface{}{"text": "Copy Guild ID", "action": "copy_guild_id", "color": "#4CAF50"},
			map[string]interface{}{"type": "separator"},
		)

		if guildID != "" {
			if hasPermission, err := perms.Service.HasGuildPermission(userID, guildID, perms.MANAGE_GUILD); err == nil && hasPermission {
				buttons = append(buttons, map[string]interface{}{"text": "Guild Settings", "action": "guild_settings", "color": "#dcddde"})
			}
		}

		buttons = append(buttons, map[string]interface{}{"text": "Leave Guild", "action": "leave_guild", "color": "#f04747"})

	case "channel":
		if guildID != "" {
			if hasPermission, err := perms.Service.HasGuildPermission(userID, guildID, perms.EDIT_CHANNEL); err == nil && hasPermission {
				buttons = append(buttons,
					map[string]interface{}{"text": "Channel Settings", "action": "channel_settings", "color": "#dcddde"},
					map[string]interface{}{"text": "Delete Channel", "action": "delete_channel", "color": "#f04747"},
					map[string]interface{}{"type": "separator"},
				)
			}
		}

		buttons = append(buttons, map[string]interface{}{"text": "Copy Channel ID", "action": "copy_channel_id", "color": "#4CAF50"})

	case "message":
		if messageID != "" {
			var messageUserID string
			found, _ := utils.QueryRow("GetMessageOwner", &messageUserID, "SELECT user_id FROM messages WHERE message_id = $1", messageID)

			canDelete := false
			if found && messageUserID == userID {
				canDelete = true
			} else if guildID != "" {
				if hasPermission, err := perms.Service.HasGuildPermission(userID, guildID, perms.DELETE_MESSAGE); err == nil && hasPermission {
					canDelete = true
				}
			}

			if canDelete {
				buttons = append(buttons,
					map[string]interface{}{"text": "Delete Message", "action": "delete_message", "color": "#f04747"},
					map[string]interface{}{"type": "separator"},
				)
			}
		}

		buttons = append(buttons,
			map[string]interface{}{"text": "Copy Message Content", "action": "copy_message_content", "color": "#4CAF50"},
			map[string]interface{}{"text": "Copy Message ID", "action": "copy_message_id", "color": "#4CAF50"},
		)
	}

	return c.JSON(200, map[string]interface{}{"buttons": buttons})
}
