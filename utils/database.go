package utils

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
	"errors"
	"auth.com/v4/cache"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type UserRegistration struct {
	Username string
	Password string
	Result   chan error
}

var envLoaded bool = false

func LoadEnv() {
	if !envLoaded {
		godotenv.Load()
		envLoaded = true
	}
}

func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func GetEnv(key string, defaultVal interface{}) interface{} {
	LoadEnv()
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	if _, ok := defaultVal.(int); ok {
		intVal, err := strconv.Atoi(val)
		if err != nil || intVal <= 0 {
			Log.Error("database", "config_parse", "Invalid config value", nil, map[string]interface{}{
				"key": key,
				"value": val,
			})
			return defaultVal
		}
		return intVal
	}
	return val
}

func executeDBOperation(operation string, fn func() error) error {
	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	if err != nil {
		UnifiedDBErrorHandler(operation, err, nil, nil)
	} else if duration > 100*time.Millisecond {
		Log.LogWithFields(LogWarn, "database", operation, "Slow operation detected", nil, map[string]interface{}{"duration_ms": duration.Milliseconds()})
	} else {
		Log.LogWithFields(LogDebug, "database", operation, "Operation executed", nil, map[string]interface{}{"duration_ms": duration.Milliseconds()})
	}

	return err
}

func SetUserAdminStatus(userID string, isAdmin bool) error {
	err := ExecuteQuery("SetUserAdminStatus",
		"UPDATE users SET is_admin = $1 WHERE user_id = $2",
		isAdmin, userID)

	if err == nil {
		var username string
		found, _ := QueryRow("GetUsernameForCache", &username,
			"SELECT username FROM users WHERE user_id = $1", userID)
		if found {
			cache.Provider.DeleteUser(username)
		}
		cache.Provider.DeleteAdmin(userID)

		updateUserSessionAdminStatus(userID, isAdmin)
	}

	return err
}



func GetAllUsers(page, limit int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * limit

	users := []map[string]interface{}{}
	rows, err := db.Query("SELECT user_id, username, is_admin, is_banned, created_at, profile_picture FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		UnifiedDBErrorHandler("GetAllUsers", err, nil, nil)
		return users, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var username string
		var isAdmin bool
		var isBanned bool
		var createdAt time.Time
		var profilePicture sql.NullString
		if err := rows.Scan(&userID, &username, &isAdmin, &isBanned, &createdAt, &profilePicture); err != nil {
			UnifiedDBErrorHandler("GetAllUsers", err, nil, nil)
			continue
		}

		profilePicturePath := ""
		if profilePicture.Valid {
			profilePicturePath = profilePicture.String
		}

		users = append(users, map[string]interface{}{
			"user_id":         userID,
			"username":        username,
			"is_admin":        isAdmin,
			"is_banned":       isBanned,
			"created_at":      formatTimestamp(createdAt),
			"profile_picture": profilePicturePath,
		})
	}

	if err = rows.Err(); err != nil {
		UnifiedDBErrorHandler("GetAllUsers", err, nil, nil)
		return users, 0, err
	}

	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		UnifiedDBErrorHandler("GetAllUsers", err, nil, nil)
		return users, 0, err
	}

	return users, totalCount, nil
}

func configureDBConnection(db *sql.DB) {
	maxOpenConns := GetEnv("DB_MAX_OPEN_CONNS", 25).(int)
	if maxOpenConns <= 0 {
		Log.LogWithFields(LogWarn, "database", "config", "Invalid DB_MAX_OPEN_CONNS value, using default", nil, nil)
		maxOpenConns = 25
	}
	db.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := GetEnv("DB_MAX_IDLE_CONNS", 5).(int)
	if maxIdleConns <= 0 || maxIdleConns > maxOpenConns {
		Log.LogWithFields(LogWarn, "database", "config", "Invalid DB_MAX_IDLE_CONNS value, using default", nil, nil)
		maxIdleConns = 5
	}
	db.SetMaxIdleConns(maxIdleConns)

	connMaxLifetime := GetEnv("DB_CONN_MAX_LIFETIME", 30).(int)
	if connMaxLifetime <= 0 {
		Log.LogWithFields(LogWarn, "database", "config", "Invalid DB_CONN_MAX_LIFETIME value, using default", nil, nil)
		connMaxLifetime = 30
	}
	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)
}

type DBError struct {
	Operation string
	Err       error
	Code      int
}

func HandleDBError(operation string, err error) (bool, int) {
	return UnifiedDBErrorHandler(operation, err, nil, nil)
}

func ExecuteQuery(operation string, query string, args ...interface{}) error {
	return executeDBOperation(operation, func() error {
		_, err := db.Exec(query, args...)
		return err
	})
}

var db *sql.DB

type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

type ColumnSchema struct {
	Name       string
	Type       string
	Nullable   bool
	Default    string
	PrimaryKey bool
}

var DatabaseSchema = []TableSchema{
	{
		Name: "users",
		Columns: []ColumnSchema{
			{Name: "user_id", Type: "VARCHAR(20)", Nullable: false, PrimaryKey: true},
			{Name: "username", Type: "VARCHAR(50)", Nullable: false},
			{Name: "password", Type: "VARCHAR(100)", Nullable: false},
			{Name: "is_admin", Type: "BOOLEAN", Nullable: true, Default: "FALSE"},
			{Name: "is_banned", Type: "BOOLEAN", Nullable: true, Default: "FALSE"},
			{Name: "profile_picture", Type: "VARCHAR(255)", Nullable: true},
			{Name: "created_at", Type: "TIMESTAMP", Nullable: true, Default: "CURRENT_TIMESTAMP"},
		},
	},

	{
		Name: "sessions",
		Columns: []ColumnSchema{
			{Name: "token", Type: "VARCHAR(100)", Nullable: false},
			{Name: "user_id", Type: "VARCHAR(20)", Nullable: false},
			{Name: "session_id", Type: "VARCHAR(100)", Nullable: true},
			{Name: "created_at", Type: "TIMESTAMP", Nullable: true, Default: "CURRENT_TIMESTAMP"},
		},
	},
}

func InitDB() error {
	host := GetEnv("DB_HOST", "localhost").(string)
	port := GetEnv("DB_PORT", "5432").(string)
	user := GetEnv("DB_USER", "postgres").(string)
	password := GetEnv("DB_PASSWORD", "").(string)
	dbname := GetEnv("DB_NAME", "postgres").(string)
	sslmode := GetEnv("DB_SSLMODE", "disable").(string)

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		UnifiedDBErrorHandler("open_connection", err, nil, nil)
		Log.Error("database", "open_connection", "Unable to open database connection", err)
return errors.New("unable to open database connection")
	}

	configureDBConnection(db)

	err = db.Ping()
	if err != nil {
		UnifiedDBErrorHandler("ping_database", err, nil, nil)
		Log.Error("database", "ping_database", "Database connection failed", err)
return errors.New("database connection failed")
	}

	Log.Info("database", "init_database", "Database connection established successfully")

	err = RunMigrations()
	if err != nil {
		UnifiedDBErrorHandler("migrate_database", err, nil, nil)
		return err
	}

	return nil
}

func QueryRow(operation string, dest interface{}, query string, args ...interface{}) (bool, error) {
	var err error

	executeDBOperation(operation, func() error {
		err = db.QueryRow(query, args...).Scan(dest)
		return err
	})

	if err == sql.ErrNoRows {
		return false, nil
	}

	return err == nil, err
}

func CreateUser(user_id, username, hashedPassword string) error {
	return ExecuteQuery("CreateUser", "INSERT INTO users (user_id, username, password) VALUES ($1, $2, $3)", user_id, username, hashedPassword)
}

func WithTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := GetDB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func UserExists(username string) (bool, error) {
	var count int
	_, err := QueryRow("UserExists", &count, "SELECT COUNT(*) FROM users WHERE username = $1", username)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UnifiedDBErrorHandler(operation string, err error, tx *sql.Tx, batch []*UserRegistration) (bool, int) {
	if err == nil {
		return true, 0
	}

	Log.Error("database", operation, "Database operation failed", err)

	if tx != nil {
		tx.Rollback()
	}

	for _, reg := range batch {
		reg.Result <- err
		close(reg.Result)
	}

	if err == sql.ErrNoRows {
		return false, ErrInvalidCredentials
	}

	return false, ErrDatabaseError
}

func GetDB() *sql.DB {
	return db
}