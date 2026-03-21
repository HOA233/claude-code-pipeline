package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a multi-tenant workspace
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Plan        string            `json:"plan"` // free, pro, enterprise
	Status      string            `json:"status"` // active, suspended, deleted
	Settings    TenantSettings    `json:"settings"`
	Quotas      TenantQuotas      `json:"quotas"`
	Usage       TenantUsage       `json:"usage"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
}

// TenantSettings contains tenant-specific settings
type TenantSettings struct {
	Theme           string `json:"theme"`
	Language        string `json:"language"`
	Timezone        string `json:"timezone"`
	NotificationsEnabled bool `json:"notifications_enabled"`
	WebhookURL      string `json:"webhook_url"`
	MaxConcurrency  int    `json:"max_concurrency"`
	DefaultTimeout  int    `json:"default_timeout"`
	AllowedCLIs     []string `json:"allowed_clis"`
}

// TenantQuotas defines resource limits
type TenantQuotas struct {
	MaxTasks         int `json:"max_tasks"`
	MaxPipelines     int `json:"max_pipelines"`
	MaxSchedules     int `json:"max_schedules"`
	MaxConcurrent    int `json:"max_concurrent"`
	MaxStorageMB     int `json:"max_storage_mb"`
	TaskTTLHours     int `json:"task_ttl_hours"`
	RateLimitPerMin  int `json:"rate_limit_per_min"`
}

// TenantUsage tracks current resource usage
type TenantUsage struct {
	Tasks         int `json:"tasks"`
	Pipelines     int `json:"pipelines"`
	Schedules     int `json:"schedules"`
	Concurrent    int `json:"concurrent"`
	StorageMB     int `json:"storage_mb"`
	TasksToday    int `json:"tasks_today"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TenantService manages multi-tenant operations
type TenantService struct {
	mu       sync.RWMutex
	tenants  map[string]*Tenant
	bySlug   map[string]string
	plans    map[string]TenantQuotas
}

// NewTenantService creates a new tenant service
func NewTenantService() *TenantService {
	ts := &TenantService{
		tenants: make(map[string]*Tenant),
		bySlug:  make(map[string]string),
		plans:   DefaultPlans(),
	}

	// Create default tenant
	ts.CreateTenant(context.Background(), &TenantCreateRequest{
		Name: "Default",
		Slug: "default",
		Plan: "free",
	})

	return ts
}

// DefaultPlans returns default quota configurations
func DefaultPlans() map[string]TenantQuotas {
	return map[string]TenantQuotas{
		"free": {
			MaxTasks:        100,
			MaxPipelines:    10,
			MaxSchedules:    5,
			MaxConcurrent:   2,
			MaxStorageMB:    100,
			TaskTTLHours:    24,
			RateLimitPerMin: 30,
		},
		"pro": {
			MaxTasks:        1000,
			MaxPipelines:    50,
			MaxSchedules:    20,
			MaxConcurrent:   10,
			MaxStorageMB:    1000,
			TaskTTLHours:    72,
			RateLimitPerMin: 100,
		},
		"enterprise": {
			MaxTasks:        10000,
			MaxPipelines:    200,
			MaxSchedules:    100,
			MaxConcurrent:   50,
			MaxStorageMB:    10000,
			TaskTTLHours:    168,
			RateLimitPerMin: 500,
		},
	}
}

// TenantCreateRequest for creating tenants
type TenantCreateRequest struct {
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	Plan      string            `json:"plan"`
	Settings  TenantSettings    `json:"settings"`
	Metadata  map[string]string `json:"metadata"`
}

// CreateTenant creates a new tenant
func (ts *TenantService) CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Validate
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	if _, exists := ts.bySlug[slug]; exists {
		return nil, fmt.Errorf("tenant with slug '%s' already exists", slug)
	}

	plan := req.Plan
	if plan == "" {
		plan = "free"
	}

	quotas, ok := ts.plans[plan]
	if !ok {
		return nil, fmt.Errorf("invalid plan: %s", plan)
	}

	now := time.Now()
	tenant := &Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Slug:      slug,
		Plan:      plan,
		Status:    "active",
		Quotas:    quotas,
		Usage:     TenantUsage{LastUpdated: now},
		Settings:  req.Settings,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set default settings
	if tenant.Settings.MaxConcurrency == 0 {
		tenant.Settings.MaxConcurrency = quotas.MaxConcurrent
	}
	if tenant.Settings.DefaultTimeout == 0 {
		tenant.Settings.DefaultTimeout = 300
	}

	ts.tenants[tenant.ID] = tenant
	ts.bySlug[slug] = tenant.ID

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (ts *TenantService) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug
func (ts *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tenantID, exists := ts.bySlug[slug]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", slug)
	}

	return ts.tenants[tenantID], nil
}

// ListTenants lists all tenants
func (ts *TenantService) ListTenants(ctx context.Context) []*Tenant {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(ts.tenants))
	for _, t := range ts.tenants {
		tenants = append(tenants, t)
	}
	return tenants
}

// UpdateTenant updates a tenant
func (ts *TenantService) UpdateTenant(ctx context.Context, tenantID string, updates map[string]interface{}) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	if name, ok := updates["name"].(string); ok {
		tenant.Name = name
	}
	if plan, ok := updates["plan"].(string); ok {
		if quotas, exists := ts.plans[plan]; exists {
			tenant.Plan = plan
			tenant.Quotas = quotas
		}
	}
	if status, ok := updates["status"].(string); ok {
		tenant.Status = status
	}
	if settings, ok := updates["settings"].(TenantSettings); ok {
		tenant.Settings = settings
	}

	tenant.UpdatedAt = time.Now()
	return nil
}

// DeleteTenant soft-deletes a tenant
func (ts *TenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = "deleted"
	tenant.UpdatedAt = time.Now()

	return nil
}

// CheckQuota checks if a tenant can use more resources
func (ts *TenantService) CheckQuota(ctx context.Context, tenantID, resource string) error {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	if tenant.Status != "active" {
		return fmt.Errorf("tenant is not active: %s", tenant.Status)
	}

	switch resource {
	case "tasks":
		if tenant.Usage.Tasks >= tenant.Quotas.MaxTasks {
			return fmt.Errorf("task quota exceeded: %d/%d", tenant.Usage.Tasks, tenant.Quotas.MaxTasks)
		}
	case "pipelines":
		if tenant.Usage.Pipelines >= tenant.Quotas.MaxPipelines {
			return fmt.Errorf("pipeline quota exceeded: %d/%d", tenant.Usage.Pipelines, tenant.Quotas.MaxPipelines)
		}
	case "schedules":
		if tenant.Usage.Schedules >= tenant.Quotas.MaxSchedules {
			return fmt.Errorf("schedule quota exceeded: %d/%d", tenant.Usage.Schedules, tenant.Quotas.MaxSchedules)
		}
	case "concurrent":
		if tenant.Usage.Concurrent >= tenant.Quotas.MaxConcurrent {
			return fmt.Errorf("concurrency quota exceeded: %d/%d", tenant.Usage.Concurrent, tenant.Quotas.MaxConcurrent)
		}
	}

	return nil
}

// IncrementUsage increments usage counters
func (ts *TenantService) IncrementUsage(ctx context.Context, tenantID, resource string, delta int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	switch resource {
	case "tasks":
		tenant.Usage.Tasks += delta
	case "pipelines":
		tenant.Usage.Pipelines += delta
	case "schedules":
		tenant.Usage.Schedules += delta
	case "concurrent":
		tenant.Usage.Concurrent += delta
	case "tasks_today":
		tenant.Usage.TasksToday += delta
	}

	tenant.Usage.LastUpdated = time.Now()
	return nil
}

// GetTenantStats returns tenant statistics
func (ts *TenantService) GetTenantStats(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tenant, exists := ts.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return map[string]interface{}{
		"tenant_id":     tenant.ID,
		"name":          tenant.Name,
		"plan":          tenant.Plan,
		"status":        tenant.Status,
		"usage":         tenant.Usage,
		"quotas":        tenant.Quotas,
		"utilization": map[string]float64{
			"tasks":     float64(tenant.Usage.Tasks) / float64(tenant.Quotas.MaxTasks) * 100,
			"pipelines": float64(tenant.Usage.Pipelines) / float64(tenant.Quotas.MaxPipelines) * 100,
			"schedules": float64(tenant.Usage.Schedules) / float64(tenant.Quotas.MaxSchedules) * 100,
		},
	}, nil
}

// ToJSON exports tenant as JSON
func (t *Tenant) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// generateSlug generates a URL-safe slug from name
func generateSlug(name string) string {
	// Simple slug generation
	slug := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			slug += string(c)
		} else if c == ' ' || c == '_' {
			slug += "-"
		}
	}
	return slug
}