package utils

import (
	"time"
	"fmt"
	"auth.com/v4/cache"
)

func GetSessionsByUserID(userID string) ([]string, error) {
	return CacheFirstQuery(
		fmt.Sprintf("user_sessions:%s", userID),
		5*time.Minute,
		func() ([]string, bool, error) {
			sessionIDs := []string{}
			rows, err := GetDB().Query("SELECT session_id FROM sessions WHERE user_id = $1", userID)
			if err != nil {
				return sessionIDs, false, err
			}
			defer rows.Close()

			for rows.Next() {
				var sessionID string
				if err := rows.Scan(&sessionID); err != nil {
					return sessionIDs, false, err
				}
				sessionIDs = append(sessionIDs, sessionID)
			}
			return sessionIDs, len(sessionIDs) > 0, nil
		})
}

func updateUserSessionAdminStatus(userID string, isAdmin bool) {
	rows, _ := GetDB().Query("SELECT token FROM sessions WHERE user_id = $1", userID)
	defer rows.Close()
	for rows.Next() {
		var token string
		rows.Scan(&token)
		sessionData, found, _ := cache.Provider.GetSessionWithUser(token)
		if found {
			sessionData.IsAdmin = isAdmin
			cache.Provider.SetSessionWithUser(token, sessionData, time.Hour)
		}
	}
}

func ClearUserSessionCache(userID string) {
	sessionIDs, _ := GetSessionsByUserID(userID)
	for _, sessionID := range sessionIDs {
		token, _, _ := GetTokenBySessionID(sessionID)
		CleanupSession(sessionID, token)
	}
}