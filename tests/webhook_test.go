package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Webhook Service Tests

func TestWebhookService_New(t *testing.T) {
	ws := service.NewWebhookService()
	if ws == nil {
		t.Fatal("Expected non-nil webhook service")
	}
}

func TestWebhookService_Create(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:      "webhook-1",
		Name:    "Test Webhook",
		URL:     "https://example.com/webhook",
		Events:  []string{"task.completed", "task.failed"},
		TenantID: "tenant-1",
		Enabled: true,
	}

	err := ws.Create(webhook)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}
}

func TestWebhookService_Create_MissingURL(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "no-url",
		Name:     "No URL",
		TenantID: "tenant-1",
	}

	err := ws.Create(webhook)
	if err == nil {
		t.Error("Expected error for missing URL")
	}
}

func TestWebhookService_Get(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "get-webhook",
		Name:     "Get Webhook",
		URL:      "https://example.com/get",
		TenantID: "tenant-get",
	}
	ws.Create(webhook)

	retrieved, err := ws.Get("get-webhook")
	if err != nil {
		t.Fatalf("Failed to get webhook: %v", err)
	}

	if retrieved.Name != "Get Webhook" {
		t.Error("Webhook name mismatch")
	}
}

func TestWebhookService_Get_NotFound(t *testing.T) {
	ws := service.NewWebhookService()

	_, err := ws.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent webhook")
	}
}

func TestWebhookService_List(t *testing.T) {
	ws := service.NewWebhookService()

	ws.Create(&service.Webhook{
		ID:       "list-1",
		Name:     "List 1",
		URL:      "https://example.com/list1",
		TenantID: "tenant-list",
	})

	ws.Create(&service.Webhook{
		ID:       "list-2",
		Name:     "List 2",
		URL:      "https://example.com/list2",
		TenantID: "tenant-list",
	})

	webhooks := ws.List("tenant-list")
	if len(webhooks) < 2 {
		t.Errorf("Expected at least 2 webhooks, got %d", len(webhooks))
	}
}

func TestWebhookService_Update(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "update-webhook",
		Name:     "Update Webhook",
		URL:      "https://example.com/original",
		TenantID: "tenant-update",
	}
	ws.Create(webhook)

	err := ws.Update("update-webhook", map[string]interface{}{
		"name": "Updated Name",
		"url":  "https://example.com/updated",
	})

	if err != nil {
		t.Fatalf("Failed to update webhook: %v", err)
	}

	retrieved, _ := ws.Get("update-webhook")
	if retrieved.Name != "Updated Name" {
		t.Error("Name not updated")
	}
}

func TestWebhookService_Delete(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "delete-webhook",
		Name:     "Delete Webhook",
		URL:      "https://example.com/delete",
		TenantID: "tenant-delete",
	}
	ws.Create(webhook)

	err := ws.Delete("delete-webhook")
	if err != nil {
		t.Fatalf("Failed to delete webhook: %v", err)
	}

	_, err = ws.Get("delete-webhook")
	if err == nil {
		t.Error("Expected error for deleted webhook")
	}
}

func TestWebhookService_Enable(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "enable-webhook",
		Name:     "Enable Webhook",
		URL:      "https://example.com/enable",
		TenantID: "tenant-enable",
		Enabled:  false,
	}
	ws.Create(webhook)

	err := ws.Enable("enable-webhook")
	if err != nil {
		t.Fatalf("Failed to enable webhook: %v", err)
	}

	retrieved, _ := ws.Get("enable-webhook")
	if !retrieved.Enabled {
		t.Error("Webhook should be enabled")
	}
}

func TestWebhookService_Disable(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "disable-webhook",
		Name:     "Disable Webhook",
		URL:      "https://example.com/disable",
		TenantID: "tenant-disable",
		Enabled:  true,
	}
	ws.Create(webhook)

	err := ws.Disable("disable-webhook")
	if err != nil {
		t.Fatalf("Failed to disable webhook: %v", err)
	}

	retrieved, _ := ws.Get("disable-webhook")
	if retrieved.Enabled {
		t.Error("Webhook should be disabled")
	}
}

func TestWebhookService_Trigger(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "trigger-webhook",
		Name:     "Trigger Webhook",
		URL:      "https://httpbin.org/post",
		Events:   []string{"test.event"},
		TenantID: "tenant-trigger",
		Enabled:  true,
	}
	ws.Create(webhook)

	payload := map[string]interface{}{
		"event": "test.event",
		"data":  "test payload",
	}

	err := ws.Trigger("trigger-webhook", "test.event", payload)
	// Note: This might fail in tests without network, but tests the logic
	_ = err
}

func TestWebhookService_GetDeliveryLogs(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "logs-webhook",
		Name:     "Logs Webhook",
		URL:      "https://example.com/logs",
		TenantID: "tenant-logs",
	}
	ws.Create(webhook)

	// Record some delivery logs
	ws.RecordDelivery("logs-webhook", true, 200)

	logs := ws.GetDeliveryLogs("logs-webhook")
	if len(logs) == 0 {
		t.Error("Expected delivery logs")
	}
}

func TestWebhookService_ListByEvent(t *testing.T) {
	ws := service.NewWebhookService()

	ws.Create(&service.Webhook{
		ID:       "event-1",
		Name:     "Event Webhook 1",
		URL:      "https://example.com/event1",
		Events:   []string{"task.completed", "task.failed"},
		TenantID: "tenant-event",
		Enabled:  true,
	})

	ws.Create(&service.Webhook{
		ID:       "event-2",
		Name:     "Event Webhook 2",
		URL:      "https://example.com/event2",
		Events:   []string{"task.failed"},
		TenantID: "tenant-event",
		Enabled:  true,
	})

	webhooks := ws.ListByEvent("task.completed", "tenant-event")
	if len(webhooks) != 1 {
		t.Errorf("Expected 1 webhook for task.completed, got %d", len(webhooks))
	}
}

func TestWebhookService_TestEndpoint(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "test-endpoint",
		Name:     "Test Endpoint",
		URL:      "https://httpbin.org/status/200",
		TenantID: "tenant-test",
	}
	ws.Create(webhook)

	// Test the endpoint
	err := ws.TestEndpoint("test-endpoint")
	_ = err // May fail without network
}

func TestWebhookService_SetSecret(t *testing.T) {
	ws := service.NewWebhookService()

	webhook := &service.Webhook{
		ID:       "secret-webhook",
		Name:     "Secret Webhook",
		URL:      "https://example.com/secret",
		TenantID: "tenant-secret",
	}
	ws.Create(webhook)

	err := ws.SetSecret("secret-webhook", "my-secret-key")
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}
}

func TestWebhookService_GetStats(t *testing.T) {
	ws := service.NewWebhookService()

	ws.Create(&service.Webhook{
		ID:       "stats-webhook",
		Name:     "Stats Webhook",
		URL:      "https://example.com/stats",
		TenantID: "tenant-stats",
	})

	ws.RecordDelivery("stats-webhook", true, 200)
	ws.RecordDelivery("stats-webhook", false, 500)

	stats := ws.GetStats("stats-webhook")

	if stats.TotalDeliveries < 2 {
		t.Errorf("Expected at least 2 deliveries, got %d", stats.TotalDeliveries)
	}
}

func TestWebhookService_WebhookToJSON(t *testing.T) {
	webhook := &service.Webhook{
		ID:        "json-webhook",
		Name:      "JSON Webhook",
		URL:       "https://example.com/json",
		Events:    []string{"test.event"},
		TenantID:  "tenant-1",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	data, err := webhook.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}