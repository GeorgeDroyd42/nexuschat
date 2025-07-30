package utils

import (
	"bytes"
	"fmt"
	"image/gif"
	"io"
	"auth.com/v4/internal/perms"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ValidationContext struct {
	IsRegistration bool
	IsGuildName    bool
	IsDescription  bool
}

const (
	UsernameMinLength           = 3
	UsernameMaxLength           = 20
	PasswordMinLength           = 6
	PasswordMaxLength           = 100
	GuildNameMinLength          = 2
	GuildNameMaxLength          = 50
	GuildDescriptionMaxLength   = 500
	ChannelNameMaxLength        = 30
	ChannelDescriptionMaxLength = 200
	BioMaxLength                = 2000
)

func IsValidUsernameFormat(username string) bool {
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return true
}

func HasCapitalLetter(password string) bool {
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			return true
		}
	}
	return false
}

func ValidateBio(bio string) (bool, int) {
	if len(bio) > BioMaxLength {
		return false, ErrInvalidCredentials
	}
	return true, 0
}

func ValidateGuildData(name, description string) (bool, int) {
	if name == "" {
		return false, ErrGuildNameTooShort
	}

	if len(name) < GuildNameMinLength {
		return false, ErrGuildNameTooShort
	}

	if len(name) > GuildNameMaxLength {
		return false, ErrGuildNameTooLong
	}

	if len(description) > GuildDescriptionMaxLength {
		return false, ErrGuildDescriptionTooLong
	}

	return true, 0
}

func ValidateChannelData(name, description string) (bool, int) {
	name = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(name, " "))
	
	if name == "" {
		return false, ErrChannelNameRequired
	}

	if len(name) > ChannelNameMaxLength {
		return false, ErrChannelNameTooLong
	}

	if len(description) > ChannelDescriptionMaxLength {
		return false, ErrChannelDescriptionTooLong
	}

	return true, 0
}

func ValidateUsername(username string, ctx ValidationContext) (bool, int) {
	if username == "" {
		return false, ErrInvalidCredentials
	}

	if len(username) < UsernameMinLength {
		return false, ErrUsernameTooShort
	}

	if len(username) > UsernameMaxLength {
		return false, ErrUsernameTooLong
	}

	if !IsValidUsernameFormat(username) {
		return false, ErrUsernameInvalidChar
	}

	return true, 0
}

func ValidatePassword(password string, ctx ValidationContext) (bool, int) {
	if password == "" {
		return false, ErrInvalidPassword
	}

	if ctx.IsRegistration {
		if len(password) < PasswordMinLength {
			return false, ErrPasswordTooShort
		}

		if !HasCapitalLetter(password) {
			return false, ErrPasswordNoCapital
		}
	}

	if len(password) > PasswordMaxLength {
		return false, ErrPasswordTooLong
	}

	return true, 0
}

func ValidateCredentials(username, password string, ctx ValidationContext) (bool, int) {
	valid, errCode := ValidateUsername(username, ctx)
	if !valid {
		return false, errCode
	}

	valid, errCode = ValidatePassword(password, ctx)
	if !valid {
		return false, errCode
	}

	return true, 0
}

const (
	MaxProfilePictureSize = 5 * 1024 * 1024
)

var (
	AllowedProfilePictureExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	AllowedProfilePictureMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
)

func ValidateProfilePicture(fileHeader *multipart.FileHeader) ([]byte, string, bool, int) {
	if fileHeader.Size > MaxProfilePictureSize {
		return nil, "", false, ErrFileTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", false, ErrInvalidFileType
	}
	defer file.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", false, ErrInvalidFileType
	}

	// Check MIME type
	mimeType := http.DetectContentType(fileBytes)

	// Handle GIFs specially to preserve animation
	if mimeType == "image/gif" {
		// Validate it's a proper GIF
		gifReader := bytes.NewReader(fileBytes)
		_, err := gif.DecodeConfig(gifReader)
		if err != nil {
			return nil, "", false, ErrInvalidFileType
		}
		return fileBytes, "gif", true, 0
	}
	if mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/webp" {
		var ext string
		switch mimeType {
		case "image/jpeg":
			ext = "jpg"
		case "image/png":
			ext = "png"
		case "image/webp":
			ext = "webp"
		}
		return fileBytes, ext, true, 0
	}

	// Not a supported image type
	return nil, "", false, ErrInvalidFileType
}

func ValidateMessageContent(content string) (bool, int) {
	if content == "" {
		return false, ErrEmptyMessage
	}

	if len(content) > 2000 {
		return false, ErrMessageTooLong
	}

	return true, 0
}

func ValidateChannelAccess(userID, channelID string) (bool, int) {

	guildID, err := CacheFirstQuery(
		fmt.Sprintf("channel_guild:%s", channelID),
		5*time.Minute,
		func() (string, bool, error) {
			var guildID string
			found, err := QueryRow("GetGuildFromChannel", &guildID,
				"SELECT guild_id FROM channels WHERE channel_id = $1", channelID)
			return guildID, found, err
		})

	if err != nil {
		return false, ErrChannelNotFound
	}

	inGuild, err := IsUserInGuild(guildID, userID)

	if err != nil || !inGuild {
		return false, ErrUnauthorized
	}

	return true, 0
}

func ValidateGuildOwnership(userID, guildID string) (bool, int) {
	guild, found, err := GetGuildByID(guildID)
	if !found || err != nil {
		return false, ErrGuildNotFound
	}

	if guild["owner_id"] != userID {
		return false, ErrNotGuildOwner
	}

	return true, 0
}

func ValidateChannelPermissions(userID, channelID, guildID string, operation string) (string, bool, int) {
	if userID == "" {
		return "", false, ErrUnauthorized
	}

	if operation == "create" {
		if guildID == "" {
			return "", false, ErrUnauthorized
		}
		hasPermission, err := perms.Service.HasGuildPermission(userID, guildID, perms.CREATE_CHANNEL)
		if err != nil {
			return "", false, ErrDatabaseError
		}
		if !hasPermission {
			return "", false, ErrUnauthorized
		}
		return guildID, true, 0
	}

	if channelID == "" {
		return "", false, ErrChannelNotFound
	}

	return ValidateChannelPermissions(userID, channelID, "", operation)
}

func ValidateChannelOperation(userID, channelID, guildID, name, description, operation string) (string, bool, int) {
	if operation == "create" {
		if guildID == "" {
			return "", false, ErrUnauthorized
		}
		valid, errCode := ValidateChannelData(name, description)
		if !valid {
			return "", false, errCode
		}
		return ValidateChannelPermissions(userID, "", guildID, operation)
	}

	if channelID == "" {
		return "", false, ErrChannelNotFound
	}

	if operation == "edit" {
		valid, errCode := ValidateChannelData(name, description)
		if !valid {
			return "", false, errCode
		}
	}

	found, err := QueryRow("GetChannelGuild", &guildID,
		"SELECT guild_id FROM channels WHERE channel_id = $1", channelID)

	if !found || err != nil {
		return "", false, ErrChannelNotFound
	}

	return ValidateChannelPermissions(userID, "", guildID, "create")
}