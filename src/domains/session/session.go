package session

import "time"

type WhatsappUser struct {
	UserID             string     `json:"user_id"`
	AgentID            string     `json:"agent_id"`
	AgentName          string     `json:"agent_name"`
	ApiKey             string     `json:"api_key"`
	EndpointUrlRun     string     `json:"endpoint_url_run"`
	Status             string     `json:"status"`
	LastConnectedAt    *time.Time `json:"last_connected_at"`
	LastDisconnectedAt *time.Time `json:"last_disconnected_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateSessionRequest struct {
	UserID    string `json:"userId"`
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName"`
	ApiKey    string `json:"apikey"`
	// EndpointUrlRun overrides the AI endpoint; default is AI_BACKEND_URL/agents/{agentId}/execute
	EndpointUrlRun string `json:"endpointUrlRun"`
}

type CreateSessionResponse struct {
	IsReady      bool       `json:"isReady"`
	SessionState string     `json:"sessionState"`
	Qr           *QrData    `json:"qr,omitempty"`
	Timestamps   Timestamps `json:"timestamps"`
}

type QrData struct {
	ContentType string `json:"contentType"`
	Base64      string `json:"base64"`
}

type Timestamps struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GetSessionResponse struct {
	IsReady      bool    `json:"isReady"`
	HasClient    bool    `json:"hasClient"`
	SessionState string  `json:"sessionState"`
	Qr           *QrData `json:"qr,omitempty"`
}

type GetQRResponse struct {
	Qr          QrData    `json:"qr"`
	QrUpdatedAt time.Time `json:"qrUpdatedAt"`
}

type ISessionRepository interface {
	Upsert(user *WhatsappUser) error
	FindOne(userID, agentID string) (*WhatsappUser, error)
	FindByAgentID(agentID string) (*WhatsappUser, error)
	Delete(userID, agentID string) error
	List() ([]*WhatsappUser, error)
}

type ISessionUsecase interface {
	CreateSession(request CreateSessionRequest) (*CreateSessionResponse, error)
	GetSession(agentID string) (*GetSessionResponse, error)
	DeleteSession(agentID string) error
	ReconnectSession(agentID string) (*CreateSessionResponse, error)
	GetQR(agentID string) (*GetQRResponse, error)
}
