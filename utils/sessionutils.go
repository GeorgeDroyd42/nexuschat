package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"auth.com/v4/cache"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func createSessionCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     AppConfig.SessionCookieName,
		Value:    token,
		Path:     AppConfig.SessionCookiePath,
		HttpOnly: AppConfig.SessionCookieHTTPOnly,
		Secure:   AppConfig.SessionCookieSecure,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	}
}

func SetAuthCookie(c echo.Context, userID string) error {
	token, err := GenerateSessionToken()
	if err != nil {
		log(logrus.ErrorLevel, ModuleSession, "GenerateSessionToken", "", err)
		return err
	}

	err = CreateSession(token, userID)
	if err != nil {
		log(logrus.ErrorLevel, ModuleSession, "CreateSession", "", err)
		return err
	}

	cookie := createSessionCookie(token, time.Now().Add(AppConfig.SessionExpiryDuration))
	c.SetCookie(cookie)

	log(logrus.InfoLevel, ModuleSession, "SetAuthCookie", "Auth cookie set for user: "+userID, nil)
	return nil
}
func ClearAuthCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
}

func ValidateSessionToken(token string) (string, bool, error) {
	sessionData, found, err := cache.Provider.GetSessionWithUser(token)
	if found && err == nil && time.Now().Before(sessionData.ExpiresAt) {
		return sessionData.UserID, true, nil
	}

	sessionID, found, err := GetSessionIDByToken(token)
	if !found || err != nil {
		return "", false, err
	}

	userID, found, err := GetUserBySessionID(sessionID)
	if !found || err != nil {
		return "", false, err
	}
	return userID, true, nil
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
func ExtendSession(token string, duration time.Duration) error {
	newExpiresAt := time.Now().Add(duration)
	fmt.Printf("📅 Session extended to: %s\n", newExpiresAt.Format("15:04:05"))

	sessionID, found, err := GetSessionIDByTokenNoExpiry(token)
	if !found || err != nil {
		fmt.Printf("❌ ExtendSession failed - GetSessionIDByToken: found=%v err=%v\n", found, err)
		return err
	}
	userID, found, err := GetUserBySessionID(sessionID)
	if !found || err != nil {
		return err
	}

	err = ExecuteQuery("ExtendSession",
		"UPDATE sessions SET expires_at = $1 WHERE token = $2",
		newExpiresAt, token)
	if err != nil {
		fmt.Printf("❌ ExtendSession failed - Database UPDATE: %v\n", err)
		return err
	}

	err = cache.Provider.SetSession(sessionID, userID, duration)
	if err != nil {
		return err
	}

	sessionData, found, _ := cache.Provider.GetSessionWithUser(token)
	if found {
		sessionData.ExpiresAt = newExpiresAt
		cache.Provider.SetSessionWithUser(token, sessionData, duration)
	}

	return nil
}
func ValidateUserSession(c echo.Context) (string, bool, error) {
	cookie, err := c.Cookie("session")
	if err != nil {
		return "", false, err
	}

	token := cookie.Value
	return ValidateSessionToken(token)
}

func TerminateSession(sessionID string) (string, bool, error) {
	return TerminateSessionWithNotification(sessionID, false)
}

func TerminateSessionWithNotification(sessionID string, sendNotification bool) (string, bool, error) {
	userID, found, _ := GetUserBySessionID(sessionID)

	var token string
	QueryRow("GetTokenBySessionID", &token, "SELECT token FROM sessions WHERE session_id = $1", sessionID)

	cache.Provider.DeleteSession(sessionID)
	cache.Provider.DeleteSessionToken(token)

	ExecuteQuery("DeleteSession", "DELETE FROM sessions WHERE session_id = $1", sessionID)

	if found && sendNotification {
		SendEventToSpecificSession(userID, token, "session_terminated", ErrorMessages[ErrSessionTerminated])
	}
	return userID, true, nil
}
func TerminateAllUserSessions(userID string) (bool, error) {
	sessions, err := GetSessionsByUserID(userID)
	if err != nil {
		return false, err
	}

	for _, sessionID := range sessions {
		var token string
		QueryRow("GetTokenBySessionID", &token, "SELECT token FROM sessions WHERE session_id = $1", sessionID)

		cache.Provider.DeleteSession(sessionID)
		cache.Provider.DeleteSessionToken(token)
	}

	eventData := map[string]interface{}{"type": "all_sessions_terminated"}
	broadcastData, _ := json.Marshal(eventData)
	SendToUser(userID, websocket.TextMessage, broadcastData)
	return true, nil
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

func CentralRefreshSession(c echo.Context, userID string) (time.Time, error) {
	return GlobalSessionManager.RefreshSession(c, userID)
}
