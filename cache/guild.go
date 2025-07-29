package cache

import (
	"github.com/go-redis/redis"
)

func (c *RedisCache) AddUserToGuildOnline(guildID, userID, username string) error {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	
	c.RemoveUserFromGuildOffline(guildID, userID)
	
	return c.client.ZAdd(onlineKey, redis.Z{
		Score:  0,
		Member: userID,
	}).Err()
}

func (c *RedisCache) RemoveUserFromGuildOnline(guildID, userID string) error {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	return c.client.ZRem(onlineKey, userID).Err()
}

func (c *RedisCache) AddUserToGuildOffline(guildID, userID, username string) error {
	offlineKey := c.keys.GuildOfflineUsers(guildID)
	
	// Remove from online set first
	c.RemoveUserFromGuildOnline(guildID, userID)
	
	// Add to offline set with username as score for lexicographical ordering
	return c.client.ZAdd(offlineKey, redis.Z{
		Score:  0,
		Member: userID,
	}).Err()
}

func (c *RedisCache) RemoveUserFromGuildOffline(guildID, userID string) error {
	offlineKey := c.keys.GuildOfflineUsers(guildID)
	return c.client.ZRem(offlineKey, userID).Err()
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
		userIDs[i] = member
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
		userIDs[i] = member
	}
	
	return userIDs, nil
}

func (c *RedisCache) GetGuildOnlineCount(guildID string) (int, error) {
	onlineKey := c.keys.GuildOnlineUsers(guildID)
	count, err := c.client.ZCard(onlineKey).Result()
	return int(count), err
}