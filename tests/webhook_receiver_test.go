package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// WebhookReceiver Service Tests

func TestWebhookReceiverService_New(t *testing.T) {
	wr := service.NewWebhookReceiverService()
	if wr == nil {
		t.Fatal("Expected non-nil webhook receiver service")
	}
}

func TestWebhookReceiverService_RegisterEndpoint(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "endpoint-1",
		Name:     "Test Endpoint",
		Path:     "/webhooks/test",
		TenantID: "tenant-1",
		Secret:   "webhook-secret",
		Enabled:  true,
	}

	err := wr.RegisterEndpoint(endpoint)
	if err != nil {
		t.Fatalf("Failed to register endpoint: %v", err)
	}
}

func TestWebhookReceiverService_RegisterEndpoint_MissingPath(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "no-path",
		Name:     "No Path",
		TenantID: "tenant-1",
	}

	err := wr.RegisterEndpoint(endpoint)
	if err == nil {
		t.Error("Expected error for missing path")
	}
}

func TestWebhookReceiverService_GetEndpoint(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "get-endpoint",
		Name:     "Get Endpoint",
		Path:     "/webhooks/get",
		TenantID: "tenant-get",
	}
	wr.RegisterEndpoint(endpoint)

	retrieved, err := wr.GetEndpoint("get-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}

	if retrieved.Name != "Get Endpoint" {
		t.Error("Endpoint name mismatch")
	}
}

func TestWebhookReceiverService_GetEndpoint_ByPath(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "bypath-endpoint",
		Name:     "ByPath Endpoint",
		Path:     "/webhooks/bypath",
		TenantID: "tenant-bypath",
	}
	wr.RegisterEndpoint(endpoint)

	retrieved, err := wr.GetEndpointByPath("/webhooks/bypath")
	if err != nil {
		t.Fatalf("Failed to get endpoint by path: %v", err)
	}

	if retrieved.Name != "ByPath Endpoint" {
		t.Error("Endpoint name mismatch")
	}
}

func TestWebhookReceiverService_GetEndpoint_NotFound(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	_, err := wr.GetEndpoint("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent endpoint")
	}
}

func TestWebhookReceiverService_ListEndpoints(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	wr.RegisterEndpoint(&service.WebhookEndpoint{
		ID:       "list-1",
		Name:     "List 1",
		Path:     "/webhooks/list1",
		TenantID: "tenant-list",
	})

	wr.RegisterEndpoint(&service.WebhookEndpoint{
		ID:       "list-2",
		Name:     "List 2",
		Path:     "/webhooks/list2",
		TenantID: "tenant-list",
	})

	endpoints := wr.ListEndpoints("tenant-list")
	if len(endpoints) < 2 {
		t.Errorf("Expected at least 2 endpoints, got %d", len(endpoints))
	}
}

func TestWebhookReceiverService_DeleteEndpoint(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "delete-endpoint",
		Name:     "Delete Endpoint",
		Path:     "/webhooks/delete",
		TenantID: "tenant-delete",
	}
	wr.RegisterEndpoint(endpoint)

	err := wr.DeleteEndpoint("delete-endpoint")
	if err != nil {
		t.Fatalf("Failed to delete endpoint: %v", err)
	}

	_, err = wr.GetEndpoint("delete-endpoint")
	if err == nil {
		t.Error("Expected error for deleted endpoint")
	}
}

func TestWebhookReceiverService_Receive(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "receive-endpoint",
		Name:     "Receive Endpoint",
		Path:     "/webhooks/receive",
		TenantID: "tenant-receive",
		Secret:   "secret123",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	payload := map[string]interface{}{
		"event": "user.created",
		"data": map[string]interface{}{
			"user_id": "123",
			"email":   "test@example.com",
		},
	}

	receipt, err := wr.Receive("receive-endpoint", payload, map[string]string{
		"X-Signature": "signature",
	})

	if err != nil {
		t.Fatalf("Failed to receive webhook: %v", err)
	}

	if receipt.ID == "" {
		t.Error("Expected receipt ID")
	}
}

func TestWebhookReceiverService_VerifySignature(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "verify-endpoint",
		Name:     "Verify Endpoint",
		Path:     "/webhooks/verify",
		TenantID: "tenant-verify",
		Secret:   "my-secret",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	payload := []byte(`{"event":"test"}`)
	signature := wr.GenerateSignature("my-secret", payload)

	valid := wr.VerifySignature("verify-endpoint", payload, signature)
	if !valid {
		t.Error("Expected signature to be valid")
	}
}

func TestWebhookReceiverService_VerifySignature_Invalid(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "invalidsig-endpoint",
		Name:     "Invalid Sig Endpoint",
		Path:     "/webhooks/invalidsig",
		TenantID: "tenant-invalidsig",
		Secret:   "correct-secret",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	payload := []byte(`{"event":"test"}`)

	valid := wr.VerifySignature("invalidsig-endpoint", payload, "wrong-signature")
	if valid {
		t.Error("Expected signature to be invalid")
	}
}

func TestWebhookReceiverService_EnableEndpoint(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "enable-endpoint",
		Name:     "Enable Endpoint",
		Path:     "/webhooks/enable",
		TenantID: "tenant-enable",
		Enabled:  false,
	}
	wr.RegisterEndpoint(endpoint)

	err := wr.EnableEndpoint("enable-endpoint")
	if err != nil {
		t.Fatalf("Failed to enable endpoint: %v", err)
	}

	retrieved, _ := wr.GetEndpoint("enable-endpoint")
	if !retrieved.Enabled {
		t.Error("Endpoint should be enabled")
	}
}

func TestWebhookReceiverService_DisableEndpoint(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "disable-endpoint",
		Name:     "Disable Endpoint",
		Path:     "/webhooks/disable",
		TenantID: "tenant-disable",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	err := wr.DisableEndpoint("disable-endpoint")
	if err != nil {
		t.Fatalf("Failed to disable endpoint: %v", err)
	}

	retrieved, _ := wr.GetEndpoint("disable-endpoint")
	if retrieved.Enabled {
		t.Error("Endpoint should be disabled")
	}
}

func TestWebhookReceiverService_GetReceipt(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "receipt-endpoint",
		Name:     "Receipt Endpoint",
		Path:     "/webhooks/receipt",
		TenantID: "tenant-receipt",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	receipt, _ := wr.Receive("receipt-endpoint", map[string]interface{}{"test": "data"}, nil)

	retrieved, err := wr.GetReceipt(receipt.ID)
	if err != nil {
		t.Fatalf("Failed to get receipt: %v", err)
	}

	if retrieved.EndpointID != "receipt-endpoint" {
		t.Error("Endpoint ID mismatch")
	}
}

func TestWebhookReceiverService_ListReceipts(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "listreceipts-endpoint",
		Name:     "List Receipts Endpoint",
		Path:     "/webhooks/listreceipts",
		TenantID: "tenant-listreceipts",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	wr.Receive("listreceipts-endpoint", map[string]interface{}{"a": 1}, nil)
	wr.Receive("listreceipts-endpoint", map[string]interface{}{"b": 2}, nil)

	receipts := wr.ListReceipts("listreceipts-endpoint")
	if len(receipts) < 2 {
		t.Errorf("Expected at least 2 receipts, got %d", len(receipts))
	}
}

func TestWebhookReceiverService_SetHandler(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "handler-endpoint",
		Name:     "Handler Endpoint",
		Path:     "/webhooks/handler",
		TenantID: "tenant-handler",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	handler := func(payload map[string]interface{}, headers map[string]string) error {
		return nil
	}

	err := wr.SetHandler("handler-endpoint", handler)
	if err != nil {
		t.Fatalf("Failed to set handler: %v", err)
	}
}

func TestWebhookReceiverService_GetStats(t *testing.T) {
	wr := service.NewWebhookReceiverService()

	endpoint := &service.WebhookEndpoint{
		ID:       "stats-endpoint",
		Name:     "Stats Endpoint",
		Path:     "/webhooks/stats",
		TenantID: "tenant-stats",
		Enabled:  true,
	}
	wr.RegisterEndpoint(endpoint)

	wr.Receive("stats-endpoint", map[string]interface{}{"test": 1}, nil)

	stats := wr.GetStats("stats-endpoint")

	if stats.TotalReceived < 1 {
		t.Errorf("Expected at least 1 received, got %d", stats.TotalReceived)
	}
}

func TestWebhookReceiverService_WebhookEndpointToJSON(t *testing.T) {
	endpoint := &service.WebhookEndpoint{
		ID:        "json-endpoint",
		Name:      "JSON Endpoint",
		Path:      "/webhooks/json",
		TenantID:  "tenant-1",
		Secret:    "secret",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	data, err := endpoint.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestWebhookReceiverService_WebhookReceiptToJSON(t *testing.T) {
	receipt := &service.WebhookReceipt{
		ID:         "receipt-json",
		EndpointID: "endpoint-1",
		TenantID:   "tenant-1",
		Payload:    map[string]interface{}{"key": "value"},
		Headers:    map[string]string{"X-Event": "test"},
		Status:     "received",
		ReceivedAt: time.Now(),
	}

	data, err := receipt.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}