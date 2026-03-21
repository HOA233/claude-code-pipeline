package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TagService manages resource tags and labels
type TagService struct {
	mu     sync.RWMutex
	tags   map[string]*Tag
	bindings map[string][]TagBinding
}

// Tag represents a tag definition
type Tag struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Color       string    `json:"color"`
	ReadOnly    bool      `json:"read_only"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
	UsageCount  int       `json:"usage_count"`
}

// TagBinding represents a tag binding to a resource
type TagBinding struct {
	ID         string    `json:"id"`
	TagID      string    `json:"tag_id"`
	TagKey     string    `json:"tag_key"`
	TagValue   string    `json:"tag_value"`
	ResourceID string    `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	TenantID   string    `json:"tenant_id"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

// TagRule represents a tagging rule
type TagRule struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ResourceTypes []string `json:"resource_types"`
	RequiredTags []string `json:"required_tags"`
	AllowedValues map[string][]string `json:"allowed_values"`
	Enabled      bool     `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewTagService creates a new tag service
func NewTagService() *TagService {
	return &TagService{
		tags:     make(map[string]*Tag),
		bindings: make(map[string][]TagBinding),
	}
}

// CreateTag creates a new tag
func (s *TagService) CreateTag(tag *Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tag.Key == "" {
		return fmt.Errorf("key is required")
	}

	now := time.Now()
	if tag.ID == "" {
		tag.ID = generateID()
	}
	tag.CreatedAt = now
	tag.UsageCount = 0

	s.tags[tag.ID] = tag

	return nil
}

// GetTag gets a tag by ID
func (s *TagService) GetTag(id string) (*Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tag, ok := s.tags[id]
	if !ok {
		return nil, fmt.Errorf("tag not found: %s", id)
	}
	return tag, nil
}

// GetTagByKey gets a tag by key and value
func (s *TagService) GetTagByKey(key, value string) (*Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tag := range s.tags {
		if tag.Key == key && tag.Value == value {
			return tag, nil
		}
	}
	return nil, fmt.Errorf("tag not found: %s=%s", key, value)
}

// ListTags lists all tags
func (s *TagService) ListTags(category string) []*Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Tag
	for _, tag := range s.tags {
		if category == "" || tag.Category == category {
			results = append(results, tag)
		}
	}
	return results
}

// UpdateTag updates a tag
func (s *TagService) UpdateTag(id string, updates map[string]interface{}) (*Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[id]
	if !ok {
		return nil, fmt.Errorf("tag not found: %s", id)
	}

	if tag.ReadOnly {
		return nil, fmt.Errorf("tag is read-only")
	}

	if value, ok := updates["value"].(string); ok {
		tag.Value = value
	}
	if desc, ok := updates["description"].(string); ok {
		tag.Description = desc
	}
	if color, ok := updates["color"].(string); ok {
		tag.Color = color
	}

	return tag, nil
}

// DeleteTag deletes a tag
func (s *TagService) DeleteTag(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[id]
	if !ok {
		return fmt.Errorf("tag not found: %s", id)
	}

	if tag.ReadOnly {
		return fmt.Errorf("cannot delete read-only tag")
	}

	// Remove all bindings
	delete(s.bindings, id)
	delete(s.tags, id)

	return nil
}

// BindTag binds a tag to a resource
func (s *TagService) BindTag(tagID, resourceID, resourceType, tenantID, createdBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[tagID]
	if !ok {
		return fmt.Errorf("tag not found: %s", tagID)
	}

	binding := TagBinding{
		ID:           generateID(),
		TagID:        tagID,
		TagKey:       tag.Key,
		TagValue:     tag.Value,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		TenantID:     tenantID,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}

	s.bindings[tagID] = append(s.bindings[tagID], binding)
	tag.UsageCount++

	return nil
}

// UnbindTag removes a tag binding from a resource
func (s *TagService) UnbindTag(tagID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[tagID]
	if !ok {
		return fmt.Errorf("tag not found: %s", tagID)
	}

	bindings := s.bindings[tagID]
	newBindings := make([]TagBinding, 0)
	for _, b := range bindings {
		if b.ResourceID != resourceID {
			newBindings = append(newBindings, b)
		}
	}

	s.bindings[tagID] = newBindings
	tag.UsageCount = len(newBindings)

	return nil
}

// GetResourceTags gets all tags for a resource
func (s *TagService) GetResourceTags(resourceID string) []TagBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []TagBinding
	for _, bindings := range s.bindings {
		for _, b := range bindings {
			if b.ResourceID == resourceID {
				results = append(results, b)
			}
		}
	}
	return results
}

// GetResourcesByTag gets all resources with a specific tag
func (s *TagService) GetResourcesByTag(tagID string) []TagBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.bindings[tagID]
}

// GetResourcesByTagKey gets all resources with a tag key
func (s *TagService) GetResourcesByTagKey(key string) []TagBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []TagBinding
	for tagID, tag := range s.tags {
		if tag.Key == key {
			results = append(results, s.bindings[tagID]...)
		}
	}
	return results
}

// SearchTags searches tags by key or value
func (s *TagService) SearchTags(query string) []*Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Tag
	for _, tag := range s.tags {
		if containsString(tag.Key, query) || containsString(tag.Value, query) {
			results = append(results, tag)
		}
	}
	return results
}

// GetTagStats gets tag statistics
func (s *TagService) GetTagStats() *TagStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &TagStats{
		TotalTags:     len(s.tags),
		TotalBindings: 0,
		ByCategory:    make(map[string]int),
	}

	for _, tag := range s.tags {
		stats.TotalBindings += tag.UsageCount
		stats.ByCategory[tag.Category]++
	}

	return stats
}

// TagStats represents tag statistics
type TagStats struct {
	TotalTags     int            `json:"total_tags"`
	TotalBindings int            `json:"total_bindings"`
	ByCategory    map[string]int `json:"by_category"`
}

// BulkTagResources tags multiple resources at once
func (s *TagService) BulkTagResources(tagID string, resources []struct {
	ID   string
	Type string
}, tenantID, createdBy string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[tagID]
	if !ok {
		return 0, fmt.Errorf("tag not found: %s", tagID)
	}

	count := 0
	now := time.Now()
	for _, r := range resources {
		binding := TagBinding{
			ID:           generateID(),
			TagID:        tagID,
			TagKey:       tag.Key,
			TagValue:     tag.Value,
			ResourceID:   r.ID,
			ResourceType: r.Type,
			TenantID:     tenantID,
			CreatedAt:    now,
			CreatedBy:    createdBy,
		}
		s.bindings[tagID] = append(s.bindings[tagID], binding)
		count++
	}
	tag.UsageCount += count

	return count, nil
}

// CopyTags copies tags from one resource to another
func (s *TagService) CopyTags(sourceID, targetID, targetTenantID, createdBy string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceTags := make([]TagBinding, 0)
	for _, bindings := range s.bindings {
		for _, b := range bindings {
			if b.ResourceID == sourceID {
				sourceTags = append(sourceTags, b)
			}
		}
	}

	now := time.Now()
	count := 0
	for _, st := range sourceTags {
		binding := TagBinding{
			ID:           generateID(),
			TagID:        st.TagID,
			TagKey:       st.TagKey,
			TagValue:     st.TagValue,
			ResourceID:   targetID,
			ResourceType: st.ResourceType,
			TenantID:     targetTenantID,
			CreatedAt:    now,
			CreatedBy:    createdBy,
		}
		s.bindings[st.TagID] = append(s.bindings[st.TagID], binding)
		count++
	}

	return count, nil
}

// ValidateTags validates tags against rules
func (s *TagService) ValidateTags(resourceType string, tags []TagBinding) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var errors []string

	// Check required tags
	tagKeys := make(map[string]bool)
	for _, t := range tags {
		tagKeys[t.TagKey] = true
	}

	// Would check against tag rules here
	_ = resourceType
	_ = tagKeys

	return errors
}

// ToJSON serializes tag to JSON
func (t *Tag) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// ToJSON serializes binding to JSON
func (b *TagBinding) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}

func containsString(s, substr string) bool {
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