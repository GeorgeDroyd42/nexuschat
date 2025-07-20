package cache

import (
	"fmt"
	"time"
)

func (c *RedisCache) GetSession(sessionID string) (string, bool, error) {
	cacheKey := c.keys.Session(sessionID)
	var userID string
	found, err := c.Get(cacheKey, &userID)
	return userID, found, err
}

func (c *RedisCache) GetSessionFromDB(sessionID string) (string, bool, error) {
	if c.dbProvider == nil {
		return "", false, fmt.Errorf("db provider not configured")
	}

	userID, found, err := c.dbProvider.GetSessionFromDB(sessionID)
	if err != nil {
		return "", false, err
	}

	if found {
		cacheKey := c.keys.Session(sessionID)
		c.Set(cacheKey, userID, c.config.SessionTTL)
	}

	return userID, found, nil
}

func (c *RedisCache) SetSession(sessionID, userID string, expiration time.Duration) error {
	cacheKey := c.keys.Session(sessionID)
	return c.Set(cacheKey, userID, expiration)
}

func (c *RedisCache) DeleteSession(sessionID string) (bool, error) {
	cacheKey := c.keys.Session(sessionID)
	err := c.client.Del(cacheKey).Err()
	if err != nil {
		return false, err
	}

	if c.dbProvider != nil {
		return c.dbProvider.DeleteSessionFromDB(sessionID)
	}

	return true, nil
}

func (c *RedisCache) GetUserSessions(userID string) ([]string, bool, error) {
	cacheKey := c.keys.UserSessions(userID)
	var sessionIDs []string
	found, err := c.Get(cacheKey, &sessionIDs)
	return sessionIDs, found, err
}

func (c *RedisCache) SetUserSessions(userID string, sessionIDs []string, expiration time.Duration) error {
	cacheKey := c.keys.UserSessions(userID)
	return c.Set(cacheKey, sessionIDs, expiration)
}

func (c *RedisCache) GetSessionWithUser(token string) (*SessionData, bool, error) {
	key := c.keys.SessionData(token)

	var sessionData SessionData
	found, err := c.Get(key, &sessionData)
	if !found || err != nil {
		return nil, false, err
	}

	if time.Now().After(sessionData.ExpiresAt) {
		c.Delete(key)
		return nil, false, nil
	}

	return &sessionData, true, nil
}

func (c *RedisCache) SetSessionWithUser(token string, data *SessionData, ttl time.Duration) error {
	key := c.keys.SessionData(token)
	return c.Set(key, data, ttl)
}

func (r *RedisCache) DeleteSessionToken(token string) error {
	key := r.keys.SessionData(token)
	return r.Delete(key)
}
