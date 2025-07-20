package api

import (
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

func GetGuildPermissionsHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "guildId")
	if err != nil {
		return err
	}

	if err := utils.RequireGuildMembership(c, userID, guildID); err != nil {
		return err
	}

	permissions, isOwner, err := utils.GetAllGuildPermissions(userID, guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	frontendPerms := map[string]bool{
		"canManageGuild":    permissions[utils.MANAGE_GUILD],
		"canManageRoles":    permissions[utils.MANAGE_ROLES],
		"canCreateChannels": permissions[utils.CREATE_CHANNEL],
		"canEditChannels":   permissions[utils.EDIT_CHANNEL],
		"canDeleteChannels": permissions[utils.DELETE_CHANNEL],
		"canDeleteMessages": permissions[utils.DELETE_MESSAGE],
		"canKickMembers":    permissions[utils.KICK_MEMBERS],
		"canCreateInvite":   permissions[utils.CREATE_INVITE],
		"canViewWebhooks":   permissions[utils.VIEW_WEBHOOKS],
		"canCreateWebhooks": permissions[utils.CREATE_WEBHOOKS],
		"canDeleteWebhooks": permissions[utils.DELETE_WEBHOOKS],
	}

	return c.JSON(200, map[string]interface{}{
		"success":     true,
		"is_owner":    isOwner,
		"permissions": frontendPerms,
	})
}
