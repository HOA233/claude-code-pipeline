package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ConfigService manages application configuration
type ConfigService struct {
	mu      sync.RWMutex
	configs map[string]*Config
	profiles map[string]map[string]*Config
	watchers map[string][]ConfigWatcher
	history map[string][]ConfigHistory
}

// Config represents a configuration entry
type Config struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"` // string, int, bool, json, yaml
	Profile     string      `json:"profile"`
	ProjectID   string      `json:"project_id,omitempty"`
	Description string      `json:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Version     int         `json:"version"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CreatedBy   string      `json:"created_by,omitempty"`
	UpdatedBy   string      `json:"updated_by,omitempty"`
}

// ConfigHistory represents a configuration change history entry
type ConfigHistory struct {
	Version   int         `json:"version"`
	Value     interface{} `json:"value"`
	ChangedAt time.Time   `json:"changed_at"`
	ChangedBy string      `json:"changed_by"`
	Reason    string      `json:"reason,omitempty"`
}

// ConfigWatcher is a callback for config changes
type ConfigWatcher func(key string, oldValue, newValue interface{})

// ConfigSetRequest for setting configuration
type ConfigSetRequest struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Profile     string      `json:"profile"`
	ProjectID   string      `json:"project_id,omitempty"`
	Description string      `json:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	User        string      `json:"user,omitempty"`
	Reason      string      `json:"reason,omitempty"`
}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	return &ConfigService{
		configs:  make(map[string]*Config),
		profiles: make(map[string]map[string]*Config),
		watchers: make(map[string][]ConfigWatcher),
		history:  make(map[string][]ConfigHistory),
	}
}

// Set sets a configuration value
func (s *ConfigService) Set(ctx context.Context, req *ConfigSetRequest) (*Config, error) {
	if req.Key == "" {
		return nil, errors.New("key is required")
	}

	profile := req.Profile
	if profile == "" {
		profile = "default"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing config
	profileKey := s.getProfileKey(req.Key, profile, req.ProjectID)
	existing, exists := s.configs[profileKey]

	now := time.Now()
	config := &Config{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Profile:     profile,
		ProjectID:   req.ProjectID,
		Description: req.Description,
		Tags:        req.Tags,
		UpdatedAt:   now,
		UpdatedBy:   req.User,
	}

	if exists {
		config.Version = existing.Version + 1
		config.CreatedAt = existing.CreatedAt
		config.CreatedBy = existing.CreatedBy
	} else {
		config.Version = 1
		config.CreatedAt = now
		config.CreatedBy = req.User
	}

	s.configs[profileKey] = config

	// Update profile map
	if s.profiles[profile] == nil {
		s.profiles[profile] = make(map[string]*Config)
	}
	s.profiles[profile][req.Key] = config

	// Add history entry
	s.history[profileKey] = append(s.history[profileKey], ConfigHistory{
		Version:   config.Version,
		Value:     config.Value,
		ChangedAt: now,
		ChangedBy: req.User,
		Reason:    req.Reason,
	})

	// Notify watchers
	if watchers, ok := s.watchers[req.Key]; ok {
		oldValue := interface{}(nil)
		if exists {
			oldValue = existing.Value
		}
		for _, w := range watchers {
			go w(req.Key, oldValue, config.Value)
		}
	}

	return config, nil
}

// Get gets a configuration value
func (s *ConfigService) Get(ctx context.Context, key, profile, projectID string) (*Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if profile == "" {
		profile = "default"
	}

	profileKey := s.getProfileKey(key, profile, projectID)
	config, exists := s.configs[profileKey]
	if !exists {
		return nil, fmt.Errorf("config not found: %s", key)
	}

	return config, nil
}

// GetValue gets a configuration value as a specific type
func (s *ConfigService) GetValue(ctx context.Context, key, profile, projectID string) (interface{}, error) {
	config, err := s.Get(ctx, key, profile, projectID)
	if err != nil {
		return nil, err
	}
	return config.Value, nil
}

// GetString gets a configuration value as string
func (s *ConfigService) GetString(ctx context.Context, key, profile, projectID string) (string, error) {
	value, err := s.GetValue(ctx, key, profile, projectID)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Sprintf("%v", value), nil
	}
	return str, nil
}

// GetInt gets a configuration value as int
func (s *ConfigService) GetInt(ctx context.Context, key, profile, projectID string) (int, error) {
	value, err := s.GetValue(ctx, key, profile, projectID)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// GetBool gets a configuration value as bool
func (s *ConfigService) GetBool(ctx context.Context, key, profile, projectID string) (bool, error) {
	value, err := s.GetValue(ctx, key, profile, projectID)
	if err != nil {
		return false, err
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
	return b, nil
}

// GetJSON gets a configuration value as JSON
func (s *ConfigService) GetJSON(ctx context.Context, key, profile, projectID string, target interface{}) error {
	value, err := s.GetValue(ctx, key, profile, projectID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// Delete deletes a configuration
func (s *ConfigService) Delete(ctx context.Context, key, profile, projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if profile == "" {
		profile = "default"
	}

	profileKey := s.getProfileKey(key, profile, projectID)
	if _, exists := s.configs[profileKey]; !exists {
		return fmt.Errorf("config not found: %s", key)
	}

	delete(s.configs, profileKey)
	if s.profiles[profile] != nil {
		delete(s.profiles[profile], key)
	}

	return nil
}

// List lists all configurations for a profile
func (s *ConfigService) List(profile, projectID string) []*Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if profile == "" {
		profile = "default"
	}

	configs := make([]*Config, 0)
	for key, config := range s.configs {
		if projectID != "" && config.ProjectID != projectID {
			continue
		}
		if config.Profile == profile {
			configs = append(configs, config)
		}
		_ = key
	}

	return configs
}

// ListProfiles lists all profiles
func (s *ConfigService) ListProfiles() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := make([]string, 0, len(s.profiles))
	for profile := range s.profiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

// Watch registers a watcher for configuration changes
func (s *ConfigService) Watch(key string, watcher ConfigWatcher) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.watchers[key] = append(s.watchers[key], watcher)
}

// GetHistory gets the change history for a configuration
func (s *ConfigService) GetHistory(key, profile, projectID string) []ConfigHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if profile == "" {
		profile = "default"
	}

	profileKey := s.getProfileKey(key, profile, projectID)
	return s.history[profileKey]
}

// Rollback rolls back a configuration to a previous version
func (s *ConfigService) Rollback(ctx context.Context, key, profile, projectID string, version int, user string) (*Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if profile == "" {
		profile = "default"
	}

	profileKey := s.getProfileKey(key, profile, projectID)
	history := s.history[profileKey]

	var targetHistory *ConfigHistory
	for i := range history {
		if history[i].Version == version {
			targetHistory = &history[i]
			break
		}
	}

	if targetHistory == nil {
		return nil, fmt.Errorf("version %d not found", version)
	}

	existing := s.configs[profileKey]
	now := time.Now()

	config := &Config{
		Key:         key,
		Value:       targetHistory.Value,
		Type:        existing.Type,
		Profile:     profile,
		ProjectID:   projectID,
		Description: existing.Description,
		Tags:        existing.Tags,
		Version:     existing.Version + 1,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   now,
		CreatedBy:   existing.CreatedBy,
		UpdatedBy:   user,
	}

	s.configs[profileKey] = config
	s.profiles[profile][key] = config

	// Add rollback history entry
	s.history[profileKey] = append(s.history[profileKey], ConfigHistory{
		Version:   config.Version,
		Value:     config.Value,
		ChangedAt: now,
		ChangedBy: user,
		Reason:    fmt.Sprintf("Rollback to version %d", version),
	})

	return config, nil
}

// Export exports all configurations for a profile
func (s *ConfigService) Export(profile, projectID string) ([]byte, error) {
	configs := s.List(profile, projectID)
	return json.MarshalIndent(configs, "", "  ")
}

// Import imports configurations from JSON
func (s *ConfigService) Import(ctx context.Context, data []byte, profile, user string, overwrite bool) error {
	var configs []*Config
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	for _, config := range configs {
		if profile != "" {
			config.Profile = profile
		}

		existing, err := s.Get(ctx, config.Key, config.Profile, config.ProjectID)
		if err == nil && !overwrite {
			continue // Skip existing if not overwriting
		}

		req := &ConfigSetRequest{
			Key:         config.Key,
			Value:       config.Value,
			Type:        config.Type,
			Profile:     config.Profile,
			ProjectID:   config.ProjectID,
			Description: config.Description,
			Tags:        config.Tags,
			User:        user,
		}

		if existing != nil {
			req.Reason = "Import (overwrite)"
		} else {
			req.Reason = "Import"
		}

		_, err = s.Set(ctx, req)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetStats returns config service statistics
func (s *ConfigService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalHistory := 0
	for _, h := range s.history {
		totalHistory += len(h)
	}

	return map[string]interface{}{
		"total_configs":  len(s.configs),
		"total_profiles": len(s.profiles),
		"total_watchers": len(s.watchers),
		"total_history":  totalHistory,
	}
}

func (s *ConfigService) getProfileKey(key, profile, projectID string) string {
	if projectID != "" {
		return fmt.Sprintf("%s:%s:%s", profile, projectID, key)
	}
	return fmt.Sprintf("%s:%s", profile, key)
}

// ToJSON serializes a config to JSON
func (c *Config) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}