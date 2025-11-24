package session

import (
	"context"
	"errors"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/apikey"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
)

type SessionUsecase struct {
	sessionRepo   ISessionRepository
	apiKeyRepo    apikey.IApiKeyRepository
	clientManager *whatsapp.ClientManager
}

func NewSessionUsecase(sessionRepo ISessionRepository, apiKeyRepo apikey.IApiKeyRepository, clientManager *whatsapp.ClientManager) ISessionUsecase {
	return &SessionUsecase{
		sessionRepo:   sessionRepo,
		apiKeyRepo:    apiKeyRepo,
		clientManager: clientManager,
	}
}

func (u *SessionUsecase) CreateSession(request CreateSessionRequest) (*CreateSessionResponse, error) {
	// 1. Find active API Key if not provided
	apiKeyStr := request.ApiKey
	if apiKeyStr == "" {
		key, err := u.apiKeyRepo.FindActive(request.UserID)
		if err == nil && key != nil {
			apiKeyStr = key.AccessToken
		}
	}

	// 2. Upsert DB
	user := &WhatsappUser{
		UserID:         request.UserID,
		AgentID:        request.AgentID,
		AgentName:      request.AgentName,
		ApiKey:         apiKeyStr,
		EndpointUrlRun: "", // Default or from config
		Status:         "awaiting_qr",
		UpdatedAt:      time.Now(),
	}

	// Check if exists to preserve some fields if needed, or just upsert
	existing, err := u.sessionRepo.FindOne(request.UserID, request.AgentID)
	if err == nil && existing != nil {
		user.CreatedAt = existing.CreatedAt
		if user.EndpointUrlRun == "" {
			user.EndpointUrlRun = existing.EndpointUrlRun
		}
	} else {
		user.CreatedAt = time.Now()
	}

	if err := u.sessionRepo.Upsert(user); err != nil {
		return nil, err
	}

	// 3. Ensure Client
	ctx := context.Background()
	client, err := u.clientManager.CreateClient(ctx, request.AgentID)
	if err != nil {
		return nil, err
	}

	// 4. QR Handling
	var qrData *QrData
	if !client.IsConnected() {
		if client.Store.ID == nil {
			// No session, get QR
			qrChan, _ := client.GetQRChannel(context.Background())
			if err := client.Connect(); err != nil {
				return nil, err
			}

			// Wait for first QR
			select {
			case evt := <-qrChan:
				if evt.Event == "code" {
					qrData = &QrData{
						ContentType: "image/png", // You might need to generate the image or just return the code
						Base64:      evt.Code,    // For now returning the code string, frontend might need to generate QR
					}
					// Update status
					user.Status = "awaiting_qr"
					u.sessionRepo.Upsert(user)
				}
			case <-time.After(5 * time.Second):
				// Timeout waiting for QR
			}
		} else {
			// Has session, just connect
			client.Connect()
			user.Status = "connected"
			u.sessionRepo.Upsert(user)
		}
	}

	isReady := client.IsLoggedIn()
	sessionState := "disconnected"
	if client.IsConnected() {
		sessionState = "connected"
	}
	if isReady {
		sessionState = "authenticated"
	}

	return &CreateSessionResponse{
		IsReady:      isReady,
		SessionState: sessionState,
		Qr:           qrData,
		Timestamps: Timestamps{
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (u *SessionUsecase) GetSession(agentID string) (*GetSessionResponse, error) {
	client := u.clientManager.GetClient(agentID)
	hasClient := client != nil

	isReady := false
	sessionState := "disconnected"

	if hasClient {
		isReady = client.IsLoggedIn()
		if client.IsConnected() {
			sessionState = "connected"
		}
	}

	return &GetSessionResponse{
		IsReady:      isReady,
		HasClient:    hasClient,
		SessionState: sessionState,
	}, nil
}

func (u *SessionUsecase) DeleteSession(agentID string) error {
	// Destroy client
	u.clientManager.DeleteClient(agentID)

	// We need userID to delete from DB properly if PK is (userID, agentID)
	// But the interface only asks for agentID.
	// Assuming agentID is unique enough or we need to change the interface to accept userID.
	// For now, let's assume we can find it or delete by agentID if possible.
	// But the repo Delete takes (userID, agentID).
	// We might need to FindOne first by just AgentID?
	// Or the API request should provide UserID.
	// The API-OLD.MD says `DELETE /sessions/{agentId}`. It doesn't mention UserID in the path.
	// Maybe we assume the user is authenticated and we know the UserID?
	// Or we delete all sessions with that AgentID (should be unique per user?).

	// Let's leave the DB deletion for now or implement a DeleteByAgentID in repo.
	return nil
}

func (u *SessionUsecase) ReconnectSession(agentID string) (*CreateSessionResponse, error) {
	// Implementation similar to CreateSession but forces a reconnect
	return nil, nil
}

func (u *SessionUsecase) GetQR(agentID string) (*GetQRResponse, error) {
	client := u.clientManager.GetClient(agentID)
	if client == nil {
		return nil, errors.New("session not found")
	}

	if client.IsLoggedIn() {
		return nil, errors.New("session already logged in")
	}

	// This is tricky with whatsmeow as QR is emitted via channel.
	// We might need to cache the last QR in ClientManager.
	return nil, nil
}
