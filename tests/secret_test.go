package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Secret Service Tests

func TestSecretBasic_New(t *testing.T) {
	ss := service.NewSecretService()
	if ss == nil {
		t.Fatal("Expected non-nil secret service")
	}
}

func TestSecretService_Create(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "secret-1",
		Name:     "API Key",
		Value:    "super-secret-key",
		TenantID: "tenant-1",
		Type:     service.SecretTypeAPIKey,
	}

	err := ss.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Value should be encrypted, not stored in plain text
	if secret.EncryptedValue == "" {
		t.Error("Expected encrypted value to be set")
	}
}

func TestSecretService_Create_MissingName(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "no-name",
		Value:    "test",
		TenantID: "tenant-1",
	}

	err := ss.Create(secret)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestSecretService_Get(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "get-secret",
		Name:     "Get Secret",
		Value:    "secret-value",
		TenantID: "tenant-get",
	}
	ss.Create(secret)

	retrieved, err := ss.Get("get-secret")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if retrieved.Name != "Get Secret" {
		t.Error("Secret name mismatch")
	}
}

func TestSecretService_Get_NotFound(t *testing.T) {
	ss := service.NewSecretService()

	_, err := ss.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent secret")
	}
}

func TestSecretService_GetValue(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "value-secret",
		Name:     "Value Secret",
		Value:    "my-secret-value",
		TenantID: "tenant-value",
	}
	ss.Create(secret)

	value, err := ss.GetValue("value-secret")
	if err != nil {
		t.Fatalf("Failed to get secret value: %v", err)
	}

	if value != "my-secret-value" {
		t.Error("Secret value mismatch")
	}
}

func TestSecretService_List(t *testing.T) {
	ss := service.NewSecretService()

	ss.Create(&service.Secret{
		ID:       "list-1",
		Name:     "List 1",
		Value:    "v1",
		TenantID: "tenant-list",
	})

	ss.Create(&service.Secret{
		ID:       "list-2",
		Name:     "List 2",
		Value:    "v2",
		TenantID: "tenant-list",
	})

	secrets := ss.List("tenant-list")
	if len(secrets) < 2 {
		t.Errorf("Expected at least 2 secrets, got %d", len(secrets))
	}
}

func TestSecretService_Update(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "update-secret",
		Name:     "Update Secret",
		Value:    "old-value",
		TenantID: "tenant-update",
	}
	ss.Create(secret)

	err := ss.Update("update-secret", map[string]interface{}{
		"value": "new-value",
	})

	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	value, _ := ss.GetValue("update-secret")
	if value != "new-value" {
		t.Error("Secret value not updated")
	}
}

func TestSecretService_Delete(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "delete-secret",
		Name:     "Delete Secret",
		Value:    "to-delete",
		TenantID: "tenant-delete",
	}
	ss.Create(secret)

	err := ss.Delete("delete-secret")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	_, err = ss.Get("delete-secret")
	if err == nil {
		t.Error("Expected error for deleted secret")
	}
}

func TestSecretService_Rotate(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "rotate-secret",
		Name:     "Rotate Secret",
		Value:    "original-value",
		TenantID: "tenant-rotate",
	}
	ss.Create(secret)

	err := ss.Rotate("rotate-secret", "new-rotated-value")
	if err != nil {
		t.Fatalf("Failed to rotate secret: %v", err)
	}

	value, _ := ss.GetValue("rotate-secret")
	if value != "new-rotated-value" {
		t.Error("Secret not rotated")
	}

	retrieved, _ := ss.Get("rotate-secret")
	if retrieved.Version != 2 {
		t.Errorf("Expected version 2, got %d", retrieved.Version)
	}
}

func TestSecretService_GetByVersion(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "version-secret",
		Name:     "Version Secret",
		Value:    "v1",
		TenantID: "tenant-version",
	}
	ss.Create(secret)
	ss.Rotate("version-secret", "v2")
	ss.Rotate("version-secret", "v3")

	value, err := ss.GetByVersion("version-secret", 1)
	if err != nil {
		t.Fatalf("Failed to get secret by version: %v", err)
	}

	if value != "v1" {
		t.Error("Version 1 value mismatch")
	}
}

func TestSecretService_GetHistory(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "history-secret",
		Name:     "History Secret",
		Value:    "v1",
		TenantID: "tenant-history",
	}
	ss.Create(secret)
	ss.Rotate("history-secret", "v2")
	ss.Rotate("history-secret", "v3")

	history := ss.GetHistory("history-secret")
	if len(history) < 3 {
		t.Errorf("Expected at least 3 history entries, got %d", len(history))
	}
}

func TestSecretService_SetMetadata(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "meta-secret",
		Name:     "Meta Secret",
		Value:    "test",
		TenantID: "tenant-meta",
	}
	ss.Create(secret)

	err := ss.SetMetadata("meta-secret", map[string]interface{}{
		"description": "API key for external service",
		"owner":       "team-backend",
	})

	if err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}
}

func TestSecretService_GetMetadata(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "getmeta-secret",
		Name:     "Get Meta Secret",
		Value:    "test",
		TenantID: "tenant-getmeta",
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}
	ss.Create(secret)

	metadata := ss.GetMetadata("getmeta-secret")
	if metadata == nil {
		t.Fatal("Expected metadata")
	}

	if metadata["key"] != "value" {
		t.Error("Metadata value mismatch")
	}
}

func TestSecretService_SetExpiry(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "expiry-secret",
		Name:     "Expiry Secret",
		Value:    "test",
		TenantID: "tenant-expiry",
	}
	ss.Create(secret)

	expiry := time.Now().Add(30 * 24 * time.Hour)
	err := ss.SetExpiry("expiry-secret", expiry)
	if err != nil {
		t.Fatalf("Failed to set expiry: %v", err)
	}
}

func TestSecretService_IsExpired(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "checkexpired-secret",
		Name:     "Check Expired Secret",
		Value:    "test",
		TenantID: "tenant-checkexpired",
	}
	ss.Create(secret)

	// Not expired yet
	if ss.IsExpired("checkexpired-secret") {
		t.Error("Secret should not be expired")
	}

	// Set to past
	ss.SetExpiry("checkexpired-secret", time.Now().Add(-time.Hour))

	if !ss.IsExpired("checkexpired-secret") {
		t.Error("Secret should be expired")
	}
}

func TestSecretService_CleanupExpired(t *testing.T) {
	ss := service.NewSecretService()

	secret := &service.Secret{
		ID:       "cleanup-secret",
		Name:     "Cleanup Secret",
		Value:    "test",
		TenantID: "tenant-cleanup",
	}
	ss.Create(secret)
	ss.SetExpiry("cleanup-secret", time.Now().Add(-time.Hour))

	deleted := ss.CleanupExpired()
	if deleted < 1 {
		t.Error("Expected at least 1 expired secret to be cleaned up")
	}
}

func TestSecretService_SecretTypes(t *testing.T) {
	types := []service.SecretType{
		service.SecretTypeAPIKey,
		service.SecretTypePassword,
		service.SecretTypeCertificate,
		service.SecretTypeSSHKey,
		service.SecretTypeToken,
		service.SecretTypeGeneric,
	}

	for _, st := range types {
		if string(st) == "" {
			t.Errorf("Secret type %s is empty", st)
		}
	}
}

func TestSecretService_SecretToJSON(t *testing.T) {
	secret := &service.Secret{
		ID:        "json-secret",
		Name:      "JSON Secret",
		TenantID:  "tenant-1",
		Type:      service.SecretTypeAPIKey,
		Version:   1,
		CreatedAt: time.Now(),
	}

	data, err := secret.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}