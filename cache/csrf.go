package cache

import (
	"time"
)

func (c *RedisCache) GetCSRFToken(sessionID string) (string, bool, error) {
	cacheKey := c.keys.CSRFToken(sessionID)
	var token string
	found, err := c.Get(cacheKey, &token)
	return token, found, err
}

func (c *RedisCache) SetCSRFToken(sessionID, token string, expiration time.Duration) error {
	cacheKey := c.keys.CSRFToken(sessionID)
	return c.Set(cacheKey, token, expiration)
}

func (c *RedisCache) DeleteCSRFToken(sessionID string) error {
	cacheKey := c.keys.CSRFToken(sessionID)
	return c.Delete(cacheKey)
}
