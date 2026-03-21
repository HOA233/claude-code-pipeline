package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key,omitempty"` // Only returned on creation
	KeyHash     string    `json:"-"`
	UserID      string    `json:"user_id"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	Enabled     bool      `json:"enabled"`
	RateLimit   int       `json:"rate_limit"` // Requests per minute
}

// APIKeyService manages API keys
type APIKeyService struct {
	redis *repository.RedisClient
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(redis *repository.RedisClient) *APIKeyService {
	return &APIKeyService{redis: redis}
}

// CreateKey creates a new API key
func (s *APIKeyService) CreateKey(ctx context.Context, name, userID string, roles, permissions []string, expiresAt time.Time, rateLimit int) (*APIKey, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	key := "pk_" + hex.EncodeToString(keyBytes)

	// Generate key ID
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}
	id := "key_" + hex.EncodeToString(idBytes)

	apiKey := &APIKey{
		ID:          id,
		Name:        name,
		Key:         key,
		KeyHash:     hashKey(key),
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		Enabled:     true,
		RateLimit:   rateLimit,
	}

	// Save to Redis
	if err := s.saveKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

// GetKey retrieves an API key by ID
func (s *APIKeyService) GetKey(ctx context.Context, id string) (*APIKey, error) {
	data, err := s.redis.GetCache(ctx, "apikey:"+id)
	if err != nil {
		return nil, fmt.Errorf("key not found: %w", err)
	}

	var key APIKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	return &key, nil
}

// ValidateKey validates an API key and returns its details
func (s *APIKeyService) ValidateKey(ctx context.Context, key string) (*APIKey, error) {
	if !strings.HasPrefix(key, "pk_") {
		return nil, fmt.Errorf("invalid key format")
	}

	keyHash := hashKey(key)

	// Look up key by hash
	keys, err := s.redis.ListCacheKeys(ctx, "apikey:*")
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	for _, k := range keys {
		apiKey, err := s.GetKey(ctx, strings.TrimPrefix(k, "apikey:"))
		if err != nil {
			continue
		}

		if apiKey.KeyHash == keyHash {
			// Check if key is enabled
			if !apiKey.Enabled {
				return nil, fmt.Errorf("key is disabled")
			}

			// Check expiration
			if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
				return nil, fmt.Errorf("key has expired")
			}

			// Update last used
			apiKey.LastUsedAt = time.Now()
			_ = s.saveKey(ctx, apiKey)

			return apiKey, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// RevokeKey revokes an API key
func (s *APIKeyService) RevokeKey(ctx context.Context, id string) error {
	key, err := s.GetKey(ctx, id)
	if err != nil {
		return err
	}

	key.Enabled = false
	return s.saveKey(ctx, key)
}

// DeleteKey permanently deletes an API key
func (s *APIKeyService) DeleteKey(ctx context.Context, id string) error {
	return s.redis.DeleteCache(ctx, "apikey:"+id)
}

// ListKeys lists all API keys for a user
func (s *APIKeyService) ListKeys(ctx context.Context, userID string) ([]*APIKey, error) {
	keys, err := s.redis.ListCacheKeys(ctx, "apikey:*")
	if err != nil {
		return nil, err
	}

	var apiKeys []*APIKey
	for _, k := range keys {
		key, err := s.GetKey(ctx, strings.TrimPrefix(k, "apikey:"))
		if err != nil {
			continue
		}

		if userID == "" || key.UserID == userID {
			// Don't return the key hash
			key.KeyHash = ""
			apiKeys = append(apiKeys, key)
		}
	}

	return apiKeys, nil
}

// HasPermission checks if a key has a specific permission
func (k *APIKey) HasPermission(permission string) bool {
	for _, p := range k.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// HasRole checks if a key has a specific role
func (k *APIKey) HasRole(role string) bool {
	for _, r := range k.Roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}

func (s *APIKeyService) saveKey(ctx context.Context, key *APIKey) error {
	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}

	// Set TTL based on expiration
	ttl := time.Duration(0)
	if !key.ExpiresAt.IsZero() {
		ttl = time.Until(key.ExpiresAt)
	}

	return s.redis.SetCache(ctx, "apikey:"+key.ID, data, ttl)
}

func hashKey(key string) string {
	// Simple hash for key comparison
	// In production, use bcrypt or similar
	return hex.EncodeToString([]byte(key)[:16])
}