package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// SecretService manages secrets and sensitive configuration
type SecretService struct {
	mu           sync.RWMutex
	secrets      map[string]*Secret
	encryptedKey []byte
	version      int
	auditLog     []SecretAuditEntry
}

// Secret represents a stored secret
type Secret struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Value       string            `json:"value,omitempty"` // Encrypted value
	TenantID    string            `json:"tenant_id"`
	ProjectID   string            `json:"project_id,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Version     int               `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedBy   string            `json:"created_by"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
}

// SecretAuditEntry represents an audit log entry
type SecretAuditEntry struct {
	ID        string    `json:"id"`
	SecretID  string    `json:"secret_id"`
	Action    string    `json:"action"` // create, read, update, delete
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SecretCreateRequest for creating secrets
type SecretCreateRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Value       string            `json:"value"`
	TenantID    string            `json:"tenant_id"`
	ProjectID   string            `json:"project_id,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedBy   string            `json:"created_by"`
}

// SecretUpdateRequest for updating secrets
type SecretUpdateRequest struct {
	Description string            `json:"description,omitempty"`
	Value       string            `json:"value,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// NewSecretService creates a new secret service
func NewSecretService(encryptionKey string) (*SecretService, error) {
	if len(encryptionKey) < 32 {
		encryptionKey = encryptionKey + string(make([]byte, 32-len(encryptionKey)))
	}
	if len(encryptionKey) > 32 {
		encryptionKey = encryptionKey[:32]
	}

	return &SecretService{
		secrets:      make(map[string]*Secret),
		encryptedKey: []byte(encryptionKey),
		version:      1,
		auditLog:     make([]SecretAuditEntry, 0),
	}, nil
}

// CreateSecret creates a new secret
func (s *SecretService) CreateSecret(ctx context.Context, req *SecretCreateRequest) (*Secret, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Value == "" {
		return nil, errors.New("value is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Check if secret already exists
	s.mu.RLock()
	for _, sec := range s.secrets {
		if sec.Name == req.Name && sec.TenantID == req.TenantID {
			s.mu.RUnlock()
			return nil, fmt.Errorf("secret '%s' already exists", req.Name)
		}
	}
	s.mu.RUnlock()

	// Encrypt the value
	encryptedValue, err := s.encrypt(req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	now := time.Now()
	secret := &Secret{
		ID:          generateID(),
		Name:        req.Name,
		Description: req.Description,
		Value:       encryptedValue,
		TenantID:    req.TenantID,
		ProjectID:   req.ProjectID,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		Version:     s.version,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   req.ExpiresAt,
		CreatedBy:   req.CreatedBy,
	}

	s.mu.Lock()
	s.secrets[secret.ID] = secret
	s.auditLog = append(s.auditLog, SecretAuditEntry{
		ID:        generateID(),
		SecretID:  secret.ID,
		Action:    "create",
		UserID:    req.CreatedBy,
		Timestamp: now,
	})
	s.mu.Unlock()

	// Don't return the encrypted value
	secretCopy := *secret
	secretCopy.Value = ""

	return &secretCopy, nil
}

// GetSecret retrieves a secret by ID
func (s *SecretService) GetSecret(ctx context.Context, id string, decrypt bool) (*Secret, error) {
	s.mu.RLock()
	secret, exists := s.secrets[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("secret not found: %s", id)
	}

	// Check expiration
	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return nil, errors.New("secret has expired")
	}

	secretCopy := *secret

	if decrypt {
		decryptedValue, err := s.decrypt(secret.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret: %w", err)
		}
		secretCopy.Value = decryptedValue

		// Update last used
		now := time.Now()
		s.mu.Lock()
		secret.LastUsedAt = &now
		s.auditLog = append(s.auditLog, SecretAuditEntry{
			ID:        generateID(),
			SecretID:  secret.ID,
			Action:    "read",
			UserID:    ctx.Value("user_id").(string),
			Timestamp: now,
		})
		s.mu.Unlock()
	} else {
		secretCopy.Value = ""
	}

	return &secretCopy, nil
}

// GetSecretByName retrieves a secret by name
func (s *SecretService) GetSecretByName(ctx context.Context, tenantID, name string, decrypt bool) (*Secret, error) {
	s.mu.RLock()
	var found *Secret
	for _, secret := range s.secrets {
		if secret.Name == name && secret.TenantID == tenantID {
			found = secret
			break
		}
	}
	s.mu.RUnlock()

	if found == nil {
		return nil, fmt.Errorf("secret not found: %s", name)
	}

	return s.GetSecret(ctx, found.ID, decrypt)
}

// UpdateSecret updates a secret
func (s *SecretService) UpdateSecret(ctx context.Context, id string, req *SecretUpdateRequest) (*Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	secret, exists := s.secrets[id]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", id)
	}

	if req.Description != "" {
		secret.Description = req.Description
	}
	if req.Value != "" {
		encryptedValue, err := s.encrypt(req.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt secret: %w", err)
		}
		secret.Value = encryptedValue
		secret.Version++
	}
	if req.Tags != nil {
		secret.Tags = req.Tags
	}
	if req.Metadata != nil {
		secret.Metadata = req.Metadata
	}
	if req.ExpiresAt != nil {
		secret.ExpiresAt = req.ExpiresAt
	}

	secret.UpdatedAt = time.Now()

	s.auditLog = append(s.auditLog, SecretAuditEntry{
		ID:        generateID(),
		SecretID:  secret.ID,
		Action:    "update",
		UserID:    ctx.Value("user_id").(string),
		Timestamp: secret.UpdatedAt,
	})

	secretCopy := *secret
	secretCopy.Value = ""

	return &secretCopy, nil
}

// DeleteSecret deletes a secret
func (s *SecretService) DeleteSecret(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	secret, exists := s.secrets[id]
	if !exists {
		return fmt.Errorf("secret not found: %s", id)
	}

	s.auditLog = append(s.auditLog, SecretAuditEntry{
		ID:        generateID(),
		SecretID:  secret.ID,
		Action:    "delete",
		UserID:    ctx.Value("user_id").(string),
		Timestamp: time.Now(),
	})

	delete(s.secrets, id)
	return nil
}

// ListSecrets lists all secrets for a tenant
func (s *SecretService) ListSecrets(tenantID, projectID string) []*Secret {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secrets := make([]*Secret, 0)
	for _, secret := range s.secrets {
		if tenantID != "" && secret.TenantID != tenantID {
			continue
		}
		if projectID != "" && secret.ProjectID != projectID {
			continue
		}

		// Don't include the value
		secretCopy := *secret
		secretCopy.Value = ""
		secrets = append(secrets, &secretCopy)
	}
	return secrets
}

// GetAuditLog returns the audit log
func (s *SecretService) GetAuditLog(secretID string) []SecretAuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if secretID == "" {
		return s.auditLog
	}

	entries := make([]SecretAuditEntry, 0)
	for _, entry := range s.auditLog {
		if entry.SecretID == secretID {
			entries = append(entries, entry)
		}
	}
	return entries
}

// RotateKey rotates the encryption key
func (s *SecretService) RotateKey(ctx context.Context, newKey string) error {
	if len(newKey) < 32 {
		newKey = newKey + string(make([]byte, 32-len(newKey)))
	}
	if len(newKey) > 32 {
		newKey = newKey[:32]
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-encrypt all secrets with new key
	for id, secret := range s.secrets {
		// Decrypt with old key
		decryptedValue, err := s.decrypt(secret.Value)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret %s: %w", id, err)
		}

		// Store decrypted temporarily
		secret.Value = decryptedValue
	}

	// Update key
	s.encryptedKey = []byte(newKey)
	s.version++

	// Re-encrypt all secrets
	for id, secret := range s.secrets {
		encryptedValue, err := s.encrypt(secret.Value)
		if err != nil {
			return fmt.Errorf("failed to re-encrypt secret %s: %w", id, err)
		}
		secret.Value = encryptedValue
		secret.Version = s.version
	}

	return nil
}

// encrypt encrypts a value using AES-GCM
func (s *SecretService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a value using AES-GCM
func (s *SecretService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetStats returns secret service statistics
func (s *SecretService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expired := 0
	now := time.Now()
	for _, secret := range s.secrets {
		if secret.ExpiresAt != nil && now.After(*secret.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_secrets": len(s.secrets),
		"expired_secrets": expired,
		"audit_entries": len(s.auditLog),
		"version":       s.version,
	}
}

// ToJSON serializes a secret to JSON
func (s *Secret) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}