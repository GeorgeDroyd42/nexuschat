package cache

import (
	"database/sql"
	"time"
)

type DBProvider struct {
	db     *sql.DB
	keys   KeyGenerator
	config Config
}

func NewDBProvider(db *sql.DB, keys KeyGenerator) *DBProvider {
	return &DBProvider{
		db:     db,
		keys:   keys,
		config: DefaultConfig,
	}
}

func (p *DBProvider) WithConfig(config Config) *DBProvider {
	p.config = config
	return p
}

func (p *DBProvider) GetSessionFromDB(sessionID string) (string, bool, error) {
	now := time.Now()
	var userID string
	var expiresAt time.Time

	rows, err := p.db.Query("SELECT user_id, expires_at FROM sessions WHERE session_id = $1", sessionID)
	if err != nil {
		return "", false, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&userID, &expiresAt)
		if err != nil {
			return "", false, err
		}

		if now.After(expiresAt) {
			return "", false, nil
		}

		return userID, true, nil
	}

	return "", false, nil
}

func (p *DBProvider) DeleteSessionFromDB(sessionID string) (bool, error) {
	result, err := p.db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (p *DBProvider) GetAdminFromDB(userID string) (bool, bool, error) {
	var isAdmin bool
	err := p.db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", userID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, false, nil // User doesn't exist
	}
	if err != nil {
		return false, false, err
	}
	return isAdmin, true, nil // isAdmin, found, error
}

func (p *DBProvider) GetUserFromDB(username string) (string, bool, error) {
	var password string
	err := p.db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&password)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return password, true, nil
}

func (p *DBProvider) GetUserBanFromDB(userID string) (bool, bool, error) {
	var isBanned bool
	err := p.db.QueryRow("SELECT is_banned FROM users WHERE user_id = $1", userID).Scan(&isBanned)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return isBanned, true, nil
}
