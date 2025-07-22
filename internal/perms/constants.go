package perms

// Guild Permissions - Single source of truth for all permission definitions
const (
	MANAGE_GUILD    = "manage_guild"
	MANAGE_ROLES    = "manage_roles"
	CREATE_CHANNEL  = "create_channel"
	EDIT_CHANNEL    = "edit_channel"
	DELETE_CHANNEL  = "delete_channel"
	DELETE_MESSAGE  = "delete_message"
	KICK_MEMBERS    = "kick_members"
	CREATE_INVITE   = "create_invite"
	VIEW_WEBHOOKS   = "view_webhooks"
	CREATE_WEBHOOKS = "create_webhooks"
	DELETE_WEBHOOKS = "delete_webhooks"
)

// Frontend permission names mapping
var FrontendPermissions = map[string]string{
	MANAGE_GUILD:    "canManageGuild",
	MANAGE_ROLES:    "canManageRoles",
	CREATE_CHANNEL:  "canCreateChannels",
	EDIT_CHANNEL:    "canEditChannels",
	DELETE_CHANNEL:  "canDeleteChannels",
	DELETE_MESSAGE:  "canDeleteMessages",
	KICK_MEMBERS:    "canKickMembers",
	CREATE_INVITE:   "canCreateInvite",
	VIEW_WEBHOOKS:   "canViewWebhooks",
	CREATE_WEBHOOKS: "canCreateWebhooks",
	DELETE_WEBHOOKS: "canDeleteWebhooks",
}

// Default role permissions
var AdminRolePermissions = []string{
	MANAGE_GUILD, MANAGE_ROLES, CREATE_CHANNEL, EDIT_CHANNEL, DELETE_CHANNEL,
	DELETE_MESSAGE, KICK_MEMBERS, CREATE_INVITE, VIEW_WEBHOOKS, CREATE_WEBHOOKS, DELETE_WEBHOOKS,
}

var ModeratorRolePermissions = []string{
	EDIT_CHANNEL, DELETE_MESSAGE, KICK_MEMBERS, VIEW_WEBHOOKS,
}

var DefaultUserPermissions = []string{}