package cache

import (
	"time"

	"github.com/go-redis/redis"
)

func (c *RedisCache) GetWebSocketConnections(userID string) ([]string, bool, error) {
	cacheKey := c.keys.WebSocket(userID)
	var connections []string
	found, err := c.Get(cacheKey, &connections)
	return connections, found, err
}

func (c *RedisCache) AddWebSocketConnection(userID, sessionID string, data map[string]string, expiration time.Duration) error {
	connKey := c.keys.WebSocketConnection(sessionID)
	err := c.Set(connKey, data, expiration)
	if err != nil {
		return err
	}

	cacheKey := c.keys.WebSocket(userID)
	var connections []string
	found, err := c.Get(cacheKey, &connections)
	if err != nil && err != redis.Nil {
		return err
	}

	if !found {
		connections = []string{}
	}

	for _, conn := range connections {
		if conn == sessionID {
			return nil
		}
	}

	connections = append(connections, sessionID)
	return c.Set(cacheKey, connections, expiration)
}

func (c *RedisCache) RemoveWebSocketConnection(userID, sessionID string) error {
	connKey := c.keys.WebSocketConnection(sessionID)
	err := c.Delete(connKey)
	if err != nil {
		return err
	}

	cacheKey := c.keys.WebSocket(userID)
	var connections []string
	found, err := c.Get(cacheKey, &connections)
	if err != nil && err != redis.Nil {
		return err
	}

	if !found {
		return nil
	}

	newConnections := []string{}
	for _, conn := range connections {
		if conn != sessionID {
			newConnections = append(newConnections, conn)
		}
	}

	return c.Set(cacheKey, newConnections, c.config.DefaultTTL)
}

func (c *RedisCache) GetWebSocketConnectionData(sessionID string) (map[string]string, bool, error) {
	connKey := c.keys.WebSocketConnection(sessionID)
	var data map[string]string
	found, err := c.Get(connKey, &data)
	return data, found, err
}

func (c *RedisCache) AddTypingUser(channelID, userID string, duration time.Duration) error {
	typingKey := c.keys.TypingIndicator(channelID)
	var typingUsers map[string]int64
	found, err := c.Get(typingKey, &typingUsers)
	if err != nil && err != redis.Nil {
		return err
	}
	if !found {
		typingUsers = make(map[string]int64)
	}
	typingUsers[userID] = time.Now().Unix()
	return c.Set(typingKey, typingUsers, duration)
}

func (c *RedisCache) RemoveTypingUser(channelID, userID string) error {
	typingKey := c.keys.TypingIndicator(channelID)
	var typingUsers map[string]int64
	found, err := c.Get(typingKey, &typingUsers)
	if err != nil && err != redis.Nil {
		return err
	}
	if !found {
		return nil
	}
	delete(typingUsers, userID)
	if len(typingUsers) == 0 {
		return c.Delete(typingKey)
	}
	return c.Set(typingKey, typingUsers, c.config.DefaultTTL)
}

func (c *RedisCache) GetTypingUsers(channelID string) ([]string, error) {
	typingKey := c.keys.TypingIndicator(channelID)
	var typingUsers map[string]int64
	found, err := c.Get(typingKey, &typingUsers)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if !found {
		return []string{}, nil
	}

	now := time.Now().Unix()
	activeUsers := []string{}
	expired := false

	for userID, timestamp := range typingUsers {
		if now-timestamp < 10 {
			activeUsers = append(activeUsers, userID)
		} else {
			delete(typingUsers, userID)
			expired = true
		}
	}

	if expired {
		if len(typingUsers) == 0 {
			c.Delete(typingKey)
		} else {
			c.Set(typingKey, typingUsers, c.config.DefaultTTL)
		}
	}

	return activeUsers, nil
}
