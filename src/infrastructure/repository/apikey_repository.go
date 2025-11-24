package repository

import (
	"database/sql"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/apikey"
)

type ApiKeyRepository struct {
	db *sql.DB
}

func NewApiKeyRepository(db *sql.DB) apikey.IApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

func (r *ApiKeyRepository) FindActive(userID string) (*apikey.ApiKey, error) {
	query := `SELECT user_id, access_token, is_active, created_at FROM api_keys WHERE user_id = $1 AND is_active = true ORDER BY created_at DESC LIMIT 1`
	// Note: Adjust query for SQLite if needed (true -> 1)

	row := r.db.QueryRow(query, userID)

	var key apikey.ApiKey
	err := row.Scan(
		&key.UserID,
		&key.AccessToken,
		&key.IsActive,
		&key.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *ApiKeyRepository) FindByToken(token string) (*apikey.ApiKey, error) {
	query := `SELECT user_id, access_token, is_active, created_at FROM api_keys WHERE access_token = $1 AND is_active = true`
	row := r.db.QueryRow(query, token)

	var key apikey.ApiKey
	err := row.Scan(
		&key.UserID,
		&key.AccessToken,
		&key.IsActive,
		&key.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}
