package agent

type RunRequest struct {
	Input      string                 `json:"input"`
	Message    string                 `json:"message"`
	SessionID  string                 `json:"session_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

type RunResponse struct {
	Reply     string `json:"reply"`
	ReplySent bool   `json:"replySent"`
	TraceID   string `json:"traceId,omitempty"`
}

type SendMessageRequest struct {
	To              string `json:"to"`
	Message         string `json:"message"`
	QuotedMessageID string `json:"quotedMessageId,omitempty"`
}

type SendMessageResponse struct {
	Delivered bool `json:"delivered"`
}

type SendMediaRequest struct {
	To       string `json:"to"`
	Data     string `json:"data,omitempty"` // base64
	URL      string `json:"url,omitempty"`  // URL to download
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type SendMediaResponse struct {
	Delivered   bool   `json:"delivered"`
	PreviewPath string `json:"previewPath,omitempty"`
}

type IAgentUsecase interface {
	ExecuteRun(agentID, apiKey string, request RunRequest) (*RunResponse, error)
	SendMessage(agentID, apiKey string, request SendMessageRequest) (*SendMessageResponse, error)
	SendMedia(agentID, apiKey string, request SendMediaRequest) (*SendMediaResponse, error)
}
