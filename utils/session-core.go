package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"auth.com/v4/cache"
	"github.com/labstack/echo/v4"
)

func GenerateSessionID(entityType string) string {
	var prefix string
	switch entityType {
	case "user":
		prefix = "1"
	case "guild":
		prefix = "2"
	case "channel":
		prefix = "3"
	case "session":
		prefix = "4"
	case "message":
		prefix = "5"
	case "webhook":
		prefix = "6"
	default:
		prefix = "9"
	}

	max := 99999999999999 // 14 digits
	min := 10000000000000 // 14 digits

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()%int64(max-min+1)+int64(min))
	}

	sessionID := fmt.Sprintf("%s%d", prefix, n.Int64()+int64(min))

	return sessionID
}

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func CreateSession(token, userID string) error {
	sessionID := GenerateSessionID("session")
	expiresAt := time.Now().Add(AppConfig.SessionExpiryDuration)

	err := cache.Provider.SetSession(sessionID, userID, AppConfig.SessionExpiryDuration)
	if err != nil {
		return err
	}

	username, _ := GetUsernameByID(userID)
	isAdmin, _ := IsUserAdmin(userID)
	isBanned, _ := IsUserBanned(userID)

	sessionData := &cache.SessionData{
		UserID:    userID,
		Username:  username,
		IsAdmin:   isAdmin,
		IsBanned:  isBanned,
		ExpiresAt: expiresAt,
	}

	cache.Provider.SetSessionWithUser(token, sessionData, AppConfig.SessionExpiryDuration)

	return ExecuteQuery("CreateSession",
		"INSERT INTO sessions (token, user_id, session_id, expires_at) VALUES ($1, $2, $3, $4)",
		token, userID, sessionID, expiresAt)
}

func GetSessionUser(token string) (string, bool, error) {
	sessionID, found, err := GetSessionIDByToken(token)
	if !found || err != nil {
		return "", false, err
	}
	return GetUserBySessionID(sessionID)
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

func GetSessionIDByToken(token string) (string, bool, error) {
	var sessionID string
	found, err := QueryRow("GetSessionIDByToken", &sessionID,
		"SELECT session_id FROM sessions WHERE token = $1 AND expires_at > NOW()", token)
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

func GetSessionIDByTokenNoExpiry(token string) (string, bool, error) {
	var sessionID string
	found, err := QueryRow("GetSessionIDByTokenNoExpiry", &sessionID,
		"SELECT session_id FROM sessions WHERE token = $1", token)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return sessionID, false, nil
	}
	return sessionID, found, err
}

func CentralRefreshSession(c echo.Context, userID string) (time.Time, error) {
	return GlobalSessionManager.RefreshSession(c, userID)
}