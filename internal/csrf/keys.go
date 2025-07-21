package csrf

import "fmt"

type StandardKeyGenerator struct{}

func (kg StandardKeyGenerator) CSRFToken(sessionID string) string {
	return fmt.Sprintf("csrf:%s", sessionID)
}

var DefaultKeys = StandardKeyGenerator{}