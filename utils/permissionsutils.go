package utils

import "github.com/labstack/echo/v4"

func HasGuildPermission(userID, guildID, permission string) (bool, error) {
	if guildID == "" || userID == "" {
		return false, nil
	}

	// Check if user is guild owner (owners have all permissions)
	guild, found, err := GetGuildByID(guildID)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	if guild["owner_id"] == userID {
		return true, nil
	}

	// Check role-based permissions
	var hasPermission bool
	err = GetDB().QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM guild_member_roles gmr
			JOIN guild_roles gr ON gmr.role_id = gr.role_id
			WHERE gmr.guild_id = $1 AND gmr.user_id = $2 
			AND $3 = ANY(gr.permissions)
		)
	`, guildID, userID, permission).Scan(&hasPermission)

	return hasPermission, err
}

func HasChannelPermission(userID, channelID, permission string) (bool, error) {
	if channelID == "" || userID == "" {
		return false, nil
	}

	var guildID string
	err := GetDB().QueryRow("SELECT guild_id FROM channels WHERE channel_id = $1", channelID).Scan(&guildID)
	if err != nil {
		return false, err
	}
	if guildID == "" {
		return false, nil
	}

	return HasGuildPermission(userID, guildID, permission)
}

func GetAllGuildPermissions(userID, guildID string) (map[string]bool, bool, error) {
	if guildID == "" || userID == "" {
		return make(map[string]bool), false, nil
	}

	guild, found, err := GetGuildByID(guildID)
	if err != nil || !found {
		return make(map[string]bool), false, err
	}

	isOwner := guild["owner_id"] == userID
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

	if isOwner {
		return permissions, true, nil
	}

	rows, err := GetDB().Query(`
		SELECT DISTINCT UNNEST(gr.permissions) as permission
		FROM guild_member_roles gmr
		JOIN guild_roles gr ON gmr.role_id = gr.role_id
		WHERE gmr.guild_id = $1 AND gmr.user_id = $2
	`, guildID, userID)

	if err != nil {
		return permissions, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			continue
		}
		permissions[permission] = true
	}

	return permissions, isOwner, nil
}

func CreateDefaultRoles(guildID string) error {
	adminRoleID := GenerateSessionID("role")
	modRoleID := GenerateSessionID("role")

	_, err := GetDB().Exec(`
		INSERT INTO guild_roles (role_id, guild_id, name, permissions, color, position) VALUES
		($1, $2, 'Administrator', $3, '#e74c3c', 100),
		($4, $2, 'Moderator', $5, '#f39c12', 50)
	`, adminRoleID, guildID,
		AdminRolePermissions,
		modRoleID,
		ModeratorRolePermissions)

	return err
}

func HasAnyGuildPermission(userID, guildID string, permissions []string) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := HasGuildPermission(userID, guildID, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}
func RequireChannelPermission(c echo.Context, userID, channelID, permission string) (string, error) {
	hasPermission, err := HasChannelPermission(userID, channelID, permission)
	if err != nil {
		return "", SendErrorResponse(c, ErrChannelNotFound)
	}
	if !hasPermission {
		return "", SendErrorResponse(c, ErrUnauthorized)
	}

	// Get guild ID for return value (some callers need it)
	var guildID string
	GetDB().QueryRow("SELECT guild_id FROM channels WHERE channel_id = $1", channelID).Scan(&guildID)
	return guildID, nil
}
func RequireWebhookPermission(c echo.Context, userID, webhookID, permission string) (string, string, error) {
	var channelID, guildID string
	err := GetDB().QueryRow("SELECT c.channel_id, c.guild_id FROM webhooks w JOIN channels c ON w.channel_id = c.channel_id WHERE w.webhook_id = $1", webhookID).Scan(&channelID, &guildID)
	if err != nil {
		return "", "", SendErrorResponse(c, ErrChannelNotFound)
	}

	hasPermission, err := HasGuildPermission(userID, guildID, permission)
	if err != nil {
		return channelID, guildID, SendErrorResponse(c, ErrDatabaseError)
	}
	if !hasPermission {
		return channelID, guildID, SendErrorResponse(c, ErrUnauthorized)
	}

	return channelID, guildID, nil
}
func RequireAnyGuildPermission(c echo.Context, userID, guildID string, permissions []string) error {
	if err := RequireGuildMembership(c, userID, guildID); err != nil {
		return err
	}

	hasPermission, err := HasAnyGuildPermission(userID, guildID, permissions)
	if err != nil {
		return SendErrorResponse(c, ErrDatabaseError)
	}
	if !hasPermission {
		return SendErrorResponse(c, ErrInsufficientPermissions)
	}

	return nil
}
