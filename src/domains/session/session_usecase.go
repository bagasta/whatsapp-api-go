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
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
)

type SessionUsecase struct {
	sessionRepo   ISessionRepository
	apiKeyRepo    apikey.IApiKeyRepository
	clientManager *whatsapp.ClientManager
}

var ErrQRNotReady = errors.New("qr not ready; retry shortly while device is emitting QR")

// fetchNextQR tries to subscribe to the QR channel and return the next code quickly without
// tearing down the existing connection. Useful when the client is connected but cache is empty.
func (u *SessionUsecase) fetchNextQR(ctx context.Context, client *whatsmeow.Client, agentID string) (*QrData, error) {
	qrCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	qrChan, err := client.GetQRChannel(qrCtx)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case evt, ok := <-qrChan:
			if !ok {
				return nil, ErrQRNotReady
			}
			if evt.Event != "code" {
				continue
			}
			png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			encoded := base64.StdEncoding.EncodeToString(png)
			qr := &QrData{ContentType: "image/png", Base64: encoded}
			u.clientManager.CacheQR(agentID, qr.ContentType, qr.Base64)
			return qr, nil
		case <-qrCtx.Done():
			return nil, ErrQRNotReady
		}
	}
}

func NewSessionUsecase(sessionRepo ISessionRepository, apiKeyRepo apikey.IApiKeyRepository, clientManager *whatsapp.ClientManager) ISessionUsecase {
	return &SessionUsecase{
		sessionRepo:   sessionRepo,
		apiKeyRepo:    apiKeyRepo,
		clientManager: clientManager,
	}
}

// listenAndCacheQR connects the client and keeps caching rotating QR codes.
// Returns the first QR code emitted or an error if none is received within the timeout.
func (u *SessionUsecase) listenAndCacheQR(ctx context.Context, client *whatsmeow.Client, agentID string) (*QrData, error) {
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		if errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			return nil, errors.New("session already saved; reconnect or logout first")
		}
		return nil, err
	}

	if err := client.Connect(); err != nil {
		return nil, err
	}

	firstQR := make(chan *QrData, 1)
	go func() {
		defer close(firstQR)
		for evt := range qrChan {
			if evt.Event != "code" {
				continue
			}
			png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			encoded := base64.StdEncoding.EncodeToString(png)
			qr := &QrData{ContentType: "image/png", Base64: encoded}
			u.clientManager.CacheQR(agentID, qr.ContentType, qr.Base64)

			// capture the first QR to return
			select {
			case firstQR <- qr:
			default:
			}
		}
	}()

	select {
	case qr, ok := <-firstQR:
		if !ok || qr == nil {
			return nil, errors.New("qr channel closed")
		}
		return qr, nil
	case <-time.After(60 * time.Second):
		return nil, errors.New("qr timeout")
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

	// If store exists but not logged in, reset the agent store to avoid invalid/expired QR attempts
	if client.Store.ID != nil && !client.IsLoggedIn() {
		logrus.Warnf("Resetting store for agent %s before QR (store exists but not logged in)", request.AgentID)
		if err := u.clientManager.DeleteClient(request.AgentID); err != nil {
			logrus.Warnf("Failed to delete client for agent %s: %v", request.AgentID, err)
		}
		client, err = u.clientManager.CreateClient(ctx, request.AgentID)
		if err != nil {
			return nil, err
		}
	}

	// 4. QR Handling
	var qrData *QrData
	if !client.IsConnected() {
		if client.Store.ID == nil {
			qrData, err = u.listenAndCacheQR(ctx, client, request.AgentID)
			if err != nil {
				return nil, err
			}
			user.Status = "awaiting_qr"
			u.sessionRepo.Upsert(user)
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

	ctx := context.Background()

	if ct, b64, ok, ts := u.clientManager.GetCachedQR(agentID); ok {
		return &GetQRResponse{Qr: QrData{ContentType: ct, Base64: b64}, QrUpdatedAt: ts}, nil
	}

	// If connected but waiting for scan, QR emitter is already running; wait for cache to update
	if client.IsConnected() && !client.IsLoggedIn() {
		logrus.Debugf("QR already emitting for agent %s; waiting for next cached code", agentID)
		// short wait loop for refreshed cache (e.g., rotate every ~15s)
		for i := 0; i < 3; i++ {
			time.Sleep(3 * time.Second)
			if ct, b64, ok, ts := u.clientManager.GetCachedQR(agentID); ok {
				return &GetQRResponse{Qr: QrData{ContentType: ct, Base64: b64}, QrUpdatedAt: ts}, nil
			}
		}
		// still empty: subscribe again to get a fresh code without disconnecting
		qr, err := u.fetchNextQR(ctx, client, agentID)
		if err != nil {
			return nil, err
		}
		return &GetQRResponse{Qr: *qr, QrUpdatedAt: time.Now()}, nil
	}

	// Not connected and no cache: start QR flow
	qr, err := u.listenAndCacheQR(ctx, client, agentID)
	if err != nil {
		return nil, err
	}
	return &GetQRResponse{Qr: *qr, QrUpdatedAt: time.Now()}, nil
}

func (u *SessionUsecase) getCachedQR(agentID string) *QrData {
	if ct, b64, ok, _ := u.clientManager.GetCachedQR(agentID); ok {
		return &QrData{ContentType: ct, Base64: b64}
	}
	return nil
}
