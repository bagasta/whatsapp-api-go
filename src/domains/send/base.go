package send

type BaseRequest struct {
	Phone       string `json:"phone" form:"phone"`
	AgentID     string `json:"agent_id,omitempty" form:"agent_id"`
	Duration    *int   `json:"duration,omitempty" form:"duration"`
	IsForwarded bool   `json:"is_forwarded,omitempty" form:"is_forwarded"`
}
