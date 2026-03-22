package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// AuthService handles authentication and authorization
type AuthService struct {
	mu          sync.RWMutex
	apiKeys     map[string]*APIKeyAuth
	sessions    map[string]*Session
	permissions map[string][]string
	jwtSecret   string
}

// APIKeyAuth represents an API key for auth service
type APIKeyAuth struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	TenantID    string            `json:"tenant_id"`
	UserID      string            `json:"user_id"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	RateLimit   int               `json:"rate_limit"` // requests per minute
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
	Enabled      bool             `json:"enabled"`
}

// Session represents a user session
type Session struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	TenantID     string                 `json:"tenant_id"`
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refresh_token"`
	Data         map[string]interface{} `json:"data"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	LastActivity time.Time              `json:"last_activity"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	APIKey    string `json:"api_key"`
	Token     string `json:"token"`
	TenantID  string `json:"tenant_id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
}

// AuthResult represents an authentication result
type AuthResult struct {
	Authorized bool     `json:"authorized"`
	UserID     string   `json:"user_id,omitempty"`
	TenantID   string   `json:"tenant_id,omitempty"`
	SessionID  string   `json:"session_id,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *AuthConfig) *AuthService {
	if cfg == nil {
		cfg = &AuthConfig{
			JWTSecret:     generateRandomString(32),
			TokenExpiry:   time.Hour * 24,
			RefreshExpiry: time.Hour * 24 * 7,
		}
	}

	return &AuthService{
		apiKeys:     make(map[string]*APIKeyAuth),
		sessions:    make(map[string]*Session),
		permissions: make(map[string][]string),
		jwtSecret:   cfg.JWTSecret,
	}
}

// CreateAPIKey creates a new API key
func (s *AuthService) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyAuth, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	key := generateAPIKey()
	now := time.Now()

	apiKey := &APIKeyAuth{
		ID:          generateID(),
		Key:         key,
		Name:        req.Name,
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Permissions: req.Permissions,
		Metadata:    req.Metadata,
		RateLimit:   req.RateLimit,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   now,
		Enabled:     true,
	}

	s.mu.Lock()
	s.apiKeys[key] = apiKey
	s.mu.Unlock()

	return apiKey, nil
}

// CreateAPIKeyRequest for creating API keys
type CreateAPIKeyRequest struct {
	Name        string            `json:"name"`
	TenantID    string            `json:"tenant_id"`
	UserID      string            `json:"user_id"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	RateLimit   int               `json:"rate_limit"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// ValidateAPIKey validates an API key
func (s *AuthService) ValidateAPIKey(ctx context.Context, key string) (*APIKeyAuth, error) {
	s.mu.RLock()
	apiKey, exists := s.apiKeys[key]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid API key")
	}

	if !apiKey.Enabled {
		return nil, errors.New("API key is disabled")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New("API key has expired")
	}

	// Update last used
	now := time.Now()
	s.mu.Lock()
	apiKey.LastUsedAt = &now
	s.mu.Unlock()

	return apiKey, nil
}

// RevokeAPIKey revokes an API key
func (s *AuthService) RevokeAPIKey(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.apiKeys {
		if v.ID == keyID {
			delete(s.apiKeys, k)
			return nil
		}
	}

	return errors.New("API key not found")
}

// ListAPIKeys lists all API keys for a tenant
func (s *AuthService) ListAPIKeys(tenantID string) []*APIKeyAuth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*APIKeyAuth, 0)
	for _, k := range s.apiKeys {
		if tenantID == "" || k.TenantID == tenantID {
			// Don't expose the actual key
			kCopy := *k
			kCopy.Key = maskAPIKey(k.Key)
			keys = append(keys, &kCopy)
		}
	}
	return keys
}

// CreateSession creates a new session
func (s *AuthService) CreateSession(ctx context.Context, userID, tenantID string) (*Session, error) {
	session := &Session{
		ID:           generateID(),
		UserID:       userID,
		TenantID:     tenantID,
		Token:        generateToken(s.jwtSecret),
		RefreshToken: generateRefreshToken(),
		Data:         make(map[string]interface{}),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour * 24),
		LastActivity: time.Now(),
	}

	s.mu.Lock()
	s.sessions[session.Token] = session
	s.mu.Unlock()

	return session, nil
}

// ValidateSession validates a session token
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*Session, error) {
	s.mu.RLock()
	session, exists := s.sessions[token]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
		return nil, errors.New("session expired")
	}

	// Update last activity
	s.mu.Lock()
	session.LastActivity = time.Now()
	s.mu.Unlock()

	return session, nil
}

// RefreshSession refreshes a session
func (s *AuthService) RefreshSession(ctx context.Context, refreshToken string) (*Session, error) {
	s.mu.RLock()
	var foundSession *Session
	for _, s := range s.sessions {
		if s.RefreshToken == refreshToken {
			foundSession = s
			break
		}
	}
	s.mu.RUnlock()

	if foundSession == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new tokens
	s.mu.Lock()
	delete(s.sessions, foundSession.Token)
	foundSession.Token = generateToken(s.jwtSecret)
	foundSession.RefreshToken = generateRefreshToken()
	foundSession.ExpiresAt = time.Now().Add(time.Hour * 24)
	s.sessions[foundSession.Token] = foundSession
	s.mu.Unlock()

	return foundSession, nil
}

// RevokeSession revokes a session
func (s *AuthService) RevokeSession(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, token)
	return nil
}

// CheckPermission checks if a user has permission for an action
func (s *AuthService) CheckPermission(ctx context.Context, userID, action, resource string) bool {
	s.mu.RLock()
	permissions, exists := s.permissions[userID]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	// Check for wildcard permission
	for _, p := range permissions {
		if p == "*" || p == action || p == action+":"+resource {
			return true
		}
	}

	return false
}

// GrantPermission grants permissions to a user
func (s *AuthService) GrantPermission(ctx context.Context, userID string, permissions []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.permissions[userID] = append(s.permissions[userID], permissions...)
	return nil
}

// RevokePermission revokes permissions from a user
func (s *AuthService) RevokePermission(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.permissions, userID)
	return nil
}

// Authenticate authenticates a request
func (s *AuthService) Authenticate(ctx context.Context, req *AuthRequest) *AuthResult {
	// Try API key first
	if req.APIKey != "" {
		apiKey, err := s.ValidateAPIKey(ctx, req.APIKey)
		if err != nil {
			return &AuthResult{
				Authorized: false,
				Error:      err.Error(),
			}
		}

		return &AuthResult{
			Authorized:  true,
			UserID:      apiKey.UserID,
			TenantID:    apiKey.TenantID,
			Permissions: apiKey.Permissions,
		}
	}

	// Try session token
	if req.Token != "" {
		session, err := s.ValidateSession(ctx, req.Token)
		if err != nil {
			return &AuthResult{
				Authorized: false,
				Error:      err.Error(),
			}
		}

		return &AuthResult{
			Authorized: true,
			UserID:     session.UserID,
			TenantID:   session.TenantID,
			SessionID:  session.ID,
		}
	}

	return &AuthResult{
		Authorized: false,
		Error:      "no credentials provided",
	}
}

// GetStats returns auth service statistics
func (s *AuthService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeSessions := 0
	for _, s := range s.sessions {
		if time.Now().Before(s.ExpiresAt) {
			activeSessions++
		}
	}

	return map[string]interface{}{
		"api_keys_total":   len(s.apiKeys),
		"sessions_total":   len(s.sessions),
		"active_sessions":  activeSessions,
		"users_with_perms": len(s.permissions),
	}
}

// Helper functions

func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "cp_" + base64.URLEncoding.EncodeToString(b)
}

func generateToken(secret string) string {
	b := make([]byte, 32)
	rand.Read(b)
	h := sha256.New()
	h.Write(b)
	h.Write([]byte(secret))
	h.Write([]byte(time.Now().String()))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "ref_" + base64.URLEncoding.EncodeToString(b)
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func maskAPIKey(key string) string {
	if len(key) < 12 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}

// ToJSON serializes API key to JSON
func (k *APIKeyAuth) ToJSON() ([]byte, error) {
	return json.Marshal(k)
}

// ToJSON serializes session to JSON
func (s *Session) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}