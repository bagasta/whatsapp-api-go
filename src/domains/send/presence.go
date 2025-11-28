package send

type PresenceRequest struct {
	Type        string `json:"type" form:"type"`
	AgentID     string `json:"agent_id,omitempty" form:"agent_id"`
	IsForwarded bool   `json:"is_forwarded" form:"is_forwarded"`
}
