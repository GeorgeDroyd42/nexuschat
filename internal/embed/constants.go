package embed

import "time"

const (
	CacheKeyPrefix = "embed:"
	UserAgent      = "NexusChat"
	MaxRedirects   = 3
	RequestTimeout = 5 * time.Second
	CacheDuration  = 1 * time.Hour
)