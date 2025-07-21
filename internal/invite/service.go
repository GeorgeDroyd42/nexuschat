package invite

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

var Service = &inviteService{}

type inviteService struct {
	db DBProvider
}

func Initialize(db DBProvider) {
	Service.db = db
}

func (s *inviteService) GenerateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	rand.Read(code)
	for i := range code {
		code[i] = charset[code[i]%byte(len(charset))]
	}
	return string(code)
}

func (s *inviteService) CreateInviteCode(guildID, createdBy string) (string, error) {
	var code string
	err := s.db.QueryRow(`
		SELECT invite_code FROM guild_invites 
		WHERE guild_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1
	`, guildID).Scan(&code)

	if err == nil {
		return code, nil
	}

	for i := 0; i < 5; i++ {
		code = s.GenerateInviteCode()
		_, err = s.db.Exec(`
			INSERT INTO guild_invites (invite_code, guild_id, created_by, uses_count, max_uses)
			VALUES ($1, $2, $3, 0, NULL)
		`, code, guildID, createdBy)

		if err == nil {
			return code, nil
		}
	}
	return "", err
}

func (s *inviteService) GetGuildByInviteCode(code string) (string, error) {
	var guildID string
	err := s.db.QueryRow(`
		SELECT guild_id FROM guild_invites 
		WHERE invite_code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, strings.ToUpper(code)).Scan(&guildID)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("invalid invite code")
	}
	return guildID, err
}

func (s *inviteService) IsValidInviteCode(code string) bool {
	return len(code) == 8 && regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(strings.ToUpper(code))
}