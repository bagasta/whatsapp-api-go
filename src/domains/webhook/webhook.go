package webhook

import "time"

// Config represents webhook destination settings (e.g., n8n).
type Config struct {
	AgentID   string    `json:"agentId,omitempty"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IWebhookConfigRepository interface {
	GetDefault() (*Config, error)
	UpsertDefault(cfg *Config) error
	GetByAgent(agentID string) (*Config, error)
	UpsertByAgent(cfg *Config) error
}

type IWebhookConfigUsecase interface {
	GetDefault() (*Config, error)
	SaveDefault(url, secret string) (*Config, error)
	GetForAgent(agentID string) (*Config, error)
	SaveForAgent(agentID, url, secret string) (*Config, error)
	ListSessions() ([]SessionSummary, error)
	SyncRuntimeConfig() error
	ResolveWebhooks(agentID string) ([]string, string)
}

// SessionSummary is a lightweight view of a session for admin listing.
type SessionSummary struct {
	UserID    string `json:"userId"`
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName"`
	Status    string `json:"status"`
}
