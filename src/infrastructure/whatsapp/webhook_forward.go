package whatsapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/sirupsen/logrus"
)

var submitWebhookFn = submitWebhook
var webhookResolver = func(agentID string) ([]string, string) {
	return config.WhatsappWebhook, config.WhatsappWebhookSecret
}

// SetWebhookResolver allows higher layers to provide dynamic webhook targets per agent.
func SetWebhookResolver(fn func(agentID string) ([]string, string)) {
	if fn != nil {
		webhookResolver = fn
	}
}

// forwardPayloadToConfiguredWebhooks attempts to deliver the provided payload to every configured webhook URL.
// It only returns an error when all webhook deliveries fail. Partial failures are logged and suppressed so
// successful targets still receive the event.
func forwardPayloadToConfiguredWebhooks(ctx context.Context, payload map[string]any, eventName, agentID string) error {
	urls, secret := webhookResolver(agentID)
	// Clean empty URLs to avoid noisy attempts
	filtered := make([]string, 0, len(urls))
	for _, u := range urls {
		if strings.TrimSpace(u) != "" {
			filtered = append(filtered, u)
		}
	}

	total := len(filtered)
	logrus.Infof("Forwarding %s for agent %s to %d webhook(s)", eventName, agentID, total)

	if total == 0 {
		logrus.Infof("No webhook configured for %s (agent %s); skipping dispatch", eventName, agentID)
		return nil
	}

	var (
		failed    []string
		successes int
	)
	for _, url := range filtered {
		if err := submitWebhookFn(ctx, payload, url, secret); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", url, err))
			logrus.Warnf("Failed forwarding %s (agent %s) to %s: %v", eventName, agentID, url, err)
			continue
		}
		successes++
	}

	if len(failed) == total {
		return pkgError.WebhookError(fmt.Sprintf("all webhook URLs failed for %s (agent %s): %s", eventName, agentID, strings.Join(failed, "; ")))
	}

	if len(failed) > 0 {
		logrus.Warnf("Some webhook URLs failed for %s (agent %s) (succeeded: %d/%d): %s", eventName, agentID, successes, total, strings.Join(failed, "; "))
	} else {
		logrus.Infof("%s forwarded to all webhook(s) for agent %s", eventName, agentID)
	}

	return nil
}
