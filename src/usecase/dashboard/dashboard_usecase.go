package dashboard

import (
	"errors"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/dashboard"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
)

type DashboardUsecase struct {
	repo        dashboard.IDashboardRepository
	sessionRepo session.ISessionRepository
}

func NewDashboardUsecase(repo dashboard.IDashboardRepository, sessionRepo session.ISessionRepository) dashboard.IDashboardUsecase {
	return &DashboardUsecase{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

func (u *DashboardUsecase) Login(agentID, apiKey string) (string, error) {
	user, err := u.sessionRepo.FindByAgentID(agentID)
	if err != nil {
		return "", errors.New("invalid agent ID or API key")
	}
	if user.ApiKey != apiKey {
		return "", errors.New("invalid agent ID or API key")
	}
	return user.ApiKey, nil
}

func (u *DashboardUsecase) GetAnalytics(agentID string) (*dashboard.DashboardAnalytics, error) {
	return u.repo.GetAnalytics(agentID)
}

func (u *DashboardUsecase) LogAiHandled(agentID, messageID, userID string) error {
	return u.repo.LogAiMessage(agentID, messageID, userID, "success")
}
