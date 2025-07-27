// Create utils/auth_utils.go from scratch with proper package declaration:

package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"auth.com/v4/cache"
	"github.com/labstack/echo/v4"
)

type UserData struct {
	Username string
	Password string
}

func Initialize() {
	InitAppConfig()

	err := InitDB()
	if err != nil {
		Log.Error("auth", "init_db", "Failed to initialize auth database", err)
	} else {
		Log.Info("auth", "initialize", "Database initialized successfully")
	}
}

func HandleAuthResult(c echo.Context, success bool, errCode int, userID string, isRegister bool) error {
	if !success {
		return SendErrorResponse(c, errCode)
	}

	err := SetAuthCookie(c, userID)
	if err != nil {
		return SendErrorResponse(c, ErrDatabaseError)
	}

	return nil
}
func PerformLogoutBySessionID(sessionID string) {
	if sessionID != "" {
		TerminateSessionWithNotification(sessionID, true)
	}
}

func PerformLogout(c echo.Context) {
	cookie, err := c.Cookie("session")
	if err == nil && cookie.Value != "" {
		sessionID, found, _ := GetSessionIDByToken(cookie.Value)
		if found {
			TerminateSessionWithNotification(sessionID, true)
		}
	}
	ClearAuthCookie(c)
}
func ExtractCredentials(c echo.Context) (string, string) {
	username := c.FormValue("username")
	password := c.FormValue("password")
	return username, password
}

func AsyncOperation(operation string, fn func() error) {
	go func() {
	if err := fn(); err != nil {
		Log.Error("auth", "async_operation", "Async operation failed", err, map[string]interface{}{
			"operation": operation,
		})
	} else {
		Log.Info("auth", "async_operation", "Async operation completed successfully", map[string]interface{}{
			"operation": operation,
		})
	}
	}()
}

func IsUserAdmin(userID string) (bool, error) {
	return cache.Service.CheckAdmin(userID)
}

func generateSuccessResponse(c echo.Context, userID string, isRegistration bool) error {
	successCode := RegisterSuccess
	if !isRegistration {
		successCode = LoginSuccess
	}

	baseMessage := ErrorMessages[successCode]
	finalMessage := strings.Replace(baseMessage, "{username}", userID, 1)

	return SendSuccessResponse(c, finalMessage)
}

func ProcessAuthRequest(c echo.Context, isRegistration bool) error {
	username, password := ExtractCredentials(c)

	valid, errCode := ValidateCredentials(username, password, ValidationContext{IsRegistration: isRegistration})
	if !valid {
		return SendErrorResponse(c, errCode)
	}

	if isRegistration {
		hashedPw, err := HashPassword(password)
		if err != nil {
			return SendErrorResponse(c, ErrDatabaseError)
		}

		user_id := GenerateSessionID("user")

		profilePicturePath := ""
		file, err := c.FormFile("profile_picture")
		if err == nil && file != nil {
			imageData, outputFormat, valid, _ := ValidateProfilePicture(file)
			if valid {
				os.MkdirAll("public/pfps", 0750)
				cleanUsername := strings.ReplaceAll(username, "/", "_")
				filename := fmt.Sprintf("%s_%d.%s", cleanUsername, time.Now().Unix(), outputFormat)
				dst := filepath.Join("public/pfps", filename)
				if err := os.WriteFile(dst, imageData, 0644); err == nil {
					profilePicturePath = "/public/pfps/" + filename
				}
			}
		}

		err = WithTransaction(func(tx *sql.Tx) error {
			_, err := tx.Exec("INSERT INTO users (user_id, username, password, profile_picture) VALUES ($1, $2, $3, $4)",
				user_id, username, hashedPw, profilePicturePath)
			return err
		})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return SendErrorResponse(c, ErrUserExists)
			}
			return SendErrorResponse(c, ErrDatabaseError)
		}

		cache.Provider.SetUser(username, hashedPw)
		if err := HandleAuthResult(c, true, 0, user_id, isRegistration); err != nil {
			return err
		}

	} else {
		// For login, get user_id from username first
		var userID string
		found, err := QueryRow("GetUserIDFromUsername", &userID,
			"SELECT user_id FROM users WHERE username = $1", username)
		if !found || err != nil {
			return SendErrorResponse(c, ErrInvalidCredentials)
		}

		// Check if user is banned BEFORE validating password
		isBanned, err := IsUserBanned(userID)
		if err == nil && isBanned {
			return SendErrorResponse(c, ErrAccountSuspended)
		}

		hashedPw, exists, err := cache.Provider.GetUser(username)
		if err != nil || !exists {
			return SendErrorResponse(c, ErrInvalidCredentials)
		}

		if err := VerifyPassword(hashedPw, password); err != nil {
			return SendErrorResponse(c, ErrInvalidCredentials)
		}
		if err := HandleAuthResult(c, true, 0, userID, isRegistration); err != nil {
			return err
		}

	}

	return generateSuccessResponse(c, username, isRegistration)
}

func ProcessLoginRequest(c echo.Context) error {
	return ProcessAuthRequest(c, false)
}
func ProcessRegistrationRequest(c echo.Context) error {
	return ProcessAuthRequest(c, true)
}
