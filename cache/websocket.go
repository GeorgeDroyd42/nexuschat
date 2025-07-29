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


