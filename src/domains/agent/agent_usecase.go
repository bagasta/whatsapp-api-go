package agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/apikey"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type AgentUsecase struct {
	sessionRepo   session.ISessionRepository
	apiKeyRepo    apikey.IApiKeyRepository
	clientManager *whatsapp.ClientManager
	httpClient    *http.Client
}

func NewAgentUsecase(sessionRepo session.ISessionRepository, apiKeyRepo apikey.IApiKeyRepository, clientManager *whatsapp.ClientManager) IAgentUsecase {
	return &AgentUsecase{
		sessionRepo:   sessionRepo,
		apiKeyRepo:    apiKeyRepo,
		clientManager: clientManager,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (u *AgentUsecase) ExecuteRun(agentID, apiKey string, request RunRequest) (*RunResponse, error) {
	traceID := uuid.New().String()

	// 1. Validate API Key & Get User
	keyData, err := u.apiKeyRepo.FindByToken(apiKey)
	if err != nil {
		return nil, errors.New("UNAUTHORIZED")
	}
	userID := keyData.UserID

	// 2. Get session info
	sessionData, err := u.sessionRepo.FindOne(userID, agentID)
	if err != nil {
		return nil, errors.New("SESSION_NOT_FOUND")
	}

	// 3. Normalize input
	input := request.Input
	if input == "" {
		input = request.Message
	}
	if input == "" {
		return nil, errors.New("input or message is required")
	}

	sessionID := request.SessionID
	if sessionID == "" {
		return nil, errors.New("session_id is required")
	}

	// Normalize JID
	jid := normalizeJID(sessionID)

	// Set default parameters
	params := request.Parameters
	if params == nil {
		params = map[string]interface{}{
			"max_steps": 5,
		}
	}

	// 4. Call AI Backend
	endpoint := sessionData.EndpointUrlRun
	if endpoint == "" {
		// Fallback to default
		endpoint = fmt.Sprintf("%s/agents/%s/execute", getAIBackendURL(), agentID)
	}

	aiPayload := map[string]interface{}{
		"input":      input,
		"session_id": jid,
		"parameters": params,
	}

	reply, err := u.callAIBackend(endpoint, apiKey, aiPayload, traceID)
	if err != nil {
		logrus.Errorf("[%s] AI call failed: %v", traceID, err)
		return nil, err
	}

	// 5. Send reply via WhatsApp if present
	replySent := false
	if reply != "" {
		client := u.clientManager.GetClient(agentID)
		if client == nil || !client.IsLoggedIn() {
			return nil, errors.New("SESSION_NOT_READY")
		}

		if err := u.sendText(client, jid, reply); err != nil {
			logrus.Errorf("[%s] Failed to send reply: %v", traceID, err)
		} else {
			replySent = true
		}
	}

	return &RunResponse{
		Reply:     reply,
		ReplySent: replySent,
		TraceID:   traceID,
	}, nil
}

func (u *AgentUsecase) SendMessage(agentID, apiKey string, request SendMessageRequest) (*SendMessageResponse, error) {
	// Validate
	if request.To == "" || request.Message == "" {
		return nil, errors.New("to and message are required")
	}

	// Get client
	client := u.clientManager.GetClient(agentID)
	if client == nil || !client.IsLoggedIn() {
		return nil, errors.New("SESSION_NOT_READY")
	}

	// Normalize JID
	jid := normalizeJID(request.To)

	// Send message
	if err := u.sendText(client, jid, request.Message); err != nil {
		return nil, err
	}

	return &SendMessageResponse{Delivered: true}, nil
}

func (u *AgentUsecase) SendMedia(agentID, apiKey string, request SendMediaRequest) (*SendMediaResponse, error) {
	// Validate
	if request.To == "" {
		return nil, errors.New("to is required")
	}
	if request.Data == "" && request.URL == "" {
		return nil, errors.New("data or url is required")
	}

	// Get client
	client := u.clientManager.GetClient(agentID)
	if client == nil || !client.IsLoggedIn() {
		return nil, errors.New("SESSION_NOT_READY")
	}

	// Prepare media data
	var mediaData []byte
	var err error

	if request.Data != "" {
		// Decode base64
		mediaData, err = base64.StdEncoding.DecodeString(request.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 data: %w", err)
		}
	} else if request.URL != "" {
		// Download from URL
		resp, err := http.Get(request.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to download media: %w", err)
		}
		defer resp.Body.Close()
		mediaData, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read media body: %w", err)
		}
	}

	// Determine mime type
	mimeType := request.MimeType
	if mimeType == "" {
		mimeType = http.DetectContentType(mediaData)
	}

	// Determine media type for WhatsApp
	var appMediaType whatsmeow.MediaType
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		appMediaType = whatsmeow.MediaImage
	case strings.HasPrefix(mimeType, "video/"):
		appMediaType = whatsmeow.MediaVideo
	case strings.HasPrefix(mimeType, "audio/"):
		appMediaType = whatsmeow.MediaAudio
	default:
		appMediaType = whatsmeow.MediaDocument
	}

	// Upload media
	uploaded, err := client.Upload(context.Background(), mediaData, appMediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media to WhatsApp: %w", err)
	}

	// Construct message
	msg := &waProto.Message{}

	switch appMediaType {
	case whatsmeow.MediaImage:
		msg.ImageMessage = &waProto.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
			Caption:       proto.String(request.Caption),
		}
	case whatsmeow.MediaVideo:
		msg.VideoMessage = &waProto.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
			Caption:       proto.String(request.Caption),
		}
	case whatsmeow.MediaAudio:
		msg.AudioMessage = &waProto.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
		}
	case whatsmeow.MediaDocument:
		filename := request.Filename
		if filename == "" {
			filename = "file"
		}
		msg.DocumentMessage = &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
			FileName:      proto.String(filename),
			Caption:       proto.String(request.Caption),
		}
	}

	// Send message
	jid := normalizeJID(request.To)
	recipient, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	_, err = client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &SendMediaResponse{Delivered: true}, nil
}

// Helper functions

func (u *AgentUsecase) callAIBackend(endpoint, apiKey string, payload map[string]interface{}, traceID string) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("X-Trace-ID", traceID)

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI_TIMEOUT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI_DOWNSTREAM_ERROR: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Reply string `json:"reply"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result.Reply, nil
}

func (u *AgentUsecase) sendText(client *whatsmeow.Client, jid, text string) error {
	recipient, err := types.ParseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	_, err = client.SendMessage(context.Background(), recipient, msg)
	return err
}

func normalizeJID(input string) string {
	// Remove any whitespace
	input = strings.TrimSpace(input)

	// If already has @, return as is
	if strings.Contains(input, "@") {
		return input
	}

	// Add default suffix for user
	return input + "@s.whatsapp.net"
}

func getAIBackendURL() string {
	return config.AiBackendURL
}
