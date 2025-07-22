package perms

import (
	"database/sql"
	"net/http"
	"github.com/labstack/echo/v4"
)

const (
	ErrChannelNotFound         = 2012
	ErrUnauthorized           = 1003
	ErrDatabaseError          = 1007
	ErrInsufficientPermissions = 2020
	ErrUserNotInGuild         = 2010
)

var Service *PermissionService

type PermissionService struct {
	db *sql.DB
}

func InitService(database *sql.DB) {
	Service = &PermissionService{
		db: database,
	}
}

func (s *PermissionService) sendErrorResponse(c echo.Context, errCode int) error {
	statusCode := http.StatusInternalServerError
	message := "Unknown error"
	
	switch errCode {
	case ErrChannelNotFound:
		statusCode = http.StatusNotFound
		message = "Channel does not exist"
	case ErrUnauthorized:
		statusCode = http.StatusUnauthorized
		message = "You do not have permission to access this resource."
	case ErrDatabaseError:
		statusCode = http.StatusInternalServerError
		message = "Database error"
	case ErrInsufficientPermissions:
		statusCode = http.StatusForbidden
		message = "Unauthorized"
	case ErrUserNotInGuild:
		statusCode = http.StatusForbidden
		message = "You do not have permission to view info for this guild, as you are not in it"
	}
	
	return c.JSON(statusCode, echo.Map{"error": message})
}

func (s *PermissionService) isUserInGuild(guildID, userID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM guild_members WHERE guild_id = $1 AND user_id = $2)", guildID, userID).Scan(&exists)
	return exists, err
}

func (s *PermissionService) requireGuildMembership(c echo.Context, userID, guildID string) error {
	isInGuild, err := s.isUserInGuild(guildID, userID)
	if err != nil {
		return s.sendErrorResponse(c, ErrDatabaseError)
	}
	if !isInGuild {
		return s.sendErrorResponse(c, ErrUserNotInGuild)
	}
	return nil
}

func (s *PermissionService) HasGuildPermission(userID, guildID, permission string) (bool, error) {
	if guildID == "" || userID == "" {
		return false, nil
	}

	var ownerID string
	err := s.db.QueryRow("SELECT owner_id FROM guilds WHERE guild_id = $1", guildID).Scan(&ownerID)
	if err != nil {
		return false, err
	}
	
	return ownerID == userID, nil
}

func (s *PermissionService) HasChannelPermission(userID, channelID, permission string) (bool, error) {
	if channelID == "" || userID == "" {
		return false, nil
	}

	var guildID string
	err := s.db.QueryRow("SELECT guild_id FROM channels WHERE channel_id = $1", channelID).Scan(&guildID)
	if err != nil {
		return false, err
	}
	if guildID == "" {
		return false, nil
	}

	return s.HasGuildPermission(userID, guildID, permission)
}

func (s *PermissionService) GetAllGuildPermissions(userID, guildID string) (map[string]bool, bool, error) {
	if guildID == "" || userID == "" {
		return make(map[string]bool), false, nil
	}

	var ownerID string
	err := s.db.QueryRow("SELECT owner_id FROM guilds WHERE guild_id = $1", guildID).Scan(&ownerID)
	if err != nil {
		return make(map[string]bool), false, err
	}

	isOwner := ownerID == userID
	permissions := map[string]bool{
		MANAGE_GUILD:    isOwner,
		MANAGE_ROLES:    isOwner,
		CREATE_CHANNEL:  isOwner,
		EDIT_CHANNEL:    isOwner,
		DELETE_CHANNEL:  isOwner,
		DELETE_MESSAGE:  isOwner,
		KICK_MEMBERS:    isOwner,
		CREATE_INVITE:   isOwner,
		VIEW_WEBHOOKS:   isOwner,
		CREATE_WEBHOOKS: isOwner,
		DELETE_WEBHOOKS: isOwner,
	}

	return permissions, isOwner, nil
}

func (s *PermissionService) HasAnyGuildPermission(userID, guildID string, permissions []string) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := s.HasGuildPermission(userID, guildID, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

func (s *PermissionService) RequireChannelPermission(c echo.Context, userID, channelID, permission string) (string, error) {
	hasPermission, err := s.HasChannelPermission(userID, channelID, permission)
	if err != nil {
		return "", s.sendErrorResponse(c, ErrChannelNotFound)
	}
	if !hasPermission {
		return "", s.sendErrorResponse(c, ErrUnauthorized)
	}

	var guildID string
	s.db.QueryRow("SELECT guild_id FROM channels WHERE channel_id = $1", channelID).Scan(&guildID)
	return guildID, nil
}

func (s *PermissionService) RequireWebhookPermission(c echo.Context, userID, webhookID, permission string) (string, string, error) {
	var channelID, guildID string
	err := s.db.QueryRow("SELECT c.channel_id, c.guild_id FROM webhooks w JOIN channels c ON w.channel_id = c.channel_id WHERE w.webhook_id = $1", webhookID).Scan(&channelID, &guildID)
	if err != nil {
		return "", "", s.sendErrorResponse(c, ErrChannelNotFound)
	}

	hasPermission, err := s.HasGuildPermission(userID, guildID, permission)
	if err != nil {
		return channelID, guildID, s.sendErrorResponse(c, ErrDatabaseError)
	}
	if !hasPermission {
		return channelID, guildID, s.sendErrorResponse(c, ErrUnauthorized)
	}

	return channelID, guildID, nil
}

func (s *PermissionService) RequireAnyGuildPermission(c echo.Context, userID, guildID string, permissions []string) error {
	if err := s.requireGuildMembership(c, userID, guildID); err != nil {
		return err
	}

	hasPermission, err := s.HasAnyGuildPermission(userID, guildID, permissions)
	if err != nil {
		return s.sendErrorResponse(c, ErrDatabaseError)
	}
	if !hasPermission {
		return s.sendErrorResponse(c, ErrInsufficientPermissions)
	}

	return nil
}