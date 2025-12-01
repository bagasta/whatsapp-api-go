package repository

import (
	"database/sql"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/dashboard"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) dashboard.IDashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) LogAiMessage(agentID, messageID, userID, status string) error {
	query := `INSERT INTO ai_message_logs (agent_id, message_id, user_id, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, agentID, messageID, userID, status)
	return err
}

func (r *DashboardRepository) GetAnalytics(agentID string) (*dashboard.DashboardAnalytics, error) {
	query := `SELECT COUNT(*) FROM ai_message_logs WHERE agent_id = $1 AND status = 'success'`
	var count int64
	err := r.db.QueryRow(query, agentID).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &dashboard.DashboardAnalytics{
		TotalAiHandled: count,
	}, nil
}
