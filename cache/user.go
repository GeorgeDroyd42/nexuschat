package cache

import (
	"time"
)

func (c *RedisCache) GetUser(username string) (string, bool, error) {
	cacheKey := c.keys.User(username)
	var password string
	found, err := c.Get(cacheKey, &password)
	if found && err == nil {
		return password, true, nil
	}

	if c.dbProvider != nil {
		password, found, err := c.dbProvider.GetUserFromDB(username)
		if err != nil {
			return "", false, err
		}
		if found {
			c.Set(cacheKey, password, c.config.UserTTL)
			return password, true, nil
		}
	}

	return "", false, nil
}

func (c *RedisCache) SetUser(username, hashedPw string) error {
	cacheKey := c.keys.User(username)
	return c.Set(cacheKey, hashedPw, c.config.UserTTL)
}

func (c *RedisCache) DeleteUser(username string) error {
	cacheKey := c.keys.User(username)
	return c.Delete(cacheKey)
}

func (c *RedisCache) GetUserBan(userID string) (bool, bool, error) {
	cacheKey := c.keys.UserBan(userID)
	var isBanned bool
	found, err := c.Get(cacheKey, &isBanned)
	if found && err == nil {
		return isBanned, true, nil
	}

	if c.dbProvider != nil {
		isBanned, found, err := c.dbProvider.GetUserBanFromDB(userID)
		if err != nil {
			return false, false, err
		}
		if found {
			c.Set(cacheKey, isBanned, c.config.DefaultTTL)
			return isBanned, true, nil
		}
	}

	return false, false, nil
}

func (c *RedisCache) SetUserBan(userID string, isBanned bool, expiration time.Duration) error {
	cacheKey := c.keys.UserBan(userID)
	return c.Set(cacheKey, isBanned, expiration)
}

func (c *RedisCache) DeleteUserBan(userID string) error {
	cacheKey := c.keys.UserBan(userID)
	return c.Delete(cacheKey)
}
