package apikey

import "time"

type ApiKey struct {
	UserID      string    `json:"user_id"`
	AccessToken string    `json:"access_token"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type IApiKeyRepository interface {
	FindActive(userID string) (*ApiKey, error)
	FindByToken(token string) (*ApiKey, error)
}
