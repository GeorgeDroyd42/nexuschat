package utils

import (
	"net/http"
	"time"
	"auth.com/v4/cache"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

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
		logrus.WithFields(logrus.Fields{"module": "session", "action": "generate_token"}).WithError(err).Error("Failed to generate session token")
		return err
	}

	err = CreateSession(token, userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"module": "session",
			"action": "create_session",
		}).WithError(err).Error("Failed to create session")
		return err
	}

	cookie := createSessionCookie(token, time.Now().Add(AppConfig.SessionExpiryDuration))
	c.SetCookie(cookie)

	logrus.WithFields(logrus.Fields{
		"module": "session",
		"action": "set_auth_cookie", 
		"user_id": userID,
	}).Info("Auth cookie set for user")
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

func ValidateUserSession(c echo.Context) (string, bool, error) {
	cookie, err := c.Cookie("session")
	if err != nil {
		return "", false, err
	}

	token := cookie.Value
	return ValidateSessionToken(token)
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

