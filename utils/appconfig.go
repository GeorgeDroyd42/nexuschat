package utils

import (
	"os"
	"strconv"
	"strings"
	"time"
)

var AppConfig struct {
	SessionExpiryDuration time.Duration
	SessionCookieName     string
	SessionCookiePath     string
	SessionCookieSecure   bool
	SessionCookieHTTPOnly bool
	// Add these new Redis fields
	RedisHost               string
	RedisPort               string
	RedisPassword           string
	RedisDB                 int
	MaxWSConnectionsPerUser int
	MaxWSMessageSize        int
	MaxWSMessagesPerMinute  int
	AllowedOrigins          []string
	UsersPerPage            int
	MaxGuildsPerUser        int
	MembersPerPage          int 
	AllMembers              int
}

func InitAppConfig() {
	// Existing defaults
	AppConfig.SessionExpiryDuration = 24 * time.Hour
	AppConfig.SessionCookieName = "session"
	AppConfig.MaxWSMessageSize = 16384
	AppConfig.MaxGuildsPerUser = 300
	
	AppConfig.UsersPerPage = 50 // ADD THIS LINE
	AppConfig.MembersPerPage = 25  // ADD THIS LINE
	AppConfig.AllMembers = -1
	AppConfig.MaxWSMessagesPerMinute = 60
	AppConfig.SessionCookiePath = "/"
	AppConfig.SessionCookieSecure = true
	AppConfig.MaxWSConnectionsPerUser = 5
	AppConfig.SessionCookieHTTPOnly = true
	AppConfig.AllowedOrigins = []string{"http://localhost:8080", "https://nexuschat.loophole.site"}
	// Add Redis defaults
	AppConfig.RedisHost = "localhost"
	AppConfig.RedisPort = "6379"
	AppConfig.RedisPassword = ""
	AppConfig.RedisDB = 0

	// Existing env overrides
	if envDuration := os.Getenv("SESSION_EXPIRY_HOURS"); envDuration != "" {
		if hours, err := strconv.Atoi(envDuration); err == nil && hours > 0 {
			AppConfig.SessionExpiryDuration = time.Duration(hours) * time.Hour
		}
	}

	// Add Redis env overrides
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		AppConfig.RedisHost = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		AppConfig.RedisPort = redisPort
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		AppConfig.RedisPassword = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if db, err := strconv.Atoi(redisDB); err == nil {
			AppConfig.RedisDB = db
		}
	}

	// Add CORS env override
	if allowedOrigins := os.Getenv("ALLOWED_ORIGINS"); allowedOrigins != "" {
		for _, origin := range strings.Split(allowedOrigins, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				AppConfig.AllowedOrigins = append(AppConfig.AllowedOrigins, trimmed)
			}
		}
	}
}
