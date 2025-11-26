package whatsapp

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestForwardPayloadToConfiguredWebhooks_NoWebhooksConfigured(t *testing.T) {
	ctx := context.Background()
	payload := map[string]any{"foo": "bar"}

	originalResolver := webhookResolver
	webhookResolver = func(string) ([]string, string) { return nil, "" }
	defer func() { webhookResolver = originalResolver }()

	originalSubmit := submitWebhookFn
	submitWebhookFn = func(context.Context, map[string]any, string, string) error {
		t.Fatal("submitWebhookFn should not be invoked when no webhooks are configured")
		return nil
	}
	defer func() { submitWebhookFn = originalSubmit }()

	if err := forwardPayloadToConfiguredWebhooks(ctx, payload, "test", "agent-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestForwardPayloadToConfiguredWebhooks_PartialFailure(t *testing.T) {
	ctx := context.Background()
	payload := map[string]any{"foo": "bar"}

	originalResolver := webhookResolver
	webhookResolver = func(string) ([]string, string) {
		return []string{"https://success", "https://fail", "https://success2"}, "secret"
	}
	defer func() { webhookResolver = originalResolver }()

	originalSubmit := submitWebhookFn
	var attempts []string
	submitWebhookFn = func(_ context.Context, _ map[string]any, url, _ string) error {
		attempts = append(attempts, url)
		if strings.Contains(url, "fail") {
			return errors.New("boom")
		}
		return nil
	}
	defer func() { submitWebhookFn = originalSubmit }()

	if err := forwardPayloadToConfiguredWebhooks(ctx, payload, "test", "agent-1"); err != nil {
		t.Fatalf("expected partial failure to return nil, got %v", err)
	}

	if len(attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(attempts))
	}
}

func TestForwardPayloadToConfiguredWebhooks_AllFail(t *testing.T) {
	ctx := context.Background()
	payload := map[string]any{"foo": "bar"}

	originalResolver := webhookResolver
	webhookResolver = func(string) ([]string, string) {
		return []string{"https://fail1", "https://fail2"}, "secret"
	}
	defer func() { webhookResolver = originalResolver }()

	originalSubmit := submitWebhookFn
	submitWebhookFn = func(_ context.Context, _ map[string]any, url, _ string) error {
		return errors.New("failure for " + url)
	}
	defer func() { submitWebhookFn = originalSubmit }()

	if err := forwardPayloadToConfiguredWebhooks(ctx, payload, "test", "agent-1"); err == nil {
		t.Fatalf("expected error when all webhooks fail")
	}
}
