package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	
	"github.com/labstack/echo/v4"
)

func GenerateSessionID(entityType string) string {
	var prefix string
	switch entityType {
	case "user":
		prefix = "1"
	case "guild":
		prefix = "2"
	case "channel":
		prefix = "3"
	case "session":
		prefix = "4"
	case "message":
		prefix = "5"
	case "webhook":
		prefix = "6"
	default:
		prefix = "9"
	}

	max := 99999999999999 // 14 digits
	min := 10000000000000 // 14 digits

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()%int64(max-min+1)+int64(min))
	}

	sessionID := fmt.Sprintf("%s%d", prefix, n.Int64()+int64(min))

	return sessionID
}

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}


func CentralRefreshSession(c echo.Context, userID string) (time.Time, error) {
	return GlobalSessionManager.RefreshSession(c, userID)
}