package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientManager struct {
	clients     map[string]*whatsmeow.Client
	dbs         map[string]*sqlstore.Container
	mu          sync.RWMutex
	chatStorage domainChatStorage.IChatStorageRepository

	qrCache map[string]cachedQR
}

type cachedQR struct {
	contentType string
	base64      string
	updatedAt   int64 // unix seconds
}

func NewClientManager(chatStorage domainChatStorage.IChatStorageRepository) *ClientManager {
	return &ClientManager{
		clients:     make(map[string]*whatsmeow.Client),
		dbs:         make(map[string]*sqlstore.Container),
		chatStorage: chatStorage,
		qrCache:     make(map[string]cachedQR),
	}
}

func (cm *ClientManager) GetClient(agentID string) *whatsmeow.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[agentID]
}

func (cm *ClientManager) CreateClient(ctx context.Context, agentID string) (*whatsmeow.Client, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.clients[agentID]; ok {
		return client, nil
	}

	// Use separate DB for each agent to ensure isolation.
	// IMPORTANT: Even if main DB is Postgres, sharing the same store across multiple
	// agents will cause "stream replaced" conflicts. To avoid that, we keep a dedicated
	// SQLite store per agent.
	dbPath := fmt.Sprintf("file:%s/whatsapp-%s.db?_foreign_keys=on&_journal_mode=WAL", config.PathStorages, agentID)

	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
	storeContainer, err := sqlstore.New(ctx, "sqlite3", dbPath, dbLog)
	if err != nil {
		return nil, err
	}

	cm.dbs[agentID] = storeContainer

	// Get or Create Device
	// If SQLite (separate DB), GetFirstDevice is fine because there's only one.
	// If Postgres (shared DB), we have a problem: GetFirstDevice returns *any* device.
	// We need a way to link AgentID <-> DeviceID (JID).

	// TEMPORARY SOLUTION:
	// For this implementation, I will assume SQLite (Separate Files) is the primary mode for Multi-Agent
	// because it maps 1:1 with the "Session File" concept.
	// If Postgres is used, this code might need a mapping table "AgentDevice" or add "device_jid" to "whatsapp_user".

	device, err := storeContainer.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}
	if device == nil {
		// Create new device
		device = storeContainer.NewDevice()
	}

	// Configure device properties
	osName := fmt.Sprintf("%s %s", config.AppOs, config.AppVersion)
	store.DeviceProps.PlatformType = &config.AppPlatform
	store.DeviceProps.Os = &osName

	// Create Client
	clientLog := waLog.Stdout(fmt.Sprintf("Client-%s", agentID), config.WhatsappLogLevel, true)
	client := whatsmeow.NewClient(device, clientLog)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	// Add Event Handler
	client.AddEventHandler(func(rawEvt interface{}) {
		// Recover from pairing failures caused by a corrupted store (e.g., FK constraint errors)
		if pairErr, ok := rawEvt.(*events.PairError); ok {
			if pairErr.Error != nil {
				logrus.Errorf("Pairing failed for agent %s: %v", agentID, pairErr.Error)
			} else {
				logrus.Errorf("Pairing failed for agent %s with unknown error", agentID)
			}

			// Reset the client + per-agent DB so the next QR attempt starts clean
			if err := cm.DeleteClient(agentID); err != nil {
				logrus.Warnf("Failed to reset client after pair error for agent %s: %v", agentID, err)
			}
			return
		}

		// include agentID to avoid cross-agent conflicts and enable per-agent hooks
		handler(ctx, rawEvt, cm.chatStorage, agentID, client)
	})

	cm.clients[agentID] = client
	return client, nil
}

func (cm *ClientManager) DeleteClient(agentID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.clients[agentID]; ok {
		client.Disconnect()
		delete(cm.clients, agentID)
	}

	if db, ok := cm.dbs[agentID]; ok {
		db.Close()
		delete(cm.dbs, agentID)
	}

	// Remove DB file if SQLite
	if !isPostgres(config.DBURI) {
		dbFile := filepath.Join(config.PathStorages, fmt.Sprintf("whatsapp-%s.db", agentID))
		_ = os.Remove(dbFile)
		_ = os.Remove(dbFile + "-shm")
		_ = os.Remove(dbFile + "-wal")
	}

	delete(cm.qrCache, agentID)

	return nil
}

// CacheQR stores the latest QR code for an agent.
func (cm *ClientManager) CacheQR(agentID string, contentType, base64data string) {
	cm.mu.Lock()
	cm.qrCache[agentID] = cachedQR{contentType: contentType, base64: base64data, updatedAt: time.Now().Unix()}
	cm.mu.Unlock()
}

// GetCachedQR returns the cached QR if present.
func (cm *ClientManager) GetCachedQR(agentID string) (contentType, base64data string, ok bool, updatedAt time.Time) {
	cm.mu.RLock()
	qr, exists := cm.qrCache[agentID]
	cm.mu.RUnlock()
	if !exists {
		return "", "", false, time.Time{}
	}
	return qr.contentType, qr.base64, true, time.Unix(qr.updatedAt, 0)
}

func isPostgres(uri string) bool {
	return len(uri) > 8 && (uri[:9] == "postgres:" || uri[:11] == "postgresql:")
}
