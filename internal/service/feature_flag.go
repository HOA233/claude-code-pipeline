package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// FeatureFlagService manages feature flags/toggles
type FeatureFlagService struct {
	mu     sync.RWMutex
	flags  map[string]*FeatureFlag
	rules  map[string][]TargetingRule
	stats  map[string]*FlagStats
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Enabled     bool              `json:"enabled"`
	Type        FlagType          `json:"type"` // boolean, multivariate
	Variations  []Variation       `json:"variations,omitempty"`
	Default     interface{}       `json:"default"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FlagType represents flag type
type FlagType string

const (
	FlagTypeBoolean     FlagType = "boolean"
	FlagTypeMultivariate FlagType = "multivariate"
)

// Variation represents a flag variation
type Variation struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// TargetingRule represents a targeting rule
type TargetingRule struct {
	ID         string        `json:"id"`
	FlagKey    string        `json:"flag_key"`
	Conditions []Condition   `json:"conditions"`
	Variation  string        `json:"variation"`
	Priority   int           `json:"priority"`
	Enabled    bool          `json:"enabled"`
}

// Condition represents a targeting condition
type Condition struct {
	Attribute string      `json:"attribute"`
	Operator  string      `json:"operator"` // eq, ne, in, contains, gt, lt, gte, lte
	Value     interface{} `json:"value"`
}

// FlagStats represents flag evaluation statistics
type FlagStats struct {
	Key          string    `json:"key"`
	Evaluations  int64     `json:"evaluations"`
	TrueCount    int64     `json:"true_count"`
	FalseCount   int64     `json:"false_count"`
	LastError    string    `json:"last_error,omitempty"`
	LastEvaluated time.Time `json:"last_evaluated"`
}

// EvaluationContext represents evaluation context
type EvaluationContext struct {
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EvaluationResult represents evaluation result
type EvaluationResult struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Variation string      `json:"variation,omitempty"`
	Reason    string      `json:"reason"`
}

// NewFeatureFlagService creates a new feature flag service
func NewFeatureFlagService() *FeatureFlagService {
	return &FeatureFlagService{
		flags: make(map[string]*FeatureFlag),
		rules: make(map[string][]TargetingRule),
		stats: make(map[string]*FlagStats),
	}
}

// CreateFlag creates a new feature flag
func (s *FeatureFlagService) CreateFlag(ctx context.Context, flag *FeatureFlag) error {
	if flag.Key == "" {
		return errors.New("key is required")
	}
	if flag.Name == "" {
		return errors.New("name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key already exists
	for _, existing := range s.flags {
		if existing.Key == flag.Key {
			return fmt.Errorf("flag with key '%s' already exists", flag.Key)
		}
	}

	now := time.Now()
	if flag.ID == "" {
		flag.ID = generateID()
	}
	flag.CreatedAt = now
	flag.UpdatedAt = now

	// Set default variations for boolean flags
	if flag.Type == FlagTypeBoolean && len(flag.Variations) == 0 {
		flag.Variations = []Variation{
			{ID: "true", Name: "True", Value: true},
			{ID: "false", Name: "False", Value: false},
		}
	}

	s.flags[flag.ID] = flag
	s.stats[flag.Key] = &FlagStats{Key: flag.Key}

	return nil
}

// GetFlag gets a flag by ID or key
func (s *FeatureFlagService) GetFlag(idOrKey string) (*FeatureFlag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try by ID first
	if flag, exists := s.flags[idOrKey]; exists {
		return flag, nil
	}

	// Try by key
	for _, flag := range s.flags {
		if flag.Key == idOrKey {
			return flag, nil
		}
	}

	return nil, fmt.Errorf("flag not found: %s", idOrKey)
}

// UpdateFlag updates a feature flag
func (s *FeatureFlagService) UpdateFlag(ctx context.Context, id string, updates map[string]interface{}) (*FeatureFlag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	flag, exists := s.flags[id]
	if !exists {
		return nil, fmt.Errorf("flag not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		flag.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		flag.Description = desc
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		flag.Enabled = enabled
	}
	if tags, ok := updates["tags"].([]string); ok {
		flag.Tags = tags
	}
	if defaultVal, ok := updates["default"]; ok {
		flag.Default = defaultVal
	}

	flag.UpdatedAt = time.Now()

	return flag, nil
}

// DeleteFlag deletes a feature flag
func (s *FeatureFlagService) DeleteFlag(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	flag, exists := s.flags[id]
	if !exists {
		return fmt.Errorf("flag not found: %s", id)
	}

	delete(s.flags, id)
	delete(s.rules, flag.Key)
	delete(s.stats, flag.Key)

	return nil
}

// ToggleFlag toggles a flag on/off
func (s *FeatureFlagService) ToggleFlag(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	flag, exists := s.flags[id]
	if !exists {
		return fmt.Errorf("flag not found: %s", id)
	}

	flag.Enabled = enabled
	flag.UpdatedAt = time.Now()

	return nil
}

// Evaluate evaluates a flag for a given context
func (s *FeatureFlagService) Evaluate(ctx context.Context, key string, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find flag
	var flag *FeatureFlag
	for _, f := range s.flags {
		if f.Key == key {
			flag = f
			break
		}
	}

	if flag == nil {
		return nil, fmt.Errorf("flag not found: %s", key)
	}

	// Update stats
	stats := s.stats[key]
	stats.Evaluations++
	stats.LastEvaluated = time.Now()

	result := &EvaluationResult{
		Key:    key,
		Reason: "default",
	}

	// Check if flag is enabled
	if !flag.Enabled {
		stats.FalseCount++
		result.Value = flag.Default
		result.Reason = "flag_disabled"
		return result, nil
	}

	// Check targeting rules
	rules := s.rules[key]
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if s.matchesConditions(rule.Conditions, evalCtx) {
			// Find variation
			for _, v := range flag.Variations {
				if v.ID == rule.Variation {
					stats.TrueCount++
					result.Value = v.Value
					result.Variation = v.ID
					result.Reason = "targeting_rule"
					return result, nil
				}
			}
		}
	}

	// Return default
	stats.FalseCount++
	result.Value = flag.Default
	result.Reason = "default"

	return result, nil
}

// EvaluateAll evaluates all flags for a context
func (s *FeatureFlagService) EvaluateAll(ctx context.Context, evalCtx *EvaluationContext) map[string]*EvaluationResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make(map[string]*EvaluationResult)
	for _, flag := range s.flags {
		// Skip locked evaluation for batch
		result := &EvaluationResult{
			Key:    flag.Key,
			Reason: "default",
		}

		if !flag.Enabled {
			result.Value = flag.Default
			result.Reason = "flag_disabled"
		} else {
			// Check rules
			matched := false
			for _, rule := range s.rules[flag.Key] {
				if rule.Enabled && s.matchesConditions(rule.Conditions, evalCtx) {
					for _, v := range flag.Variations {
						if v.ID == rule.Variation {
							result.Value = v.Value
							result.Variation = v.ID
							result.Reason = "targeting_rule"
							matched = true
							break
						}
					}
					if matched {
						break
					}
				}
			}

			if !matched {
				result.Value = flag.Default
			}
		}

		results[flag.Key] = result
	}

	return results
}

// AddTargetingRule adds a targeting rule
func (s *FeatureFlagService) AddTargetingRule(rule *TargetingRule) error {
	if rule.FlagKey == "" {
		return errors.New("flag_key is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if rule.ID == "" {
		rule.ID = generateID()
	}

	s.rules[rule.FlagKey] = append(s.rules[rule.FlagKey], *rule)
	return nil
}

// RemoveTargetingRule removes a targeting rule
func (s *FeatureFlagService) RemoveTargetingRule(flagKey, ruleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rules := s.rules[flagKey]
	for i, r := range rules {
		if r.ID == ruleID {
			s.rules[flagKey] = append(rules[:i], rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("rule not found: %s", ruleID)
}

// ListFlags lists all flags
func (s *FeatureFlagService) ListFlags(tag string) []*FeatureFlag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flags := make([]*FeatureFlag, 0)
	for _, flag := range s.flags {
		if tag == "" || containsStr(flag.Tags, tag) {
			flags = append(flags, flag)
		}
	}
	return flags
}

// GetStats gets flag statistics
func (s *FeatureFlagService) GetStats(key string) (*FlagStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats, exists := s.stats[key]
	if !exists {
		return nil, fmt.Errorf("stats not found for flag: %s", key)
	}
	return stats, nil
}

// matchesConditions checks if context matches conditions
func (s *FeatureFlagService) matchesConditions(conditions []Condition, ctx *EvaluationContext) bool {
	if ctx == nil {
		return false
	}

	for _, cond := range conditions {
		var value interface{}
		switch cond.Attribute {
		case "user_id":
			value = ctx.UserID
		case "session_id":
			value = ctx.SessionID
		case "tenant_id":
			value = ctx.TenantID
		default:
			if ctx.Attributes != nil {
				value = ctx.Attributes[cond.Attribute]
			}
		}

		if !s.evaluateCondition(cond.Operator, value, cond.Value) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (s *FeatureFlagService) evaluateCondition(operator string, actual, expected interface{}) bool {
	switch operator {
	case "eq":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	case "ne":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected)
	case "in":
		if arr, ok := expected.([]interface{}); ok {
			for _, v := range arr {
				if fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", v) {
					return true
				}
			}
		}
		return false
	case "contains":
		return containsStr(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "gt", "lt", "gte", "lte":
		// Numeric comparisons
		actFloat, ok1 := toFloat64(actual)
		expFloat, ok2 := toFloat64(expected)
		if !ok1 || !ok2 {
			return false
		}
		switch operator {
		case "gt":
			return actFloat > expFloat
		case "lt":
			return actFloat < expFloat
		case "gte":
			return actFloat >= expFloat
		case "lte":
			return actFloat <= expFloat
		}
	}
	return false
}

// ToJSON serializes flag to JSON
func (f *FeatureFlag) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

// Helper functions
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}