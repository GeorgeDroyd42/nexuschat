package cache

import (
	"database/sql"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	Provider CacheProvider
	Keys     KeyGenerator = DefaultKeys
)

func Initialize(redisOptions *redis.Options, db *sql.DB, logger func(level logrus.Level, module, operation, message string, err error)) error {
	factory := NewFactory().
		WithRedisOptions(redisOptions).
		WithLogger(logger)

	redisCache, err := factory.CreateRedisCache()
	if err != nil {
		return err
	}

	dbProvider := NewDBProvider(db, Keys)
	redisCache.WithDBProvider(dbProvider)

	Provider = redisCache
	return nil
}
