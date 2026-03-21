package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
)

// WorkspaceManager manages isolated workspaces for multi-tenant support
type WorkspaceManager struct {
	redis     *repository.RedisClient
	workspaces sync.Map
	mu        sync.RWMutex
}

// Workspace represents an isolated workspace
type Workspace struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Owner       string                 `json:"owner"`
	TenantID    string                 `json:"tenant_id"`
	Description string                 `json:"description,omitempty"`

	// Resource limits
	MaxTasks    int                    `json:"max_tasks"`
	MaxPipelines int                   `json:"max_pipelines"`
	MaxConcurrency int                 `json:"max_concurrency"`

	// Isolation settings
	IsolatedSkills bool                `json:"isolated_skills"`
	IsolatedData   bool                `json:"isolated_data"`

	// Metadata
	Config      map[string]interface{} `json:"config,omitempty"`
	Tags        []string               `json:"tags,omitempty"`

	// Status
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// WorkspaceStats contains workspace statistics
type WorkspaceStats struct {
	WorkspaceID    string `json:"workspace_id"`
	TotalTasks     int    `json:"total_tasks"`
	ActiveTasks    int    `json:"active_tasks"`
	TotalPipelines int    `json:"total_pipelines"`
	TotalRuns      int    `json:"total_runs"`
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(redis *repository.RedisClient) *WorkspaceManager {
	return &WorkspaceManager{
		redis: redis,
	}
}

// CreateWorkspace creates a new workspace
func (wm *WorkspaceManager) CreateWorkspace(ctx context.Context, name, owner, tenantID string, opts ...WorkspaceOption) (*Workspace, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace := &Workspace{
		ID:             fmt.Sprintf("ws-%s", uuid.New().String()[:8]),
		Name:           name,
		Owner:          owner,
		TenantID:       tenantID,
		MaxTasks:       100,
		MaxPipelines:   50,
		MaxConcurrency: 5,
		IsolatedSkills: true,
		IsolatedData:   true,
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Config:         make(map[string]interface{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(workspace)
	}

	// Save to Redis
	if err := wm.saveWorkspace(ctx, workspace); err != nil {
		return nil, err
	}

	wm.workspaces.Store(workspace.ID, workspace)
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (wm *WorkspaceManager) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	// Check memory cache
	if val, ok := wm.workspaces.Load(id); ok {
		return val.(*Workspace), nil
	}

	// Load from Redis
	return wm.loadWorkspace(ctx, id)
}

// ListWorkspaces lists all workspaces for a tenant
func (wm *WorkspaceManager) ListWorkspaces(ctx context.Context, tenantID string) ([]*Workspace, error) {
	keys, err := wm.redis.ListWorkspaceKeys(ctx)
	if err != nil {
		return nil, err
	}

	workspaces := make([]*Workspace, 0)
	for _, key := range keys {
		ws, err := wm.loadWorkspace(ctx, key)
		if err != nil {
			continue
		}
		if tenantID == "" || ws.TenantID == tenantID {
			workspaces = append(workspaces, ws)
		}
	}

	return workspaces, nil
}

// UpdateWorkspace updates a workspace
func (wm *WorkspaceManager) UpdateWorkspace(ctx context.Context, id string, updates map[string]interface{}) error {
	workspace, err := wm.GetWorkspace(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		workspace.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		workspace.Description = description
	}
	if maxTasks, ok := updates["max_tasks"].(int); ok {
		workspace.MaxTasks = maxTasks
	}
	if maxPipelines, ok := updates["max_pipelines"].(int); ok {
		workspace.MaxPipelines = maxPipelines
	}
	if active, ok := updates["active"].(bool); ok {
		workspace.Active = active
	}

	workspace.UpdatedAt = time.Now()

	if err := wm.saveWorkspace(ctx, workspace); err != nil {
		return err
	}

	wm.workspaces.Store(id, workspace)
	return nil
}

// DeleteWorkspace deletes a workspace
func (wm *WorkspaceManager) DeleteWorkspace(ctx context.Context, id string) error {
	wm.workspaces.Delete(id)
	return wm.redis.DeleteWorkspace(ctx, id)
}

// GetWorkspaceStats returns statistics for a workspace
func (wm *WorkspaceManager) GetWorkspaceStats(ctx context.Context, workspaceID string) (*WorkspaceStats, error) {
	// Verify workspace exists
	_, err := wm.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	stats := &WorkspaceStats{
		WorkspaceID: workspaceID,
	}

	// Get tasks count for this workspace
	tasks, _ := wm.redis.GetAllTasks(ctx)
	for _, task := range tasks {
		// In real implementation, tasks would have workspace_id
		stats.TotalTasks++
		if task.Status == model.TaskStatusRunning {
			stats.ActiveTasks++
		}
	}

	return stats, nil
}

// CheckResourceLimit checks if a workspace can create more resources
func (wm *WorkspaceManager) CheckResourceLimit(ctx context.Context, workspaceID string, resourceType string) (bool, error) {
	workspace, err := wm.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	stats, err := wm.GetWorkspaceStats(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	switch resourceType {
	case "task":
		return stats.TotalTasks < workspace.MaxTasks, nil
	case "pipeline":
		return stats.TotalPipelines < workspace.MaxPipelines, nil
	default:
		return true, nil
	}
}

// WorkspaceOption is a functional option for workspace creation
type WorkspaceOption func(*Workspace)

// WithMaxTasks sets the maximum number of tasks
func WithMaxTasks(max int) WorkspaceOption {
	return func(w *Workspace) {
		w.MaxTasks = max
	}
}

// WithMaxPipelines sets the maximum number of pipelines
func WithMaxPipelines(max int) WorkspaceOption {
	return func(w *Workspace) {
		w.MaxPipelines = max
	}
}

// WithMaxConcurrency sets the maximum concurrency
func WithMaxConcurrency(max int) WorkspaceOption {
	return func(w *Workspace) {
		w.MaxConcurrency = max
	}
}

// WithExpiry sets an expiry time for the workspace
func WithExpiry(expiry time.Time) WorkspaceOption {
	return func(w *Workspace) {
		w.ExpiresAt = &expiry
	}
}

// WithIsolation sets isolation options
func WithIsolation(isolatedSkills, isolatedData bool) WorkspaceOption {
	return func(w *Workspace) {
		w.IsolatedSkills = isolatedSkills
		w.IsolatedData = isolatedData
	}
}

// WithConfig sets custom configuration
func WithConfig(config map[string]interface{}) WorkspaceOption {
	return func(w *Workspace) {
		w.Config = config
	}
}

func (wm *WorkspaceManager) saveWorkspace(ctx context.Context, workspace *Workspace) error {
	data, err := json.Marshal(workspace)
	if err != nil {
		return err
	}
	return wm.redis.SaveWorkspace(ctx, workspace.ID, data)
}

func (wm *WorkspaceManager) loadWorkspace(ctx context.Context, id string) (*Workspace, error) {
	data, err := wm.redis.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, err
	}

	return &workspace, nil
}