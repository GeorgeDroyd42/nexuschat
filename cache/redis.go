package cache

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type RedisCache struct {
	client     *redis.Client
	keys       KeyGenerator
	config     Config
	logger     func(level logrus.Level, module, operation, message string, err error)
	dbProvider *DBProvider
}

func NewRedisCache(client *redis.Client, keys KeyGenerator, logger func(level logrus.Level, module, operation, message string, err error)) *RedisCache {
	return &RedisCache{
		client: client,
		keys:   keys,
		config: DefaultConfig,
		logger: logger,
	}
}

func (c *RedisCache) WithDBProvider(dbProvider *DBProvider) *RedisCache {
	c.dbProvider = dbProvider
	return c
}

func (c *RedisCache) SetNX(key string, value interface{}, expiration time.Duration) bool {
	data, err := json.Marshal(value)
	if err != nil {
		return false
	}
	return c.client.SetNX(key, data, expiration).Val()
}

func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(key, data, expiration).Err()
}

func (c *RedisCache) Get(key string, dest interface{}) (bool, error) {
	data, err := c.client.Get(key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(data, dest)
	return true, err
}

func (c *RedisCache) Delete(key string) error {
	return c.client.Del(key).Err()
}

func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}
