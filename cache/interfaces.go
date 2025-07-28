package cache

import (
	"time"
)

type SessionData struct {
	UserID    string    `json:"user_id"`            // Primary user identifier
	Username  string    `json:"username,omitempty"` // For display purposes
	IsAdmin   bool      `json:"is_admin"`           // Admin status (cached)
	IsBanned  bool      `json:"is_banned"`          // Ban status (cached)
	ExpiresAt time.Time `json:"expires_at"`         // Session expiration
}

// CacheProvider defines the core interface for cache operations
type CacheProvider interface {
	// Basic operations
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) (bool, error)
	Delete(key string) error

	// User-related operations
	GetUser(username string) (string, bool, error)
	SetUser(username, hashedPw string) error
	DeleteUser(username string) error

	// Admin-related operations

	GetAdmin(userID string) (bool, bool, error)
	SetAdmin(userID string, isAdmin bool) error
	DeleteAdmin(userID string) error

	// Session-related operations
	GetSession(sessionID string) (string, bool, error)
	GetSessionFromDB(sessionID string) (string, bool, error)
	SetSession(sessionID, userID string, expiration time.Duration) error
	DeleteSession(sessionID string) (bool, error)
	GetSessionWithUser(token string) (*SessionData, bool, error)
	SetSessionWithUser(token string, data *SessionData, ttl time.Duration) error
	DeleteSessionToken(token string) error
	// User sessions and ban operations
	GetUserSessions(userID string) ([]string, bool, error)
	SetUserSessions(userID string, sessionIDs []string, expiration time.Duration) error
	GetUserBan(userID string) (bool, bool, error)
	SetUserBan(userID string, isBanned bool, expiration time.Duration) error
	DeleteUserBan(userID string) error
	GetWebSocketConnections(userID string) ([]string, bool, error)
	AddWebSocketConnection(userID, sessionID string, data map[string]string, expiration time.Duration) error
	RemoveWebSocketConnection(userID, sessionID string) error
	GetWebSocketConnectionData(sessionID string) (map[string]string, bool, error)
	AddTypingUser(channelID, userID string, duration time.Duration) error
	RemoveTypingUser(channelID, userID string) error
	GetTypingUsers(channelID string) ([]string, error)
	RemoveUserFromAllTypingIndicators(userID string) ([]string, error)
	PublishMessage(channel string, message interface{}) error
	Subscribe(channels ...string) (PubSubSubscription, error)
	SetNX(key string, value interface{}, expiration time.Duration) bool
}

type KeyGenerator interface {
	User(username string) string
	GuildOnlineUsers(guildID string) string
	GuildOfflineUsers(guildID string) string	
	Admin(userID string) string
	Session(tokenOrID string) string
	UserSessions(userID string) string
	WebSocket(userID string) string
	UserBan(userID string) string
	WebSocketConnection(sessionID string) string
	SessionData(token string) string
	TypingIndicator(channelID string) string
}

// PubSubSubscription defines the interface for PubSub subscriptions
type PubSubSubscription interface {
	ReceiveMessage() (PubSubMessage, error)
	Close() error
}

// PubSubMessage defines the interface for PubSub messages
type PubSubMessage interface {
	GetChannel() string
	GetPayload() string
}
