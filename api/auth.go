package api

import (
	"net/http"

	"auth.com/v4/internal/csrf"
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

func RegisterHandler(c echo.Context) error {
	file, err := c.FormFile("profile_picture")
	if err != nil && err != http.ErrMissingFile {
		return utils.SendErrorResponse(c, utils.ErrInvalidFileType)
	}

	if file != nil {
		_, _, valid, errCode := utils.ValidateProfilePicture(file)
		if !valid {
			return utils.SendErrorResponse(c, errCode)
		}
	}
	return utils.ProcessRegistrationRequest(c)
}

func LoginHandler(c echo.Context) error {
	return utils.ProcessLoginRequest(c)
}

func LogoutHandler(c echo.Context) error {
	utils.PerformLogout(c)

	return utils.SendSuccessResponse(c, utils.ErrorMessages[utils.LogoutSuccess])
}

func GetCSRFToken(c echo.Context) error {
	token := c.Get("csrf").(string)

	cookie, err := c.Cookie("session")
	if err == nil && cookie.Value != "" {
		csrf.Service.StoreToken(cookie.Value, token)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"csrf_token": token,
	})
}

func RefreshSessionHandler(c echo.Context) error {
	// Debug: Print the cookie being received
	cookie, err := c.Cookie("session")
if err == nil {
			utils.Log.Debug("auth", "session_refresh", "Session refresh received cookie", map[string]interface{}{"cookie_prefix": cookie.Value[:12] + "..."})
		} else {
			utils.Log.Error("auth", "session_refresh", "Session refresh missing cookie", err)
		}

	userID := utils.GetUserID(c)
	if userID == "" {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	newExpiry, err := utils.CentralRefreshSession(c, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success":    true,
		"expires_at": newExpiry.Unix(),
	})
}

func BanUserHandler(c echo.Context) error {
	userID := c.Param("userid")

	if userID == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	err := utils.SetUserBanStatus(userID, true)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return utils.SendSuccessResponse(c, "User banned successfully")
}
func UnbanUserHandler(c echo.Context) error {
	userID := c.Param("userid")
	if userID == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	// Check if user exists using existing function
	_, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrUserNotFound)
	}

	err = utils.SetUserBanStatus(userID, false)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return utils.SendSuccessResponse(c, "User unbanned successfully")
}
