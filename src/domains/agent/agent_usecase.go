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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/apikey"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/metrics"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

type AgentUsecase struct {
	sessionRepo   session.ISessionRepository
	apiKeyRepo    apikey.IApiKeyRepository
	clientManager *whatsapp.ClientManager
	httpClient    *http.Client

	limiterMu sync.Mutex
	limiters  map[string]*agentLimiter
}

type agentLimiter struct {
	limiter *rate.Limiter
	queue   chan struct{}
}

func NewAgentUsecase(sessionRepo session.ISessionRepository, apiKeyRepo apikey.IApiKeyRepository, clientManager *whatsapp.ClientManager) IAgentUsecase {
	return &AgentUsecase{
		sessionRepo:   sessionRepo,
		apiKeyRepo:    apiKeyRepo,
		clientManager: clientManager,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		limiters: make(map[string]*agentLimiter),
	}
}

func (u *AgentUsecase) ExecuteRun(agentID, apiKey string, request RunRequest) (*RunResponse, error) {
	traceID := uuid.New().String()
	ctx := context.Background()

	if err := u.acquire(agentID, ctx); err != nil {
		return nil, err
	}

	userID, err := u.validateAPIKey(agentID, apiKey)
	if err != nil {
		return nil, err
	}

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

	// Ensure client ready before calling AI (needed for typing + reply)
	client := u.clientManager.GetClient(agentID)
	if client == nil || !client.IsLoggedIn() {
		return nil, errors.New("SESSION_NOT_READY")
	}

	// 4. Typing indicator ON while waiting for AI
	u.sendTyping(ctx, client, jid, true)
	defer u.sendTyping(context.Background(), client, jid, false)

	// 5. Call AI Backend
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

	replyRaw, err := u.callAIBackend(endpoint, apiKey, aiPayload, traceID)
	if err != nil {
		logrus.Errorf("[%s] AI call failed: %v", traceID, err)
		return nil, err
	}
	reply := sanitizeReply(replyRaw)

	// 6. Send reply via WhatsApp if present
	replySent := false
	if reply != "" {
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
	if err := u.acquire(agentID, context.Background()); err != nil {
		return nil, err
	}
	if _, err := u.validateAPIKey(agentID, apiKey); err != nil {
		return nil, err
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
	if err := u.acquire(agentID, context.Background()); err != nil {
		return nil, err
	}
	if _, err := u.validateAPIKey(agentID, apiKey); err != nil {
		return nil, err
	}

	// Get client
	client := u.clientManager.GetClient(agentID)
	if client == nil || !client.IsLoggedIn() {
		return nil, errors.New("SESSION_NOT_READY")
	}

	const maxMediaSize = 10 * 1024 * 1024 // 10MB

	// Prepare media data
	var mediaData []byte
	var err error

	if request.Data != "" {
		mediaData, err = base64.StdEncoding.DecodeString(request.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 data: %w", err)
		}
		if len(mediaData) > maxMediaSize {
			return nil, errors.New("MEDIA_TOO_LARGE")
		}
	} else if request.URL != "" {
		// HEAD to check size when possible
		if resp, err := http.Head(request.URL); err == nil {
			if resp.ContentLength > 0 && resp.ContentLength > maxMediaSize {
				return nil, errors.New("MEDIA_TOO_LARGE")
			}
		}
		resp, err := http.Get(request.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to download media: %w", err)
		}
		defer resp.Body.Close()

		limited := &io.LimitedReader{R: resp.Body, N: maxMediaSize + 1}
		mediaData, err = io.ReadAll(limited)
		if err != nil {
			return nil, fmt.Errorf("failed to read media body: %w", err)
		}
		if int64(len(mediaData)) > maxMediaSize {
			return nil, errors.New("MEDIA_TOO_LARGE")
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

	// Save preview for images (compat with API-OLD)
	previewPath := ""
	if strings.HasPrefix(mimeType, "image/") {
		uuidName := uuid.New().String()
		ext := filepath.Ext(request.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		previewPath = filepath.Join(config.PathSendItems, "preview-"+uuidName+ext)
		_ = os.MkdirAll(config.PathSendItems, 0755)
		_ = os.WriteFile(previewPath, mediaData, 0644)
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

	metrics.IncMessagesSent()

	return &SendMediaResponse{Delivered: true, PreviewPath: previewPath}, nil
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

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AI response: %w", err)
	}

	// Support multiple shapes:
	// 1) { "reply": "text" }
	// 2) LangChain style: { "output": { "output": "text", ... }, "status": "COMPLETED" }
	// 3) Fallback to top-level "output" string
	var result struct {
		Reply    string `json:"reply"`
		Response string `json:"response"` // backend AI variant
		Output   struct {
			Output string `json:"output"`
			Final  string `json:"final"`
			Msg    string `json:"message"`
		} `json:"output"`
		OutputString string `json:"output"`
		Message      string `json:"message"`
		ErrorMessage string `json:"error_message"`
	}

	_ = json.Unmarshal(raw, &result)

	reply := result.Response
	if reply == "" {
		reply = result.Reply
	}
	if reply == "" {
		switch {
		case result.Output.Output != "":
			reply = result.Output.Output
		case result.Output.Final != "":
			reply = result.Output.Final
		case result.Output.Msg != "":
			reply = result.Output.Msg
		case result.OutputString != "":
			reply = result.OutputString
		case result.Message != "":
			reply = result.Message
		}
	}

	if reply == "" {
		return "", fmt.Errorf("AI response missing reply text: %s", string(raw))
	}

	return reply, nil
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
	if err == nil {
		metrics.IncMessagesSent()
	}
	return err
}

// sendTyping sends chat presence typing indicators; errors are logged but not fatal.
func (u *AgentUsecase) sendTyping(ctx context.Context, client *whatsmeow.Client, jid string, start bool) {
	recipient, err := types.ParseJID(jid)
	if err != nil || client == nil {
		return
	}
	var presence types.ChatPresence
	if start {
		presence = types.ChatPresenceComposing
	} else {
		presence = types.ChatPresencePaused
	}
	// Use text media hint so WA shows typing indicator
	if err := client.SendChatPresence(ctx, recipient, presence, types.ChatPresenceMediaText); err != nil {
		logrus.Debugf("sendTyping failed (%v): %v", presence, err)
	}
}

// sanitizeReply normalizes AI responses to be WhatsApp-friendly (strip markdown noise).
func sanitizeReply(text string) string {
	if text == "" {
		return ""
	}
	// Remove code fences
	text = regexp.MustCompile("(?s)```.*?```").ReplaceAllString(text, "")
	// Remove headings ###
	text = regexp.MustCompile(`(?m)^#{1,6}\s*`).ReplaceAllString(text, "")
	// Bullet points
	text = regexp.MustCompile(`(?m)^\s*[-*]\s+`).ReplaceAllString(text, "• ")
	// Collapse multiple blank lines
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func (u *AgentUsecase) acquire(agentID string, ctx context.Context) error {
	u.limiterMu.Lock()
	lim, ok := u.limiters[agentID]
	if !ok {
		lim = &agentLimiter{
			limiter: rate.NewLimiter(rate.Every(time.Minute/100), 100),
			queue:   make(chan struct{}, 500),
		}
		u.limiters[agentID] = lim
	}
	u.limiterMu.Unlock()

	select {
	case lim.queue <- struct{}{}:
		defer func() { <-lim.queue }()
		if err := lim.limiter.Wait(ctx); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("RATE_LIMITED")
	}
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

// validateAPIKey mimics API-OLD behavior:
// 1) accept token found in api_keys (active)
// 2) accept token that matches session.ApiKey saved at /sessions creation
// 3) else unauthorized
func (u *AgentUsecase) validateAPIKey(agentID, token string) (string, error) {
	if token == "" {
		return "", errors.New("UNAUTHORIZED")
	}

	if key, err := u.apiKeyRepo.FindByToken(token); err == nil && key != nil {
		return key.UserID, nil
	}

	sessionData, err := u.sessionRepo.FindByAgentID(agentID)
	if err != nil || sessionData == nil {
		return "", errors.New("UNAUTHORIZED")
	}

	// If session has no apiKey yet, adopt incoming token
	if sessionData.ApiKey == "" {
		sessionData.ApiKey = token
		_ = u.sessionRepo.Upsert(sessionData)
		return sessionData.UserID, nil
	}

	// Match stored apiKey
	if sessionData.ApiKey == token {
		return sessionData.UserID, nil
	}

	// Accept latest active key in api_keys table
	if active, err := u.apiKeyRepo.FindActive(sessionData.UserID); err == nil && active != nil {
		if active.AccessToken == token {
			return sessionData.UserID, nil
		}
	}

	return "", errors.New("UNAUTHORIZED")
}
