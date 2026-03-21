package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// APIKey Service Tests

func TestAPIKeyService_New(t *testing.T) {
	aks := service.NewAPIKeyService()
	if aks == nil {
		t.Fatal("Expected non-nil API key service")
	}
}

func TestAPIKeyService_CreateKey(t *testing.T) {
	aks := service.NewAPIKeyService()

	key, err := aks.CreateKey("tenant-1", "test-key", []string{"read", "write"}, 30)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	if key.Key == "" {
		t.Error("Expected key to be generated")
	}

	if key.TenantID != "tenant-1" {
		t.Error("Tenant ID mismatch")
	}

	if len(key.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(key.Scopes))
	}
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	aks := service.NewAPIKeyService()

	key, _ := aks.CreateKey("tenant-1", "test-key", []string{"read"}, 30)

	validated, err := aks.ValidateKey(key.Key)
	if err != nil {
		t.Fatalf("Failed to validate key: %v", err)
	}

	if validated.TenantID != "tenant-1" {
		t.Error("Validated key tenant mismatch")
	}
}

func TestAPIKeyService_ValidateKey_Invalid(t *testing.T) {
	aks := service.NewAPIKeyService()

	_, err := aks.ValidateKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestAPIKeyService_RevokeKey(t *testing.T) {
	aks := service.NewAPIKeyService()

	key, _ := aks.CreateKey("tenant-1", "test-key", []string{"read"}, 30)

	err := aks.RevokeKey(key.ID)
	if err != nil {
		t.Fatalf("Failed to revoke key: %v", err)
	}

	_, err = aks.ValidateKey(key.Key)
	if err == nil {
		t.Error("Expected error for revoked key")
	}
}

func TestAPIKeyService_ListKeys(t *testing.T) {
	aks := service.NewAPIKeyService()

	aks.CreateKey("tenant-list", "key-1", []string{"read"}, 30)
	aks.CreateKey("tenant-list", "key-2", []string{"write"}, 30)
	aks.CreateKey("other-tenant", "key-3", []string{"read"}, 30)

	keys := aks.ListKeys("tenant-list")
	if len(keys) < 2 {
		t.Errorf("Expected at least 2 keys, got %d", len(keys))
	}
}

func TestAPIKeyService_CheckScope(t *testing.T) {
	aks := service.NewAPIKeyService()

	key, _ := aks.CreateKey("tenant-1", "test-key", []string{"read", "write"}, 30)

	if !aks.CheckScope(key.Key, "read") {
		t.Error("Expected read scope to be allowed")
	}

	if aks.CheckScope(key.Key, "admin") {
		t.Error("Expected admin scope to be denied")
	}
}

func TestAPIKeyService_Expiration(t *testing.T) {
	aks := service.NewAPIKeyService()

	// Create key that expires in 1 day
	key, _ := aks.CreateKey("tenant-1", "expiring-key", []string{"read"}, 1)

	if key.ExpiresAt == nil {
		t.Error("Expected expiration date to be set")
	}
}

func TestAPIKeyService_CleanupExpired(t *testing.T) {
	aks := service.NewAPIKeyService()

	// Create and immediately expire
	key, _ := aks.CreateKey("tenant-1", "expired-key", []string{"read"}, 0)
	// Manually set to past
	past := time.Now().Add(-24 * time.Hour)
	aks.SetExpiration(key.ID, past)

	deleted := aks.CleanupExpiredKeys()
	if deleted < 1 {
		t.Error("Expected at least 1 expired key to be cleaned up")
	}
}

func TestAPIKeyService_KeyToJSON(t *testing.T) {
	key := &service.APIKey{
		ID:        "key-1",
		Name:      "Test Key",
		TenantID:  "tenant-1",
		Key:       "sk-test-123",
		Scopes:    []string{"read", "write"},
		CreatedAt: time.Now(),
	}

	data, err := key.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}