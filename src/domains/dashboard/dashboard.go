package dashboard

import "time"

type AiMessageLog struct {
	ID        int64     `json:"id"`
	AgentID   string    `json:"agent_id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardAnalytics struct {
	TotalAiHandled int64 `json:"total_ai_handled"`
}

type IDashboardRepository interface {
	LogAiMessage(agentID, messageID, userID, status string) error
	GetAnalytics(agentID string) (*DashboardAnalytics, error)
}

type IDashboardUsecase interface {
	Login(agentID, apiKey string) (string, error)
	GetAnalytics(agentID string) (*DashboardAnalytics, error)
	LogAiHandled(agentID, messageID, userID string) error
}
