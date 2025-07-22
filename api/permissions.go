package api

import (
	"auth.com/v4/internal/perms"
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

	permissions, isOwner, err := perms.Service.GetAllGuildPermissions(userID, guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	frontendPerms := map[string]bool{
		"canManageGuild":    permissions[perms.MANAGE_GUILD],
		"canManageRoles":    permissions[perms.MANAGE_ROLES],
		"canCreateChannels": permissions[perms.CREATE_CHANNEL],
		"canEditChannels":   permissions[perms.EDIT_CHANNEL],
		"canDeleteChannels": permissions[perms.DELETE_CHANNEL],
		"canDeleteMessages": permissions[perms.DELETE_MESSAGE],
		"canKickMembers":    permissions[perms.KICK_MEMBERS],
		"canCreateInvite":   permissions[perms.CREATE_INVITE],
		"canViewWebhooks":   permissions[perms.VIEW_WEBHOOKS],
		"canCreateWebhooks": permissions[perms.CREATE_WEBHOOKS],
		"canDeleteWebhooks": permissions[perms.DELETE_WEBHOOKS],
	}

	return c.JSON(200, map[string]interface{}{
		"success":     true,
		"is_owner":    isOwner,
		"permissions": frontendPerms,
	})
}