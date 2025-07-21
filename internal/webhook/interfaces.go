package webhook

import (
	"database/sql"
)

type DBProvider interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type QueryHelper interface {
	ExecuteQuery(operation string, query string, args ...interface{}) error
}

type SessionGenerator interface {
	GenerateSessionID(entityType string) string
	GenerateSessionToken() (string, error)
}