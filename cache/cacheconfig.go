package cache

import (
	"time"
)

type Config struct {
	UserTTL         time.Duration
	AdminTTL        time.Duration
	SessionTTL      time.Duration
	UserSessionsTTL time.Duration
	DefaultTTL      time.Duration
}

var DefaultConfig = Config{
	UserTTL:         5 * time.Minute,
	AdminTTL:        5 * time.Minute,
	SessionTTL:      5 * time.Minute,
	UserSessionsTTL: 5 * time.Minute,
	DefaultTTL:      5 * time.Minute,
}
