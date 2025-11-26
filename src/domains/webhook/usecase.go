package webhook

import (
	"errors"
	"strings"
	"sync"
	"time"

	"database/sql"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainSession "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
)

type ConfigUsecase struct {
	repo        IWebhookConfigRepository
	sessionRepo domainSession.ISessionRepository
	cache       sync.Map // key: agentID or "__default__", value: *Config
}

func NewWebhookConfigUsecase(repo IWebhookConfigRepository, sessionRepo domainSession.ISessionRepository) IWebhookConfigUsecase {
	return &ConfigUsecase{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

func (u *ConfigUsecase) GetDefault() (*Config, error) {
	return u.repo.GetDefault()
}

func (u *ConfigUsecase) SaveDefault(url, secret string) (*Config, error) {
	return u.save("", url, secret)
}

func (u *ConfigUsecase) GetForAgent(agentID string) (*Config, error) {
	if strings.TrimSpace(agentID) == "" {
		return nil, errors.New("agentId is required")
	}
	return u.repo.GetByAgent(agentID)
}

func (u *ConfigUsecase) SaveForAgent(agentID, url, secret string) (*Config, error) {
	if strings.TrimSpace(agentID) == "" {
		return nil, errors.New("agentId is required")
	}
	// Ensure session exists to prevent dangling config
	if _, err := u.sessionRepo.FindByAgentID(agentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return u.save(agentID, url, secret)
}

func (u *ConfigUsecase) save(agentID, url, secret string) (*Config, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("url is required")
	}

	cfg := &Config{
		AgentID:   strings.TrimSpace(agentID),
		URL:       url,
		Secret:    strings.TrimSpace(secret),
		UpdatedAt: time.Now(),
	}

	if agentID == "" {
		if err := u.repo.UpsertDefault(cfg); err != nil {
			return nil, err
		}
		u.cache.Store("__default__", cfg)
		applyRuntimeConfig(cfg)
		return cfg, nil
	}

	if err := u.repo.UpsertByAgent(cfg); err != nil {
		return nil, err
	}
	u.cache.Store(agentID, cfg)
	return cfg, nil
}

func (u *ConfigUsecase) ListSessions() ([]SessionSummary, error) {
	sessions, err := u.sessionRepo.List()
	if err != nil {
		return nil, err
	}
	result := make([]SessionSummary, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, SessionSummary{
			UserID:    s.UserID,
			AgentID:   s.AgentID,
			AgentName: s.AgentName,
			Status:    s.Status,
		})
	}
	return result, nil
}

// SyncRuntimeConfig loads stored default config (if any) and applies it to runtime config variables.
func (u *ConfigUsecase) SyncRuntimeConfig() error {
	cfg, err := u.repo.GetDefault()
	if err != nil {
		return err
	}
	if cfg != nil {
		u.cache.Store("__default__", cfg)
		applyRuntimeConfig(cfg)
	}
	return nil
}

// ResolveWebhooks returns the webhook URLs and secret for a given agent.
// Order of precedence:
// 1) Agent-specific config (DB)
// 2) Default config (DB)
// 3) Config from environment variables
func (u *ConfigUsecase) ResolveWebhooks(agentID string) ([]string, string) {
	if cfg := u.loadFromCache(agentID); cfg != nil && cfg.URL != "" {
		return []string{cfg.URL}, cfg.Secret
	}

	if cfg, _ := u.repo.GetByAgent(agentID); cfg != nil && cfg.URL != "" {
		u.cache.Store(agentID, cfg)
		return []string{cfg.URL}, cfg.Secret
	}

	if cfg := u.loadFromCache("__default__"); cfg != nil && cfg.URL != "" {
		return []string{cfg.URL}, cfg.Secret
	}
	if cfg, _ := u.repo.GetDefault(); cfg != nil && cfg.URL != "" {
		u.cache.Store("__default__", cfg)
		return []string{cfg.URL}, cfg.Secret
	}

	return config.WhatsappWebhook, config.WhatsappWebhookSecret
}

func (u *ConfigUsecase) loadFromCache(key string) *Config {
	if val, ok := u.cache.Load(key); ok {
		if cfg, ok := val.(*Config); ok {
			return cfg
		}
	}
	return nil
}

func applyRuntimeConfig(cfg *Config) {
	config.WhatsappWebhook = []string{cfg.URL}
	config.WhatsappWebhookSecret = cfg.Secret
}
