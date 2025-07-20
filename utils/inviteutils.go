package utils

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

func GenerateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	rand.Read(code)
	for i := range code {
		code[i] = charset[code[i]%byte(len(charset))]
	}
	return string(code)
}

func CreateInviteCode(guildID, createdBy string) (string, error) {
	var code string
	err := GetDB().QueryRow(`
		SELECT invite_code FROM guild_invites 
		WHERE guild_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1
	`, guildID).Scan(&code)

	if err == nil {
		return code, nil
	}

	for i := 0; i < 5; i++ {
		code = GenerateInviteCode()
		_, err = GetDB().Exec(`
			INSERT INTO guild_invites (invite_code, guild_id, created_by, uses_count, max_uses)
			VALUES ($1, $2, $3, 0, NULL)
		`, code, guildID, createdBy)

		if err == nil {
			return code, nil
		}
	}
	return "", err
}

func GetGuildByInviteCode(code string) (string, error) {
	var guildID string
	err := GetDB().QueryRow(`
		SELECT guild_id FROM guild_invites 
		WHERE invite_code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, strings.ToUpper(code)).Scan(&guildID)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("invalid invite code")
	}
	return guildID, err
}

func IsValidInviteCode(code string) bool {
	return len(code) == 8 && regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(strings.ToUpper(code))
}
