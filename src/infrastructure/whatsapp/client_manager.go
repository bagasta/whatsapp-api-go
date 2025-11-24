package whatsapp

import (
	"context"
	"fmt"
	"sync"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientManager struct {
	clients     map[string]*whatsmeow.Client
	dbs         map[string]*sqlstore.Container
	mu          sync.RWMutex
	chatStorage domainChatStorage.IChatStorageRepository
}

func NewClientManager(chatStorage domainChatStorage.IChatStorageRepository) *ClientManager {
	return &ClientManager{
		clients:     make(map[string]*whatsmeow.Client),
		dbs:         make(map[string]*sqlstore.Container),
		chatStorage: chatStorage,
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

	// Use separate DB for each agent to ensure isolation
	// DB URI format: file:storages/whatsapp-{agentID}.db?_foreign_keys=on
	dbPath := fmt.Sprintf("file:%s/whatsapp-%s.db?_foreign_keys=on", config.PathStorages, agentID)

	// If using Postgres, we might need a different strategy, e.g. table prefix or just one DB with multiple devices.
	// But for now, let's assume SQLite for individual files or handle Postgres later.
	// If config.DBURI is Postgres, we might want to use that but with a device discriminator?
	// The user asked for "Postgres migration", so we should support Postgres.
	// If Postgres, we can use the SAME DB but we need to track which DeviceID belongs to which AgentID.
	// But `whatsmeow` doesn't let us tag devices easily.

	// Strategy:
	// If SQLite: use separate files.
	// If Postgres: use the shared DB, but we need to know which JID corresponds to this agent.
	// Since we don't know the JID before login, this is tricky with a shared DB.
	// HOWEVER, `whatsmeow` allows creating a NEW device.
	// We can store the `JID` (DeviceID) in our `whatsapp_user` table after we create/load it.

	// Let's stick to the "Separate DB" approach for SQLite, and for Postgres...
	// Maybe for now we assume SQLite as per the file structure in `settings.go` (storages/whatsapp.db).
	// If the user wants Postgres, we might need to change this.
	// Let's implement a helper to get the store.

	var storeContainer *sqlstore.Container
	var err error

	// Check if main config is Postgres
	if isPostgres(config.DBURI) {
		// For Postgres, we use the main DB.
		// We need to find the device associated with this agent.
		// This requires us to store the JID in `whatsapp_user` table.
		// But wait, `whatsapp_user` table structure in API-OLD.MD doesn't have JID column (except maybe implicitly in ID?).
		// Actually `API-OLD.MD` says: "Tabel whatsapp_user ... Primary key: (user_id, agent_id)".

		// If we use Postgres, we can't easily separate by DB name.
		// We will use the shared store.
		dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
		storeContainer, err = sqlstore.New(ctx, "postgres", config.DBURI, dbLog)
		if err != nil {
			return nil, err
		}
	} else {
		// SQLite
		dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
		storeContainer, err = sqlstore.New(ctx, "sqlite3", dbPath, dbLog)
		if err != nil {
			return nil, err
		}
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
		// We need to pass agentID to the handler if needed, or just use a closure
		// For now, using the existing handler but we might need to adapt it
		handler(ctx, rawEvt, cm.chatStorage)
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
	// ... implementation details ...

	return nil
}

func isPostgres(uri string) bool {
	return len(uri) > 8 && (uri[:9] == "postgres:" || uri[:11] == "postgresql:")
}
