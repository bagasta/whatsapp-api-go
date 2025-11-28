package whatsapp

import (
	"fmt"

	"go.mau.fi/whatsmeow"
)

// globalClientManager keeps a reference so non-agent-aware code can resolve clients by agent ID.
var globalClientManager *ClientManager

// SetClientManager registers a global ClientManager for resolving per-agent clients.
func SetClientManager(manager *ClientManager) {
	globalClientManager = manager
}

// ResolveClient returns the WhatsApp client for a given agent.
// If agentID is empty, it falls back to the legacy global client.
func ResolveClient(agentID string) (*whatsmeow.Client, error) {
	if agentID != "" {
		if globalClientManager == nil {
			return nil, fmt.Errorf("client manager not initialized")
		}
		client := globalClientManager.GetClient(agentID)
		if client == nil {
			return nil, fmt.Errorf("client for agent %s not initialized", agentID)
		}
		return client, nil
	}

	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("global whatsapp client not initialized")
	}
	return client, nil
}
