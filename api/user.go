package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"auth.com/v4/cache"
	"auth.com/v4/utils"
	"auth.com/v4/internal/perms"
	"github.com/labstack/echo/v4"
)

func GetCurrentUser(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	if !strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		return echo.NewHTTPError(404)
	}

	username, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	var profilePicture sql.NullString
	var bio sql.NullString
	var createdAt time.Time

	err = utils.GetDB().QueryRow(`
    SELECT profile_picture, bio, created_at 
    FROM users WHERE user_id = $1
`, userID).Scan(&profilePicture, &bio, &createdAt)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	profilePicturePath := ""
	if profilePicture.Valid {
		profilePicturePath = profilePicture.String
	}

	bioValue := ""
	if bio.Valid {
		bioValue = bio.String
	}

	return c.JSON(http.StatusOK, echo.Map{
		"user_id":         userID,
		"username":        username,
		"status":          "active",
		"profile_picture": profilePicturePath,
		"bio":             bioValue,
		"created_at":      createdAt,
	})
}

func GetUserList(c echo.Context) error {
	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := utils.AppConfig.UsersPerPage
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	users, totalCount, err := utils.GetAllUsers(page, limit)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"users":       users,
		"count":       len(users),
		"total_count": totalCount,
	})
}

func MakeUserAdmin(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	// Get user_id from username
	var userID string
	found, err := utils.QueryRow("GetUserIDFromUsername", &userID,
		"SELECT user_id FROM users WHERE username = $1", username)
	if !found || err != nil {
		return utils.SendErrorResponse(c, utils.ErrUserNotFound)
	}

	err = utils.SetUserAdminStatus(userID, true)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	cache.Provider.DeleteUser(username)

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "User " + username + " is now an admin",
	})
}

func DemoteUserAdmin(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	var userID string
	found, err := utils.QueryRow("GetUserIDFromUsername", &userID,
		"SELECT user_id FROM users WHERE username = $1", username)
	if !found || err != nil {
		return utils.SendErrorResponse(c, utils.ErrUserNotFound)
	}

	err = utils.SetUserAdminStatus(userID, false)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	cache.Provider.DeleteUser(username)

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "User " + username + " is no longer an admin",
	})
}

func UploadProfilePictureHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	// Get username for filename
	username, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	file, err := c.FormFile("profile_picture")
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	imageData, outputFormat, valid, errCode := utils.ValidateProfilePicture(file)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}
	os.MkdirAll("public/pfps", 0750)

	cleanUsername := strings.ReplaceAll(username, "/", "_")
	cleanUsername = strings.ReplaceAll(cleanUsername, "\\", "_")
	cleanUsername = strings.ReplaceAll(cleanUsername, "..", "_")

	filename := fmt.Sprintf("%s_%d.%s",
		cleanUsername,
		time.Now().Unix(),
		outputFormat)

	dst := filepath.Join("public/pfps", filename)

	if !strings.HasPrefix(filepath.Clean(dst), "public/pfps/") {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	err = os.WriteFile(dst, imageData, 0644)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	profilePicturePath := "/public/pfps/" + filename
	err = utils.ExecuteQuery(
		"UpdateUserProfilePicture",
		"UPDATE users SET profile_picture = $1 WHERE user_id = $2",
		profilePicturePath, userID)

	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return utils.SendSuccessResponse(c, "Profile picture uploaded successfully")
}

func GetUserGuilds(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guilds, err := utils.GetUserGuilds(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"guilds": guilds,
		"count":  len(guilds),
	})
}

func UpdateUsernameHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	var request struct {
		Username string `json:"username"`
	}

	if err := c.Bind(&request); err != nil {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	newUsername := strings.TrimSpace(request.Username)
	if newUsername == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	if len(newUsername) < 3 || len(newUsername) > 32 {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	oldUsername, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	var existingUserID string
	found, err := utils.QueryRow("CheckUsernameExists", &existingUserID,
		"SELECT user_id FROM users WHERE username = $1 AND user_id != $2", newUsername, userID)
	if found {
		return utils.SendErrorResponse(c, utils.ErrUserExists)
	}

	err = utils.ExecuteQuery("UpdateUsername",
		"UPDATE users SET username = $1 WHERE user_id = $2",
		newUsername, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}
	cache.Provider.DeleteUser(oldUsername)
	cache.Provider.DeleteUser(newUsername)
	cache.Provider.Delete(fmt.Sprintf("username:%s", userID))

	usernameChangeData := map[string]interface{}{
		"type":         "username_changed",
		"user_id":      userID,
		"old_username": oldUsername,
		"new_username": newUsername,
	}

	userGuilds, err := utils.GetUserGuilds(userID)
	if err == nil {
		for _, guild := range userGuilds {
			if guildID, ok := guild["guild_id"].(string); ok {
				usernameChangeData["guild_id"] = guildID
				messageBytes, _ := json.Marshal(usernameChangeData)
				utils.BroadcastWithRedis(1, messageBytes)
			}
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "Username updated successfully",
	})
}

func UpdateBioHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	var request struct {
		Bio string `json:"bio"`
	}

	if err := c.Bind(&request); err != nil {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	newBio := strings.TrimSpace(request.Bio)

	valid, errCode := utils.ValidateBio(newBio)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	err = utils.ExecuteQuery("UpdateBio",
		"UPDATE users SET bio = $1 WHERE user_id = $2",
		newBio, userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "Bio updated successfully",
	})
}

func GetUserProfileHandler(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	username, err := utils.GetUsernameByID(userID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	var profilePicture sql.NullString
	var bio sql.NullString
	err = utils.GetDB().QueryRow(`
		SELECT profile_picture, bio 
		FROM users WHERE user_id = $1
	`, userID).Scan(&profilePicture, &bio)

	profilePictureValue := ""
	if err == nil && profilePicture.Valid {
		profilePictureValue = profilePicture.String
	}

	bioValue := ""
	if err == nil && bio.Valid {
		bioValue = bio.String
	}

	return c.JSON(200, map[string]interface{}{
		"username":        username,
		"profile_picture": profilePictureValue,
		"bio":             bioValue,
	})
}

func CheckGuildOwnershipHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	guildID, err := utils.RequireParam(c, "id")
	if err != nil {
		return err
	}

	// Check if user is guild owner
	hasPermission, err := perms.Service.HasGuildPermission(userID, guildID, perms.MANAGE_GUILD)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"is_owner": hasPermission,
	})
}
