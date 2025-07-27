package api

import (
	"auth.com/v4/internal/invite"
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

// Generate invite code for a guild
func GenerateInviteHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "guild_id")
	if err != nil {
		return err
	}

	// Check if user is in the guild
	isInGuild, err := utils.IsUserInGuild(guildID, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if !isInGuild {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	// Generate invite code
	inviteCode, err := invite.Service.CreateInviteCode(guildID, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success":     true,
		"invite_code": inviteCode,
		"invite_url":  c.Scheme() + "://" + c.Request().Host + "/i/" + inviteCode,
	})
}

// Join guild using invite code
func JoinByInviteHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	inviteCode, err := utils.RequireParam(c, "code")
	if err != nil {
		return err
	}

	// Validate invite code format

	guildID, err := invite.Service.GetGuildByInviteCode(inviteCode)
	if err != nil {
		return utils.SendErrorResponse(c, invite.ErrInvalidInviteCode)
	}
	validUser, err := utils.GetValidUserID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	guilds, err := utils.GetUserGuilds(validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if len(guilds) >= utils.AppConfig.MaxGuildsPerUser {
		return utils.SendErrorResponse(c, utils.ErrMaxGuildsReached)
	}

	isInGuild, err := utils.IsUserInGuild(guildID, validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	if isInGuild {
		return c.JSON(200, map[string]interface{}{
			"success":        true,
			"already_member": true,
			"redirect_url":   "/v/" + guildID,
			"message":        "You're already a member of this guild",
		})
	}

	err = utils.AddGuildMember(guildID, validUser)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	utils.NotifyUserGuildAdded(validUser, guildID)
	username, err := utils.GetUsernameByID(validUser)
	if err == nil {
		utils.BroadcastMemberEvent(guildID, "member_joined", validUser, username)
	}
	return c.JSON(200, map[string]interface{}{
		"success":      true,
		"message":      "Joined guild successfully",
		"redirect_url": "/v/" + guildID,
	})
}

// Get guild info from invite code (for invite page)
func GetInviteInfoHandler(c echo.Context) error {
	inviteCode, err := utils.RequireParam(c, "code")
	if err != nil {
		return err
	}
	guildID, err := invite.Service.GetGuildByInviteCode(inviteCode)
	if err != nil {
		return c.Render(200, "guild_not_found.html", nil)
	}

	guild, found, err := utils.GetGuildByID(guildID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	if !found {
		return c.Render(200, "guild_not_found.html", nil)
	}

	// Add invite code to template data
	templateData := map[string]interface{}{
		"name":         guild["name"],
		"description":  guild["description"],
		"created_at":   guild["created_at"],
		"guild_id":     guild["guild_id"],
		"invite_code":  inviteCode,
		"member_count": guild["member_count"],
	}

	return c.Render(200, "invite.html", templateData)
}

