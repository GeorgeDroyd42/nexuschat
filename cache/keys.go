package cache

import "fmt"

// StandardKeyGenerator implements the KeyGenerator interface
type StandardKeyGenerator struct{}

func (kg StandardKeyGenerator) User(username string) string {
	return fmt.Sprintf("user:data:%s", username)
}

func (kg StandardKeyGenerator) Admin(userID string) string {
	return fmt.Sprintf("user:admin:%s", userID)
}

func (kg StandardKeyGenerator) Session(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (kg StandardKeyGenerator) UserSessions(userID string) string {
	return fmt.Sprintf("user:sessions:%s", userID)
}

func (kg StandardKeyGenerator) WebSocket(userID string) string {
	return fmt.Sprintf("ws:user:%s", userID)
}

func (kg StandardKeyGenerator) UserBan(userID string) string {
	return fmt.Sprintf("user:ban:%s", userID)
}

// DefaultKeys is the default key generator instance
var DefaultKeys = StandardKeyGenerator{}

func (kg StandardKeyGenerator) WebSocketConnection(sessionID string) string {
	return fmt.Sprintf("ws:connection:%s", sessionID)
}

func (kg StandardKeyGenerator) SessionData(token string) string {
	return fmt.Sprintf("session:data:%s", token)
}

func (kg StandardKeyGenerator) TypingIndicator(channelID string) string {
	return fmt.Sprintf("typing:%s", channelID)
}