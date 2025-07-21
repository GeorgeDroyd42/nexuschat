package csrf

import "time"

type Provider interface {
	GetCSRFToken(sessionID string) (string, bool, error)
	SetCSRFToken(sessionID, token string, expiration time.Duration) error
	DeleteCSRFToken(sessionID string) error
}

type KeyGenerator interface {
	CSRFToken(sessionID string) string
}