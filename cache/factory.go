package cache

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Factory struct {
	redisOptions *redis.Options
	config       Config
	keys         KeyGenerator
	logger       func(level logrus.Level, module, operation, message string, err error)
}

func NewFactory() *Factory {
	return &Factory{
		config: DefaultConfig,
		keys:   DefaultKeys,
	}
}

func (f *Factory) WithRedisOptions(options *redis.Options) *Factory {
	f.redisOptions = options
	return f
}

func (f *Factory) WithConfig(config Config) *Factory {
	f.config = config
	return f
}

func (f *Factory) WithKeys(keys KeyGenerator) *Factory {
	f.keys = keys
	return f
}

func (f *Factory) WithLogger(logger func(level logrus.Level, module, operation, message string, err error)) *Factory {
	f.logger = logger
	return f
}

func (f *Factory) CreateRedisCache() (*RedisCache, error) {
	client := redis.NewClient(f.redisOptions)

	_, err := client.Ping().Result()
	if err != nil {
		if f.logger != nil {
			f.logger(logrus.ErrorLevel, "cache", "redis_init", "", err)
		}
		return nil, err
	}

	if f.logger != nil {
		f.logger(logrus.InfoLevel, "cache", "redis_init", "Redis connection established", nil)
	}

	return NewRedisCache(client, f.keys, f.logger), nil
}
