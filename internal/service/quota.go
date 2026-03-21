package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// QuotaService manages resource quotas and usage tracking
type QuotaService struct {
	mu       sync.RWMutex
	quotas   map[string]*Quota
	usage    map[string]*Usage
	limiters map[string]*QuotaLimiter
}

// Quota represents a resource quota
type Quota struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	TenantID    string            `json:"tenant_id"`
	Resources   map[string]int64  `json:"resources"` // resource -> limit
	Period      QuotaPeriod       `json:"period"`
	PeriodStart time.Time         `json:"period_start"`
	PeriodEnd   time.Time         `json:"period_end"`
	Enforced    bool              `json:"enforced"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Usage represents resource usage
type Usage struct {
	TenantID    string           `json:"tenant_id"`
	Resources   map[string]int64 `json:"resources"`
	PeriodStart time.Time        `json:"period_start"`
	LastUpdated time.Time        `json:"last_updated"`
}

// QuotaPeriod represents a quota period
type QuotaPeriod string

const (
	QuotaPeriodDaily   QuotaPeriod = "daily"
	QuotaPeriodWeekly  QuotaPeriod = "weekly"
	QuotaPeriodMonthly QuotaPeriod = "monthly"
	QuotaPeriodYearly  QuotaPeriod = "yearly"
)

// QuotaLimiter tracks real-time usage
type QuotaLimiter struct {
	tenantID string
	limits   map[string]int64
	usage    map[string]int64
	mu       sync.RWMutex
}

// QuotaCheckResult represents a quota check result
type QuotaCheckResult struct {
	Allowed     bool             `json:"allowed"`
	Resource    string           `json:"resource"`
	Requested   int64            `json:"requested"`
	Available   int64            `json:"available"`
	Limit       int64            `json:"limit"`
	Current     int64            `json:"current"`
	Exceeded    []string         `json:"exceeded,omitempty"`
	ResetAt     time.Time        `json:"reset_at,omitempty"`
}

// NewQuotaService creates a new quota service
func NewQuotaService() *QuotaService {
	return &QuotaService{
		quotas:   make(map[string]*Quota),
		usage:    make(map[string]*Usage),
		limiters: make(map[string]*QuotaLimiter),
	}
}

// SetQuota sets a quota for a tenant
func (s *QuotaService) SetQuota(ctx context.Context, quota *Quota) error {
	if quota.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if quota.ID == "" {
		quota.ID = generateID()
	}
	quota.UpdatedAt = now
	if quota.CreatedAt.IsZero() {
		quota.CreatedAt = now
	}

	// Set period dates
	if quota.PeriodStart.IsZero() {
		quota.PeriodStart = s.getPeriodStart(quota.Period)
	}
	quota.PeriodEnd = s.getPeriodEnd(quota.PeriodStart, quota.Period)

	s.quotas[quota.ID] = quota

	// Initialize limiter
	s.limiters[quota.TenantID] = &QuotaLimiter{
		tenantID: quota.TenantID,
		limits:   quota.Resources,
		usage:    make(map[string]int64),
	}

	// Initialize usage
	if _, exists := s.usage[quota.TenantID]; !exists {
		s.usage[quota.TenantID] = &Usage{
			TenantID:    quota.TenantID,
			Resources:   make(map[string]int64),
			PeriodStart: quota.PeriodStart,
		}
	}

	return nil
}

// GetQuota gets a quota by ID
func (s *QuotaService) GetQuota(id string) (*Quota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, exists := s.quotas[id]
	if !exists {
		return nil, fmt.Errorf("quota not found: %s", id)
	}
	return quota, nil
}

// GetQuotaByTenant gets quota for a tenant
func (s *QuotaService) GetQuotaByTenant(tenantID string) (*Quota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, quota := range s.quotas {
		if quota.TenantID == tenantID {
			return quota, nil
		}
	}
	return nil, fmt.Errorf("quota not found for tenant: %s", tenantID)
}

// CheckQuota checks if usage is within quota limits
func (s *QuotaService) CheckQuota(ctx context.Context, tenantID, resource string, amount int64) (*QuotaCheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, err := s.getQuotaByTenantLocked(tenantID)
	if err != nil {
		return nil, err
	}

	limit, hasLimit := quota.Resources[resource]
	if !hasLimit {
		// No limit for this resource
		return &QuotaCheckResult{
			Allowed:   true,
			Resource:  resource,
			Requested: amount,
			Limit:     -1,
		}, nil
	}

	usage := s.usage[tenantID]
	currentUsage := usage.Resources[resource]
	available := limit - currentUsage

	result := &QuotaCheckResult{
		Resource:  resource,
		Requested: amount,
		Available: available,
		Limit:     limit,
		Current:   currentUsage,
		ResetAt:   quota.PeriodEnd,
	}

	if currentUsage+amount > limit {
		result.Allowed = false
		if quota.Enforced {
			return result, fmt.Errorf("quota exceeded for %s: limit %d, current %d, requested %d",
				resource, limit, currentUsage, amount)
		}
	} else {
		result.Allowed = true
	}

	return result, nil
}

// ConsumeQuota consumes quota for a resource
func (s *QuotaService) ConsumeQuota(ctx context.Context, tenantID, resource string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, err := s.getQuotaByTenantLocked(tenantID)
	if err != nil {
		return err
	}

	// Check if quota is enforced
	if quota.Enforced {
		limit, hasLimit := quota.Resources[resource]
		if hasLimit {
			usage := s.usage[tenantID]
			currentUsage := usage.Resources[resource]
			if currentUsage+amount > limit {
				return fmt.Errorf("quota exceeded for %s", resource)
			}
		}
	}

	// Update usage
	usage := s.usage[tenantID]
	if usage.Resources == nil {
		usage.Resources = make(map[string]int64)
	}
	usage.Resources[resource] += amount
	usage.LastUpdated = time.Now()

	// Update limiter
	if limiter, exists := s.limiters[tenantID]; exists {
		limiter.mu.Lock()
		limiter.usage[resource] += amount
		limiter.mu.Unlock()
	}

	return nil
}

// ReleaseQuota releases quota for a resource
func (s *QuotaService) ReleaseQuota(ctx context.Context, tenantID, resource string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage := s.usage[tenantID]
	if usage == nil {
		return nil
	}

	if usage.Resources[resource] >= amount {
		usage.Resources[resource] -= amount
	} else {
		usage.Resources[resource] = 0
	}
	usage.LastUpdated = time.Now()

	// Update limiter
	if limiter, exists := s.limiters[tenantID]; exists {
		limiter.mu.Lock()
		if limiter.usage[resource] >= amount {
			limiter.usage[resource] -= amount
		} else {
			limiter.usage[resource] = 0
		}
		limiter.mu.Unlock()
	}

	return nil
}

// GetUsage gets current usage for a tenant
func (s *QuotaService) GetUsage(tenantID string) (*Usage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usage, exists := s.usage[tenantID]
	if !exists {
		return nil, fmt.Errorf("usage not found for tenant: %s", tenantID)
	}
	return usage, nil
}

// ResetUsage resets usage for a tenant (usually at period start)
func (s *QuotaService) ResetUsage(tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, err := s.getQuotaByTenantLocked(tenantID)
	if err != nil {
		return err
	}

	// Reset usage
	s.usage[tenantID] = &Usage{
		TenantID:    tenantID,
		Resources:   make(map[string]int64),
		PeriodStart: s.getPeriodStart(quota.Period),
		LastUpdated: time.Now(),
	}

	// Update quota period
	quota.PeriodStart = s.getPeriodStart(quota.Period)
	quota.PeriodEnd = s.getPeriodEnd(quota.PeriodStart, quota.Period)
	quota.UpdatedAt = time.Now()

	// Reset limiter
	if limiter, exists := s.limiters[tenantID]; exists {
		limiter.mu.Lock()
		limiter.usage = make(map[string]int64)
		limiter.mu.Unlock()
	}

	return nil
}

// ListQuotas lists all quotas
func (s *QuotaService) ListQuotas() []*Quota {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotas := make([]*Quota, 0, len(s.quotas))
	for _, quota := range s.quotas {
		quotas = append(quotas, quota)
	}
	return quotas
}

// DeleteQuota deletes a quota
func (s *QuotaService) DeleteQuota(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, exists := s.quotas[id]
	if !exists {
		return fmt.Errorf("quota not found: %s", id)
	}

	delete(s.quotas, id)
	delete(s.limiters, quota.TenantID)

	return nil
}

// GetQuotaStatus gets quota status for a tenant
func (s *QuotaService) GetQuotaStatus(tenantID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, err := s.getQuotaByTenantLocked(tenantID)
	if err != nil {
		return nil, err
	}

	usage := s.usage[tenantID]
	if usage == nil {
		usage = &Usage{Resources: make(map[string]int64)}
	}

	status := make(map[string]interface{})
	for resource, limit := range quota.Resources {
		current := usage.Resources[resource]
		status[resource] = map[string]interface{}{
			"limit":     limit,
			"current":   current,
			"available": limit - current,
			"percent":   float64(current) / float64(limit) * 100,
		}
	}

	return map[string]interface{}{
		"tenant_id":    tenantID,
		"period":       quota.Period,
		"period_start": quota.PeriodStart,
		"period_end":   quota.PeriodEnd,
		"enforced":     quota.Enforced,
		"resources":    status,
	}, nil
}

// CheckAndResetPeriod checks and resets period if needed
func (s *QuotaService) CheckAndResetPeriod() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	reset := 0

	for _, quota := range s.quotas {
		if now.After(quota.PeriodEnd) {
			// Reset the period
			quota.PeriodStart = s.getPeriodStart(quota.Period)
			quota.PeriodEnd = s.getPeriodEnd(quota.PeriodStart, quota.Period)
			quota.UpdatedAt = now

			// Reset usage
			if usage, exists := s.usage[quota.TenantID]; exists {
				usage.Resources = make(map[string]int64)
				usage.PeriodStart = quota.PeriodStart
				usage.LastUpdated = now
			}

			// Reset limiter
			if limiter, exists := s.limiters[quota.TenantID]; exists {
				limiter.mu.Lock()
				limiter.usage = make(map[string]int64)
				limiter.mu.Unlock()
			}

			reset++
		}
	}

	return reset
}

// GetStats returns quota service statistics
func (s *QuotaService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enforced := 0
	for _, quota := range s.quotas {
		if quota.Enforced {
			enforced++
		}
	}

	return map[string]interface{}{
		"total_quotas":  len(s.quotas),
		"enforced":      enforced,
		"total_tenants": len(s.usage),
	}
}

func (s *QuotaService) getQuotaByTenantLocked(tenantID string) (*Quota, error) {
	for _, quota := range s.quotas {
		if quota.TenantID == tenantID {
			return quota, nil
		}
	}
	return nil, fmt.Errorf("quota not found for tenant: %s", tenantID)
}

func (s *QuotaService) getPeriodStart(period QuotaPeriod) time.Time {
	now := time.Now()
	switch period {
	case QuotaPeriodDaily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case QuotaPeriodWeekly:
		weekday := int(now.Weekday())
		return time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, now.Location())
	case QuotaPeriodMonthly:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case QuotaPeriodYearly:
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	default:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
}

func (s *QuotaService) getPeriodEnd(start time.Time, period QuotaPeriod) time.Time {
	switch period {
	case QuotaPeriodDaily:
		return start.AddDate(0, 0, 1)
	case QuotaPeriodWeekly:
		return start.AddDate(0, 0, 7)
	case QuotaPeriodMonthly:
		return start.AddDate(0, 1, 0)
	case QuotaPeriodYearly:
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(0, 0, 1)
	}
}

// ToJSON serializes quota to JSON
func (q *Quota) ToJSON() ([]byte, error) {
	return json.Marshal(q)
}

// ToJSON serializes usage to JSON
func (u *Usage) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}