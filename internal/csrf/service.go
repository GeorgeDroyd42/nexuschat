package csrf

import (
	"time"
)

var Service = &csrfService{}

type csrfService struct {
	cache CacheProvider
	keys  KeyGenerator
}

type CacheProvider interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) (bool, error)
	Delete(key string) error
}

func Initialize(cache CacheProvider, keys KeyGenerator) {
	Service.cache = cache
	Service.keys = keys
}

func (s *csrfService) StoreToken(sessionID string, token string) {
	cacheKey := s.keys.CSRFToken(sessionID)
	s.cache.Set(cacheKey, token, 1*time.Hour)
}

func (s *csrfService) InvalidateToken(sessionID string) {
	cacheKey := s.keys.CSRFToken(sessionID)
	s.cache.Delete(cacheKey)
}

func (s *csrfService) GetToken(sessionID string) string {
	cacheKey := s.keys.CSRFToken(sessionID)
	var token string
	found, err := s.cache.Get(cacheKey, &token)
	if err != nil || !found {
		return ""
	}
	return token
}