package session

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/apikey"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/metrics"
	"github.com/skip2/go-qrcode"
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

	endpointURL := request.EndpointUrlRun
	if endpointURL == "" {
		endpointURL = fmt.Sprintf("%s/agents/%s/execute", config.AiBackendURL, request.AgentID)
	}

	// 2. Upsert DB
	user := &WhatsappUser{
		UserID:         request.UserID,
		AgentID:        request.AgentID,
		AgentName:      request.AgentName,
		ApiKey:         apiKeyStr,
		EndpointUrlRun: endpointURL,
		Status:         "awaiting_qr",
		UpdatedAt:      time.Now(),
	}

	// Check if exists to preserve some fields if needed, or just upsert
	existing, err := u.sessionRepo.FindOne(request.UserID, request.AgentID)
	newSession := true
	if err == nil && existing != nil {
		newSession = false
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
					png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
					encoded := base64.StdEncoding.EncodeToString(png)
					qrData = &QrData{ContentType: "image/png", Base64: encoded}
					user.Status = "awaiting_qr"
					u.sessionRepo.Upsert(user)
					u.clientManager.CacheQR(request.AgentID, qrData.ContentType, qrData.Base64)
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

	if newSession {
		metrics.AddSessions(1)
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
		Qr:           u.getCachedQR(agentID),
	}, nil
}

func (u *SessionUsecase) DeleteSession(agentID string) error {
	user, err := u.sessionRepo.FindByAgentID(agentID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	u.clientManager.DeleteClient(agentID)

	if user != nil {
		_ = u.sessionRepo.Delete(user.UserID, agentID)
		metrics.AddSessions(-1)
	}

	return nil
}

func (u *SessionUsecase) ReconnectSession(agentID string) (*CreateSessionResponse, error) {
	user, err := u.sessionRepo.FindByAgentID(agentID)
	if err != nil {
		return nil, err
	}

	// Destroy old client and recreate
	u.clientManager.DeleteClient(agentID)

	req := CreateSessionRequest{
		UserID:         user.UserID,
		AgentID:        user.AgentID,
		AgentName:      user.AgentName,
		ApiKey:         user.ApiKey,
		EndpointUrlRun: user.EndpointUrlRun,
	}

	return u.CreateSession(req)
}

func (u *SessionUsecase) GetQR(agentID string) (*GetQRResponse, error) {
	client := u.clientManager.GetClient(agentID)
	if client == nil {
		return nil, errors.New("session not found")
	}

	if client.IsLoggedIn() {
		return nil, errors.New("session already logged in")
	}

	if ct, b64, ok, ts := u.clientManager.GetCachedQR(agentID); ok {
		return &GetQRResponse{Qr: QrData{ContentType: ct, Base64: b64}, QrUpdatedAt: ts}, nil
	}

	ctx := context.Background()
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return nil, err
	}

	if !client.IsConnected() {
		_ = client.Connect()
	}

	select {
	case evt := <-qrChan:
		if evt.Event != "code" {
			return nil, errors.New("qr not available")
		}
		png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
		encoded := base64.StdEncoding.EncodeToString(png)
		qr := QrData{ContentType: "image/png", Base64: encoded}
		u.clientManager.CacheQR(agentID, qr.ContentType, qr.Base64)
		return &GetQRResponse{Qr: qr, QrUpdatedAt: time.Now()}, nil
	case <-time.After(60 * time.Second):
		return nil, errors.New("qr timeout")
	}
}

func (u *SessionUsecase) getCachedQR(agentID string) *QrData {
	if ct, b64, ok, _ := u.clientManager.GetCachedQR(agentID); ok {
		return &QrData{ContentType: ct, Base64: b64}
	}
	return nil
}
