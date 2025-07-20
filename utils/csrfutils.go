package utils

import (
	"time"

	"auth.com/v4/cache"
)

func StoreCSRFToken(sessionID string, token string) {
	cache.Provider.SetCSRFToken(sessionID, token, 1*time.Hour)
}

func InvalidateCSRFToken(sessionID string) {
	cache.Provider.DeleteCSRFToken(sessionID)
}

func GetCSRFToken(sessionID string) string {
	token, found, err := cache.Provider.GetCSRFToken(sessionID)
	if err != nil || !found {
		return ""
	}
	return token
}
