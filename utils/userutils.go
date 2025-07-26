package utils

import (
	"fmt"
	"time"
	"auth.com/v4/internal/websockets"
	"auth.com/v4/cache"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 9)
	return string(bytes), err
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func IsUserBanned(userID string) (bool, error) {
	isBanned, err := cache.Service.CheckUserBan(userID)
	if err != nil {
		return false, err
	}
	found := true
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	if isBanned {
		var dbBanned bool
		dbFound, dbErr := QueryRow("VerifyUserBanStatus", &dbBanned,
			"SELECT is_banned FROM users WHERE user_id = $1", userID)
		if dbErr == nil && dbFound {
			return dbBanned, nil
		}
	}

	return isBanned, nil
}

func GetUsernameByID(userID string) (string, error) {
	username, err := CacheFirstQuery(
		fmt.Sprintf("username:%s", userID),
		cache.DefaultConfig.DefaultTTL,
		func() (string, bool, error) {
			var username string
			found, err := QueryRow("GetUsernameByID", &username, "SELECT username FROM users WHERE user_id = $1", userID)
			return username, found, err
		})
	return username, err
}

func SetUserBanStatus(userID string, isBanned bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE users SET is_banned = $1 WHERE user_id = $2", isBanned, userID)
	if err != nil {
		return err
	}

	cache.Provider.SetUserBan(userID, isBanned, cache.DefaultConfig.DefaultTTL)
	err = tx.Commit()
	if err != nil {
		return err
	}

	if isBanned {
		websockets.SendEventToUser(userID, "user_banned", ErrorMessages[ErrAccountSuspended])
		websockets.CleanupUserWebSocketConnections(userID)
	}

	rows, _ := db.Query("SELECT token FROM sessions WHERE user_id = $1", userID)
	defer rows.Close()
	for rows.Next() {
		var token string
		rows.Scan(&token)
		sessionData, found, _ := cache.Provider.GetSessionWithUser(token)
		if found {
			sessionData.IsBanned = isBanned
			cache.Provider.SetSessionWithUser(token, sessionData, time.Hour)
		}
	}

	return nil
}

func CacheFirstQuery[T any](cacheKey string, ttl time.Duration, dbQuery func() (T, bool, error)) (T, error) {
	var result T
	found, err := cache.Provider.Get(cacheKey, &result)
	if found && err == nil {
		return result, nil
	}

	lockKey := cacheKey + ":lock"
	if cache.Provider.SetNX(lockKey, "1", 10*time.Second) {
		defer cache.Provider.Delete(lockKey)

		result, found, err = dbQuery()
		if found && err == nil {
			cache.Provider.Set(cacheKey, result, ttl)
		}
	} else {
		time.Sleep(50 * time.Millisecond)
		return CacheFirstQuery(cacheKey, ttl, dbQuery)
	}

	return result, err
}
