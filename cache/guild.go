package cache

import (
	"github.com/go-redis/redis"
)

func (c *RedisCache) AddUserToGuildOnline(guildID, userID, username string) error {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	
	c.RemoveUserFromGuildOffline(guildID, userID)
	
	return c.client.ZAdd(onlineKey, redis.Z{
		Score:  0,
		Member: userID + ":" + username,
	}).Err()
}

func (c *RedisCache) RemoveUserFromGuildOnline(guildID, userID string) error {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	
	// Remove all members that start with userID:
	members, err := c.client.ZRange(onlineKey, 0, -1).Result()
	if err != nil {
		return err
	}
	
	for _, member := range members {
		if len(member) > len(userID)+1 && member[:len(userID)+1] == userID+":" {
			return c.client.ZRem(onlineKey, member).Err()
		}
	}
	return nil
}

func (c *RedisCache) AddUserToGuildOffline(guildID, userID, username string) error {
	offlineKey := c.keys.GuildOfflineUsers(guildID)
	
	// Remove from online set first
	c.RemoveUserFromGuildOnline(guildID, userID)
	
	// Add to offline set with username as score for lexicographical ordering
	return c.client.ZAdd(offlineKey, redis.Z{
		Score:  0,
		Member: userID + ":" + username,
	}).Err()
}

func (c *RedisCache) RemoveUserFromGuildOffline(guildID, userID string) error {
	offlineKey := c.keys.GuildOfflineUsers(guildID)
	
	// Remove all members that start with userID:
	members, err := c.client.ZRange(offlineKey, 0, -1).Result()
	if err != nil {
		return err
	}
	
	for _, member := range members {
		if len(member) > len(userID)+1 && member[:len(userID)+1] == userID+":" {
			return c.client.ZRem(offlineKey, member).Err()
		}
	}
	return nil
}

func (c *RedisCache) GetGuildOnlineUsers(guildID string, offset, limit int) ([]string, error) {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	
	members, err := c.client.ZRangeByLex(onlineKey, redis.ZRangeBy{
		Min:    "-",
		Max:    "+",
		Offset: int64(offset),
		Count:  int64(limit),
	}).Result()
	
	if err != nil {
		return nil, err
	}
	
	userIDs := make([]string, len(members))
	for i, member := range members {
		// Extract userID from "userID:username" format
		colonIndex := -1
		for j, char := range member {
			if char == ':' {
				colonIndex = j
				break
			}
		}
		if colonIndex > 0 {
			userIDs[i] = member[:colonIndex]
		}
	}
	
	return userIDs, nil
}

func (c *RedisCache) GetGuildOfflineUsers(guildID string, offset, limit int) ([]string, error) {
	offlineKey := c.keys.GuildOfflineUsers(guildID)
	
	members, err := c.client.ZRangeByLex(offlineKey, redis.ZRangeBy{
		Min:    "-",
		Max:    "+",
		Offset: int64(offset),
		Count:  int64(limit),
	}).Result()
	
	if err != nil {
		return nil, err
	}
	
	userIDs := make([]string, len(members))
	for i, member := range members {
		// Extract userID from "userID:username" format
		colonIndex := -1
		for j, char := range member {
			if char == ':' {
				colonIndex = j
				break
			}
		}
		if colonIndex > 0 {
			userIDs[i] = member[:colonIndex]
		}
	}
	
	return userIDs, nil
}

func (c *RedisCache) GetGuildOnlineCount(guildID string) (int, error) {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	count, err := c.client.ZCard(onlineKey).Result()
	return int(count), err
}