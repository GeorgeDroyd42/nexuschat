package utils

import (
	"time"
	"fmt"
	"auth.com/v4/cache"

)

func GetSessionIDByToken(token string, checkExpiry bool) (string, bool, error) {
	var sessionID string
	var query string
	
	if checkExpiry {
		query = "SELECT session_id FROM sessions WHERE token = $1 AND expires_at > NOW()"
	} else {
		query = "SELECT session_id FROM sessions WHERE token = $1"
	}
	
	found, err := QueryRow("GetSessionIDByToken", &sessionID, query, token)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return sessionID, false, nil
	}
	return sessionID, found, err
}

func GetUserBySessionID(sessionID string) (string, bool, error) {
	var userID string
	found, err := QueryRow("GetUserBySessionID", &userID,
		"SELECT user_id FROM sessions WHERE session_id = $1 AND expires_at > NOW()", sessionID)
	return userID, found, err
}

func GetTokenBySessionID(sessionID string) (string, bool, error) {
	var token string
	found, err := QueryRow("GetTokenBySessionID", &token, 
		"SELECT token FROM sessions WHERE session_id = $1", sessionID)
	return token, found, err
}

func DeleteSession(sessionID string) error {
	return ExecuteQuery("DeleteSession", "DELETE FROM sessions WHERE session_id = $1", sessionID)
}

func GetSessionUserID(token string) (string, bool, error) {
	var userID string
	found, err := QueryRow("GetSessionUserID", &userID,
		"SELECT user_id FROM sessions WHERE token = $1", token)
	return userID, found, err
}

func ExtendSession(token string, newExpiresAt time.Time) error {
	return ExecuteQuery("ExtendSession",
		"UPDATE sessions SET expires_at = $1 WHERE token = $2",
		newExpiresAt, token)
}

func CreateSession(token, userID, sessionID string, expiresAt time.Time) error {
	return ExecuteQuery("CreateSession",
		"INSERT INTO sessions (token, user_id, session_id, expires_at) VALUES ($1, $2, $3, $4)",
		token, userID, sessionID, expiresAt)
}

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
		var token string
		token, _, _ = GetTokenBySessionID(sessionID)
		cache.Provider.DeleteSession(sessionID)
		cache.Provider.DeleteSessionToken(token)
	}
}