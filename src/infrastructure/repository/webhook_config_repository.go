package repository

import (
	"database/sql"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
)

type WebhookConfigRepository struct {
	db *sql.DB
}

func NewWebhookConfigRepository(db *sql.DB) webhook.IWebhookConfigRepository {
	return &WebhookConfigRepository{db: db}
}

func (r *WebhookConfigRepository) GetDefault() (*webhook.Config, error) {
	row := r.db.QueryRow(`SELECT url, secret, updated_at FROM webhook_config WHERE id = 1`)

	var cfg webhook.Config
	if err := row.Scan(&cfg.URL, &cfg.Secret, &cfg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *WebhookConfigRepository) UpsertDefault(cfg *webhook.Config) error {
	query := `
		INSERT INTO webhook_config (id, url, secret, updated_at)
		VALUES (1, $1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET
			url = EXCLUDED.url,
			secret = EXCLUDED.secret,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, cfg.URL, cfg.Secret, cfg.UpdatedAt)
	return err
}

func (r *WebhookConfigRepository) GetByAgent(agentID string) (*webhook.Config, error) {
	row := r.db.QueryRow(`SELECT agent_id, url, secret, updated_at FROM agent_webhook_config WHERE agent_id = $1`, agentID)

	var cfg webhook.Config
	if err := row.Scan(&cfg.AgentID, &cfg.URL, &cfg.Secret, &cfg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *WebhookConfigRepository) UpsertByAgent(cfg *webhook.Config) error {
	query := `
		INSERT INTO agent_webhook_config (agent_id, url, secret, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(agent_id) DO UPDATE SET
			url = EXCLUDED.url,
			secret = EXCLUDED.secret,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, cfg.AgentID, cfg.URL, cfg.Secret, cfg.UpdatedAt)
	return err
}
