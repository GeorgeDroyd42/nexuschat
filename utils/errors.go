// utils/responses.go (renamed from errors.go )
package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	StatusOK                  = http.StatusOK                  // 200
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusInternalServerError = http.StatusInternalServerError // 500
)

// Application error/success codes
const (
	ErrInvalidCredentials        = 1001
	ErrUserExists                = 1002
	ErrUnauthorized              = 1003
	RegisterSuccess              = 1004
	LoginSuccess                 = 1005
	LogoutSuccess                = 1006
	ErrDatabaseError             = 1007
	ErrInvalidUsername           = 1008
	ErrInvalidPassword           = 1009
	ErrUsernameRequired          = 1010
	ErrUsernameTooShort          = 1011
	ErrUsernameTooLong           = 1012
	ErrUsernameInvalidChar       = 1013
	ErrPasswordRequired          = 1014
	ErrPasswordTooShort          = 1015
	ErrPasswordTooLong           = 1016
	ErrPasswordNoCapital         = 1017
	ErrFileTooLarge              = 2001
	ErrInvalidFileType           = 2002
	ErrUserNotFound              = 2003
	ErrAccountSuspended          = 2004
	ErrGuildNotFound             = 2005
	ErrGuildNameTooShort         = 2006
	ErrGuildNameTooLong          = 2007
	ErrGuildDescriptionTooLong   = 2008
	ErrMaxGuildsReached          = 2009
	ErrUserNotInGuild            = 2010
	ErrEmptyMessage              = 2013
	ErrMessageTooLong            = 2011
	ErrChannelNotFound           = 2012
	ErrSessionTerminated         = 2014
	ErrChannelNameRequired       = 2015
	ErrChannelNameTooLong        = 2016
	ErrChannelDescriptionTooLong = 2017
	ErrNotGuildOwner             = 2018
	ErrInsufficientPermissions   = 2020
)

var ErrorStatusMap = map[int]int{
	ErrInvalidCredentials:  StatusUnauthorized,
	ErrUserExists:          StatusBadRequest,
	ErrUnauthorized:        StatusUnauthorized,
	ErrDatabaseError:       StatusInternalServerError,
	ErrInvalidUsername:     StatusBadRequest,
	ErrInvalidPassword:     StatusBadRequest,
	ErrUsernameRequired:    StatusBadRequest,
	ErrUsernameTooShort:    StatusBadRequest,
	ErrAccountSuspended:    StatusForbidden,
	ErrSessionTerminated:   StatusForbidden,
	ErrUsernameTooLong:     StatusBadRequest,
	ErrUsernameInvalidChar: StatusBadRequest,
	ErrPasswordRequired:    StatusBadRequest,
	ErrPasswordTooShort:    StatusBadRequest,
	ErrPasswordTooLong:     StatusBadRequest,
	ErrPasswordNoCapital:   StatusBadRequest,
}

var ErrorMessages = map[int]string{
	ErrInvalidCredentials:        "Invalid username or password.",
	ErrUserExists:                "User already exists.",
	ErrAccountSuspended:          "Your account has been suspended.",
	ErrSessionTerminated:         "Your session was terminated. Please log in again.",
	ErrUnauthorized:              "You do not have permission to access this resource.",
	ErrUserNotFound:              "User does not exist",
	ErrUserNotInGuild:            "You do not have permission to view info for this guild, as you are not in it",
	ErrGuildNameTooShort:         "Guild name must be at least 2 characters.",
	ErrMaxGuildsReached:          "You have reached the maximum limit of guilds.",
	ErrGuildNameTooLong:          "Guild name cannot exceed 50 characters.",
	ErrGuildDescriptionTooLong:   "Guild description cannot exceed 500 characters.",
	ErrGuildNotFound:             "Guild does not exist",
	ErrChannelNameRequired:       "Channel name is required.",
	ErrChannelNameTooLong:        "Channel name cannot exceed 30 characters.",
	ErrChannelDescriptionTooLong: "Channel description cannot exceed 200 characters.",
	ErrNotGuildOwner:             "Only the guild owner can create, edit, or delete channels.",
	RegisterSuccess:              "Sucessfully Registered",
	ErrInsufficientPermissions:   "Unauthorized",
	LoginSuccess:                 "Logged in Successfully as {username}",
	LogoutSuccess:                "Logged out successfully",
	ErrDatabaseError:             "Database error: {error}",
	ErrInvalidUsername:           "Invalid characters in username.",
	ErrInvalidPassword:           "Invalid password format.",
	ErrUsernameRequired:          "Username is required.",
	ErrUsernameTooShort:          "Username must be at least 3 characters.",
	ErrUsernameTooLong:           "Username cannot exceed 20 characters.",
	ErrUsernameInvalidChar:       "Username can only contain letters, numbers, and underscores.",
	ErrFileTooLarge:              "File is too large. Maximum size is 5MB.",
	ErrInvalidFileType:           "Invalid file type. Only JPG and PNG images are allowed.",
	ErrPasswordRequired:          "Password is required.",
	ErrPasswordTooShort:          "Password must be at least 6 characters.",
	ErrPasswordTooLong:           "Password cannot exceed 100 characters.",
	ErrPasswordNoCapital:         "Password must contain at least one capital letter.",
}

func SendErrorResponse(c echo.Context, errCode int) error {
	statusCode := GetStatusForError(errCode)
	return c.JSON(statusCode, echo.Map{"error": ErrorMessages[errCode]})
}

func SendSuccessResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "success", "message": message})
}

func GetStatusForError(errCode int) int {
	if status, exists := ErrorStatusMap[errCode]; exists {
		return status
	}
	return StatusInternalServerError
}
