package repository

import (
	"database/sql"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) session.ISessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Upsert(user *session.WhatsappUser) error {
	query := `
		INSERT INTO whatsapp_user (user_id, agent_id, agent_name, api_key, endpoint_url_run, status, last_connected_at, last_disconnected_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, agent_id) DO UPDATE SET
			agent_name = EXCLUDED.agent_name,
			api_key = EXCLUDED.api_key,
			endpoint_url_run = EXCLUDED.endpoint_url_run,
			status = EXCLUDED.status,
			last_connected_at = EXCLUDED.last_connected_at,
			last_disconnected_at = EXCLUDED.last_disconnected_at,
			updated_at = EXCLUDED.updated_at
	`
	// Note: The query above uses Postgres syntax ($1, $2). For SQLite it might need ? or logic to switch.
	// However, standard sql package in Go often supports ? for both if using a wrapper, but here we are using raw drivers.
	// SQLite supports $1 as well.

	_, err := r.db.Exec(query,
		user.UserID,
		user.AgentID,
		user.AgentName,
		user.ApiKey,
		user.EndpointUrlRun,
		user.Status,
		user.LastConnectedAt,
		user.LastDisconnectedAt,
		time.Now(),
	)
	return err
}

func (r *SessionRepository) FindOne(userID, agentID string) (*session.WhatsappUser, error) {
	query := `SELECT user_id, agent_id, agent_name, api_key, endpoint_url_run, status, last_connected_at, last_disconnected_at, created_at, updated_at FROM whatsapp_user WHERE user_id = $1 AND agent_id = $2`
	row := r.db.QueryRow(query, userID, agentID)

	var user session.WhatsappUser
	err := row.Scan(
		&user.UserID,
		&user.AgentID,
		&user.AgentName,
		&user.ApiKey,
		&user.EndpointUrlRun,
		&user.Status,
		&user.LastConnectedAt,
		&user.LastDisconnectedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SessionRepository) Delete(userID, agentID string) error {
	query := `DELETE FROM whatsapp_user WHERE user_id = $1 AND agent_id = $2`
	_, err := r.db.Exec(query, userID, agentID)
	return err
}
