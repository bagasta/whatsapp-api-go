package database

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

func Migrate(db *sql.DB, driverName string) error {
	logrus.Info("Running database migrations...")

	var queries []string

	// Postgres specific syntax
	if driverName == "postgres" {
		queries = []string{
			`CREATE TABLE IF NOT EXISTS whatsapp_user (
				user_id VARCHAR(255) NOT NULL,
				agent_id VARCHAR(255) NOT NULL,
				agent_name VARCHAR(255),
				api_key VARCHAR(255),
				endpoint_url_run VARCHAR(255),
				status VARCHAR(50),
				last_connected_at TIMESTAMP,
				last_disconnected_at TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (user_id, agent_id)
			);`,
			`CREATE TABLE IF NOT EXISTS api_keys (
				user_id VARCHAR(255) NOT NULL,
				access_token VARCHAR(255) NOT NULL,
				is_active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS webhook_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				url TEXT NOT NULL,
				secret TEXT,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS agent_webhook_config (
				agent_id VARCHAR(255) PRIMARY KEY,
				url VARCHAR(255) NOT NULL,
				secret VARCHAR(255),
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS ai_message_logs (
				id SERIAL PRIMARY KEY,
				agent_id VARCHAR(255) NOT NULL,
				message_id VARCHAR(255),
				user_id VARCHAR(255),
				status VARCHAR(50),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
		}
	} else {
		// SQLite syntax (TIMESTAMP -> DATETIME, etc if needed, but simplified here)
		queries = []string{
			`CREATE TABLE IF NOT EXISTS whatsapp_user (
				user_id TEXT NOT NULL,
				agent_id TEXT NOT NULL,
				agent_name TEXT,
				api_key TEXT,
				endpoint_url_run TEXT,
				status TEXT,
				last_connected_at DATETIME,
				last_disconnected_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (user_id, agent_id)
			);`,
			`CREATE TABLE IF NOT EXISTS api_keys (
				user_id TEXT NOT NULL,
				access_token TEXT NOT NULL,
				is_active BOOLEAN DEFAULT 1,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS webhook_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				url TEXT NOT NULL,
				secret TEXT,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS agent_webhook_config (
				agent_id TEXT PRIMARY KEY,
				url TEXT NOT NULL,
				secret TEXT,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS ai_message_logs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				agent_id TEXT NOT NULL,
				message_id TEXT,
				user_id TEXT,
				status TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
		}
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute migration query: %s, error: %w", query, err)
		}
	}

	logrus.Info("Database migrations completed successfully.")
	return nil
}
