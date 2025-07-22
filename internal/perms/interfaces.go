package perms

import (
	"database/sql"
	"github.com/labstack/echo/v4"
)

// DatabaseProvider defines database operations needed by permission service
type DatabaseProvider interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// GuildProvider defines guild-related operations needed by permission service
type GuildProvider interface {
	GetGuildByID(guildID string) (map[string]interface{}, bool, error)
	RequireGuildMembership(c echo.Context, userID, guildID string) error
}

// UtilProvider defines utility operations needed by permission service
type UtilProvider interface {
	GenerateSessionID(prefix string) string
	SendErrorResponse(c echo.Context, errorCode int) error
}

