package utils

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Name string
	Up   func(*sql.DB) error
}

var migrations []Migration

func RegisterMigration(migration Migration) {
	migrations = append(migrations, migration)
}

func RunMigrations() error {
	for _, migration := range migrations {
		if err := migration.Up(db); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	RegisterMigration(Migration{
		Name: "create_users_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS users (
					user_id VARCHAR(20) PRIMARY KEY,
					username VARCHAR(50) NOT NULL UNIQUE,
					password VARCHAR(100) NOT NULL,
					is_admin BOOLEAN DEFAULT FALSE,
					is_banned BOOLEAN DEFAULT FALSE,
					profile_picture VARCHAR(255),
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`)
			return err
		},
	})

	RegisterMigration(Migration{
		Name: "create_sessions_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS sessions (
					token VARCHAR(100) NOT NULL,
					user_id VARCHAR(20) NOT NULL,
					session_id VARCHAR(100),
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP
				)
			`)
			return err
		},
	})

	RegisterMigration(Migration{
		Name: "create_guilds_tables",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
            CREATE TABLE IF NOT EXISTS guilds (
                guild_id VARCHAR(20) PRIMARY KEY,
                name VARCHAR(100) NOT NULL,
                description TEXT,
                owner_id VARCHAR(20) NOT NULL REFERENCES users(user_id),
                created_at TIMESTAMP DEFAULT NOW()
            )
        `)
			if err != nil {
				return err
			}

			_, err = db.Exec(`
            CREATE TABLE IF NOT EXISTS guild_members (
                guild_id VARCHAR(20) REFERENCES guilds(guild_id) ON DELETE CASCADE,
                user_id VARCHAR(20) REFERENCES users(user_id) ON DELETE CASCADE,
                joined_at TIMESTAMP DEFAULT NOW(),
                PRIMARY KEY (guild_id, user_id)
            )
        `)
			return err
		},
	})

	RegisterMigration(Migration{
		Name: "create_channels_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS channels (
				channel_id VARCHAR(255) PRIMARY KEY,
				guild_id VARCHAR(255) NOT NULL,
				name VARCHAR(50) NOT NULL,
				description VARCHAR(500),
				created_by VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT NOW(),
				FOREIGN KEY (guild_id) REFERENCES guilds(guild_id) ON DELETE CASCADE
			);
				CREATE INDEX IF NOT EXISTS idx_channels_guild_id ON channels(guild_id);
			`)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "create_messages_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
            CREATE TABLE IF NOT EXISTS messages (
                message_id VARCHAR(20) PRIMARY KEY,
                channel_id VARCHAR(255) NOT NULL REFERENCES channels(channel_id) ON DELETE CASCADE,
                user_id VARCHAR(20) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
                content TEXT NOT NULL,
                message_type VARCHAR(20) DEFAULT 'text',
                created_at TIMESTAMP DEFAULT NOW(),
                updated_at TIMESTAMP DEFAULT NOW(),
                edited BOOLEAN DEFAULT FALSE
            );
            CREATE INDEX IF NOT EXISTS idx_messages_channel_time ON messages(channel_id, created_at DESC);
            CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id);
        `)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "add_guild_profile_picture",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
				ALTER TABLE guilds ADD COLUMN IF NOT EXISTS profile_picture_url VARCHAR(255)
			`)
			return err
		},
	})

	RegisterMigration(Migration{
		Name: "add_user_bio",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
			ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
		`)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "create_webhooks_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
                                CREATE TABLE IF NOT EXISTS webhooks (
                                        webhook_id VARCHAR(255) PRIMARY KEY,
                                        channel_id VARCHAR(255) NOT NULL REFERENCES channels(channel_id) ON DELETE CASCADE,
                                        name VARCHAR(100) NOT NULL,
                                        token VARCHAR(255) NOT NULL UNIQUE,
                                        created_by VARCHAR(255) NOT NULL REFERENCES users(user_id),
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        is_active BOOLEAN DEFAULT TRUE,
                                        last_used TIMESTAMP,
                                        use_count INTEGER DEFAULT 0
                                );
                                CREATE INDEX IF NOT EXISTS idx_webhooks_channel_id ON webhooks(channel_id);
                                CREATE INDEX IF NOT EXISTS idx_webhooks_token ON webhooks(token);
                        `)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "add_webhook_profile_picture",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
			ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS profile_picture VARCHAR(255);
		`)
			return err
		},
	})


	RegisterMigration(Migration{
		Name: "modify_messages_for_webhooks",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
                ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_user_id_fkey;
                ALTER TABLE messages ALTER COLUMN user_id DROP NOT NULL;
                ALTER TABLE messages ADD COLUMN IF NOT EXISTS webhook_id VARCHAR(255);
                ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_source VARCHAR(20) DEFAULT 'user';
                ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_username VARCHAR(100);
                ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_profile_picture VARCHAR(255);
                CREATE INDEX IF NOT EXISTS idx_messages_webhook ON messages(webhook_id);
        `)
			return err
		},
	})

	RegisterMigration(Migration{
		Name: "add_guild_member_counts",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
ALTER TABLE guilds ADD COLUMN IF NOT EXISTS member_count INTEGER DEFAULT 0;
UPDATE guilds SET member_count = (
	SELECT COUNT(*) FROM guild_members gm WHERE gm.guild_id = guilds.guild_id
);
`)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "create_guild_roles_system",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
            CREATE TABLE IF NOT EXISTS guild_roles (
                role_id VARCHAR(20) PRIMARY KEY,
                guild_id VARCHAR(20) NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
                name VARCHAR(50) NOT NULL,
                permissions TEXT[] DEFAULT '{}',
                color VARCHAR(7) DEFAULT '#99AAB5',
                position INTEGER DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
            CREATE INDEX IF NOT EXISTS idx_guild_roles_guild_id ON guild_roles(guild_id);
            
            CREATE TABLE IF NOT EXISTS guild_member_roles (
                guild_id VARCHAR(20) NOT NULL,
                user_id VARCHAR(20) NOT NULL,
                role_id VARCHAR(20) NOT NULL REFERENCES guild_roles(role_id) ON DELETE CASCADE,
                assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (guild_id, user_id, role_id)
            );
            CREATE INDEX IF NOT EXISTS idx_guild_member_roles_user ON guild_member_roles(user_id);
        `)
			return err
		},
	})
	RegisterMigration(Migration{
		Name: "create_guild_invites_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
            CREATE TABLE IF NOT EXISTS guild_invites (
                invite_code VARCHAR(8) PRIMARY KEY,
                guild_id VARCHAR(20) NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
                created_by VARCHAR(20) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                expires_at TIMESTAMP NULL,
                uses_count INTEGER DEFAULT 0,
                max_uses INTEGER NULL
            );
            CREATE INDEX IF NOT EXISTS idx_guild_invites_guild_id ON guild_invites(guild_id);
            CREATE INDEX IF NOT EXISTS idx_guild_invites_expires ON guild_invites(expires_at);
        `)
			return err
		},
	})

}

func AddIfNotExists(columnName string) error {
	var targetTable string
	var columnDef ColumnSchema
	found := false

	for _, table := range DatabaseSchema {
		for _, column := range table.Columns {
			if column.Name == columnName {
				targetTable = table.Name
				columnDef = column
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("column %s not found in schema", columnName)
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name=$1 AND column_name=$2)", targetTable, columnName).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", targetTable, columnName, columnDef.Type)
		if columnDef.Default != "" {
			query += " DEFAULT " + columnDef.Default
		}
		_, err = db.Exec(query)
	}

	return err
}
