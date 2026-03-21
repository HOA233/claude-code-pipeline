package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

func TestAuthService_New(t *testing.T) {
	auth := service.NewAuthService(nil)
	if auth == nil {
		t.Fatal("Expected non-nil auth service")
	}
}

func TestAuthService_CreateAPIKey(t *testing.T) {
	auth := service.NewAuthService(nil)

	apiKey, err := auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		Name:        "test-key",
		TenantID:    "tenant-1",
		UserID:      "user-1",
		Permissions: []string{"read", "write"},
		RateLimit:   100,
	})

	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if apiKey.Name != "test-key" {
		t.Errorf("Expected name 'test-key', got '%s'", apiKey.Name)
	}
	if !apiKey.Enabled {
		t.Error("Expected API key to be enabled")
	}
}

func TestAuthService_CreateAPIKey_MissingName(t *testing.T) {
	auth := service.NewAuthService(nil)

	_, err := auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		TenantID: "tenant-1",
	})

	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestAuthService_ValidateAPIKey(t *testing.T) {
	auth := service.NewAuthService(nil)

	created, _ := auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		Name:     "test-key",
		TenantID: "tenant-1",
	})

	validated, err := auth.ValidateAPIKey(context.Background(), created.Key)
	if err != nil {
		t.Fatalf("Failed to validate API key: %v", err)
	}

	if validated.ID != created.ID {
		t.Error("API key ID mismatch")
	}
}

func TestAuthService_ValidateAPIKey_Invalid(t *testing.T) {
	auth := service.NewAuthService(nil)

	_, err := auth.ValidateAPIKey(context.Background(), "invalid-key")
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAuthService_RevokeAPIKey(t *testing.T) {
	auth := service.NewAuthService(nil)

	created, _ := auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		Name:     "test-key",
		TenantID: "tenant-1",
	})

	err := auth.RevokeAPIKey(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to revoke API key: %v", err)
	}

	_, err = auth.ValidateAPIKey(context.Background(), created.Key)
	if err == nil {
		t.Error("Expected error for revoked API key")
	}
}

func TestAuthService_ListAPIKeys(t *testing.T) {
	auth := service.NewAuthService(nil)

	for i := 0; i < 3; i++ {
		auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
			Name:     string(rune('a' + i)),
			TenantID: "tenant-1",
		})
	}

	keys := auth.ListAPIKeys("tenant-1")
	if len(keys) < 3 {
		t.Errorf("Expected at least 3 keys, got %d", len(keys))
	}

	// Check that keys are masked
	for _, k := range keys {
		if k.Key == "" {
			t.Error("Expected masked key, got empty string")
		}
	}
}

func TestAuthService_CreateSession(t *testing.T) {
	auth := service.NewAuthService(nil)

	session, err := auth.CreateSession(context.Background(), "user-1", "tenant-1")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.UserID != "user-1" {
		t.Error("Session user ID mismatch")
	}
	if session.Token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestAuthService_ValidateSession(t *testing.T) {
	auth := service.NewAuthService(nil)

	created, _ := auth.CreateSession(context.Background(), "user-1", "tenant-1")

	validated, err := auth.ValidateSession(context.Background(), created.Token)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validated.ID != created.ID {
		t.Error("Session ID mismatch")
	}
}

func TestAuthService_ValidateSession_Invalid(t *testing.T) {
	auth := service.NewAuthService(nil)

	_, err := auth.ValidateSession(context.Background(), "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid session")
	}
}

func TestAuthService_RevokeSession(t *testing.T) {
	auth := service.NewAuthService(nil)

	session, _ := auth.CreateSession(context.Background(), "user-1", "tenant-1")

	err := auth.RevokeSession(context.Background(), session.Token)
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	_, err = auth.ValidateSession(context.Background(), session.Token)
	if err == nil {
		t.Error("Expected error for revoked session")
	}
}

func TestAuthService_CheckPermission(t *testing.T) {
	auth := service.NewAuthService(nil)

	auth.GrantPermission(context.Background(), "user-1", []string{"read", "write"})

	if !auth.CheckPermission(context.Background(), "user-1", "read", "") {
		t.Error("Expected user to have read permission")
	}
	if auth.CheckPermission(context.Background(), "user-1", "admin", "") {
		t.Error("Expected user to not have admin permission")
	}
}

func TestAuthService_Authenticate_APIKey(t *testing.T) {
	auth := service.NewAuthService(nil)

	created, _ := auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		Name:        "test-key",
		TenantID:    "tenant-1",
		UserID:      "user-1",
		Permissions: []string{"read"},
	})

	result := auth.Authenticate(context.Background(), &service.AuthRequest{
		APIKey: created.Key,
	})

	if !result.Authorized {
		t.Errorf("Expected authorized, got error: %s", result.Error)
	}
	if result.UserID != "user-1" {
		t.Error("User ID mismatch")
	}
}

func TestAuthService_Authenticate_Session(t *testing.T) {
	auth := service.NewAuthService(nil)

	session, _ := auth.CreateSession(context.Background(), "user-1", "tenant-1")

	result := auth.Authenticate(context.Background(), &service.AuthRequest{
		Token: session.Token,
	})

	if !result.Authorized {
		t.Errorf("Expected authorized, got error: %s", result.Error)
	}
}

func TestAuthService_Authenticate_NoCredentials(t *testing.T) {
	auth := service.NewAuthService(nil)

	result := auth.Authenticate(context.Background(), &service.AuthRequest{})

	if result.Authorized {
		t.Error("Expected not authorized with no credentials")
	}
}

func TestAuthService_GetStats(t *testing.T) {
	auth := service.NewAuthService(nil)

	auth.CreateAPIKey(context.Background(), &service.CreateAPIKeyRequest{
		Name:     "test-key",
		TenantID: "tenant-1",
	})
	auth.CreateSession(context.Background(), "user-1", "tenant-1")

	stats := auth.GetStats()

	if stats["api_keys_total"].(int) < 1 {
		t.Error("Expected at least 1 API key")
	}
	if stats["sessions_total"].(int) < 1 {
		t.Error("Expected at least 1 session")
	}
}

func TestSecretService_New(t *testing.T) {
	secret, err := service.NewSecretService("test-encryption-key-32-chars")
	if err != nil {
		t.Fatalf("Failed to create secret service: %v", err)
	}
	if secret == nil {
		t.Fatal("Expected non-nil secret service")
	}
}

func TestSecretService_CreateSecret(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	sec, err := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:        "db-password",
		Description: "Database password",
		Value:       "super-secret-password",
		TenantID:    "tenant-1",
		CreatedBy:   "user-1",
	})

	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	if sec.Name != "db-password" {
		t.Errorf("Expected name 'db-password', got '%s'", sec.Name)
	}
	if sec.Value != "" {
		t.Error("Expected value to be hidden in response")
	}
}

func TestSecretService_CreateSecret_MissingName(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	_, err := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Value:    "test",
		TenantID: "tenant-1",
	})

	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestSecretService_CreateSecret_Duplicate(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "duplicate",
		Value:    "test",
		TenantID: "tenant-1",
	})

	_, err := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "duplicate",
		Value:    "test",
		TenantID: "tenant-1",
	})

	if err == nil {
		t.Error("Expected error for duplicate secret")
	}
}

func TestSecretService_GetSecret(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	created, _ := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "test-secret",
		Value:    "secret-value",
		TenantID: "tenant-1",
	})

	// Get without decrypting
	retrieved, err := secret.GetSecret(context.Background(), created.ID, false)
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if retrieved.Value != "" {
		t.Error("Expected empty value when not decrypting")
	}

	// Get with decrypting
	decrypted, err := secret.GetSecret(context.Background(), created.ID, true)
	if err != nil {
		t.Fatalf("Failed to get decrypted secret: %v", err)
	}

	if decrypted.Value != "secret-value" {
		t.Errorf("Expected 'secret-value', got '%s'", decrypted.Value)
	}
}

func TestSecretService_GetSecretByName(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "named-secret",
		Value:    "named-value",
		TenantID: "tenant-1",
	})

	retrieved, err := secret.GetSecretByName(context.Background(), "tenant-1", "named-secret", true)
	if err != nil {
		t.Fatalf("Failed to get secret by name: %v", err)
	}

	if retrieved.Value != "named-value" {
		t.Errorf("Expected 'named-value', got '%s'", retrieved.Value)
	}
}

func TestSecretService_GetSecret_NotFound(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	_, err := secret.GetSecret(context.Background(), "nonexistent", true)
	if err == nil {
		t.Error("Expected error for nonexistent secret")
	}
}

func TestSecretService_UpdateSecret(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	created, _ := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "update-test",
		Value:    "old-value",
		TenantID: "tenant-1",
	})

	updated, err := secret.UpdateSecret(context.Background(), created.ID, &service.SecretUpdateRequest{
		Description: "Updated description",
		Value:       "new-value",
	})

	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	if updated.Description != "Updated description" {
		t.Error("Description not updated")
	}

	// Verify new value
	decrypted, _ := secret.GetSecret(context.Background(), created.ID, true)
	if decrypted.Value != "new-value" {
		t.Errorf("Expected 'new-value', got '%s'", decrypted.Value)
	}
}

func TestSecretService_DeleteSecret(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	created, _ := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "delete-test",
		Value:    "test",
		TenantID: "tenant-1",
	})

	err := secret.DeleteSecret(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	_, err = secret.GetSecret(context.Background(), created.ID, false)
	if err == nil {
		t.Error("Expected error for deleted secret")
	}
}

func TestSecretService_ListSecrets(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	for i := 0; i < 3; i++ {
		secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
			Name:     string(rune('a' + i)),
			Value:    "test",
			TenantID: "tenant-1",
		})
	}

	secrets := secret.ListSecrets("tenant-1", "")
	if len(secrets) < 3 {
		t.Errorf("Expected at least 3 secrets, got %d", len(secrets))
	}
}

func TestSecretService_GetAuditLog(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	created, _ := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "audit-test",
		Value:    "test",
		TenantID: "tenant-1",
	})
	secret.GetSecret(context.Background(), created.ID, true)

	log := secret.GetAuditLog(created.ID)
	if len(log) < 2 {
		t.Errorf("Expected at least 2 audit entries, got %d", len(log))
	}
}

func TestSecretService_GetStats(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:     "stats-test",
		Value:    "test",
		TenantID: "tenant-1",
	})

	stats := secret.GetStats()

	if stats["total_secrets"].(int) < 1 {
		t.Error("Expected at least 1 secret")
	}
}

func TestSecretService_Expiration(t *testing.T) {
	secret, _ := service.NewSecretService("test-encryption-key-32-chars")

	pastTime := time.Now().Add(-time.Hour)
	created, _ := secret.CreateSecret(context.Background(), &service.SecretCreateRequest{
		Name:      "expired-secret",
		Value:     "test",
		TenantID:  "tenant-1",
		ExpiresAt: &pastTime,
	})

	_, err := secret.GetSecret(context.Background(), created.ID, true)
	if err == nil {
		t.Error("Expected error for expired secret")
	}
}