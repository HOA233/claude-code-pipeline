package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// GitOpsManager manages GitOps-based pipeline configurations
type GitOpsManager struct {
	mu          sync.RWMutex
	repos       map[string]*GitOpsRepo
	syncTicker  *time.Ticker
	syncEnabled bool
}

// GitOpsRepo represents a GitOps repository
type GitOpsRepo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Branch       string            `json:"branch"`
	Path         string            `json:"path"`
	AuthProvider string            `json:"auth_provider"`
	SecretRef    string            `json:"secret_ref"`
	LastSync     time.Time         `json:"last_sync"`
	LastCommit   string            `json:"last_commit"`
	Status       string            `json:"status"`
	Error        string            `json:"error,omitempty"`
	Annotations  map[string]string `json:"annotations"`
}

// PipelineDefinition represents a pipeline definition in Git
type PipelineDefinition struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Version     string                 `yaml:"version" json:"version"`
	Mode        string                 `yaml:"mode" json:"mode"`
	Steps       []StepDefinition       `yaml:"steps" json:"steps"`
	Triggers    []TriggerDefinition    `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Parameters  []ParameterDefinition  `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Environment map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// StepDefinition represents a step in a pipeline
type StepDefinition struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name" json:"name"`
	CLI         string                 `yaml:"cli" json:"cli"`
	Action      string                 `yaml:"action" json:"action"`
	Command     string                 `yaml:"command,omitempty" json:"command,omitempty"`
	Params      map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`
	DependsOn   []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Condition   string                 `yaml:"condition,omitempty" json:"condition,omitempty"`
	Timeout     int                    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retry       *RetryConfig           `yaml:"retry,omitempty" json:"retry,omitempty"`
	OnError     string                 `yaml:"on_error,omitempty" json:"on_error,omitempty"`
	Environment map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// TriggerDefinition for pipeline triggers
type TriggerDefinition struct {
	Type   string                 `yaml:"type" json:"type"` // webhook, schedule, event
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// ParameterDefinition for pipeline parameters
type ParameterDefinition struct {
	Name        string      `yaml:"name" json:"name"`
	Type        string      `yaml:"type" json:"type"`
	Required    bool        `yaml:"required" json:"required"`
	Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Validation  Validation  `yaml:"validation,omitempty" json:"validation,omitempty"`
}

// Validation rules for parameters
type Validation struct {
	Min    interface{} `yaml:"min,omitempty" json:"min,omitempty"`
	Max    interface{} `yaml:"max,omitempty" json:"max,omitempty"`
	Pattern string      `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Enum   []string    `yaml:"enum,omitempty" json:"enum,omitempty"`
}

// RetryConfig for step retry behavior
type RetryConfig struct {
	MaxAttempts int    `yaml:"max_attempts" json:"max_attempts"`
	Backoff     string `yaml:"backoff" json:"backoff"` // linear, exponential
	Interval    int    `yaml:"interval" json:"interval"`
}

// NewGitOpsManager creates a new GitOps manager
func NewGitOpsManager() *GitOpsManager {
	return &GitOpsManager{
		repos: make(map[string]*GitOpsRepo),
	}
}

// AddRepo adds a GitOps repository
func (g *GitOpsManager) AddRepo(ctx context.Context, repo *GitOpsRepo) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if repo.ID == "" {
		repo.ID = generateRepoID(repo.URL)
	}

	if _, exists := g.repos[repo.ID]; exists {
		return fmt.Errorf("repository already exists: %s", repo.ID)
	}

	// Set defaults
	if repo.Branch == "" {
		repo.Branch = "main"
	}
	if repo.Path == "" {
		repo.Path = "pipelines"
	}
	repo.Status = "pending"

	g.repos[repo.ID] = repo
	return nil
}

// RemoveRepo removes a GitOps repository
func (g *GitOpsManager) RemoveRepo(ctx context.Context, repoID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.repos[repoID]; !exists {
		return fmt.Errorf("repository not found: %s", repoID)
	}

	delete(g.repos, repoID)
	return nil
}

// GetRepo gets a GitOps repository
func (g *GitOpsManager) GetRepo(ctx context.Context, repoID string) (*GitOpsRepo, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	repo, exists := g.repos[repoID]
	if !exists {
		return nil, fmt.Errorf("repository not found: %s", repoID)
	}

	return repo, nil
}

// ListRepos lists all GitOps repositories
func (g *GitOpsManager) ListRepos(ctx context.Context) []*GitOpsRepo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	repos := make([]*GitOpsRepo, 0, len(g.repos))
	for _, r := range g.repos {
		repos = append(repos, r)
	}
	return repos
}

// SyncRepo synchronizes a GitOps repository
func (g *GitOpsManager) SyncRepo(ctx context.Context, repoID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	repo, exists := g.repos[repoID]
	if !exists {
		return fmt.Errorf("repository not found: %s", repoID)
	}

	repo.Status = "syncing"
	repo.Error = ""

	// Simulate sync (in real implementation, would use git client)
	// This would:
	// 1. Clone/fetch the repository
	// 2. Read pipeline definitions from the path
	// 3. Validate and register pipelines
	// 4. Update status

	repo.LastSync = time.Now()
	repo.Status = "synced"

	return nil
}

// SyncAll synchronizes all repositories
func (g *GitOpsManager) SyncAll(ctx context.Context) error {
	g.mu.RLock()
	repoIDs := make([]string, 0, len(g.repos))
	for id := range g.repos {
		repoIDs = append(repoIDs, id)
	}
	g.mu.RUnlock()

	var errs []string
	for _, id := range repoIDs {
		if err := g.SyncRepo(ctx, id); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// StartSync starts automatic synchronization
func (g *GitOpsManager) StartSync(interval time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.syncEnabled {
		return
	}

	g.syncTicker = time.NewTicker(interval)
	g.syncEnabled = true

	go func() {
		for range g.syncTicker.C {
			g.SyncAll(context.Background())
		}
	}()
}

// StopSync stops automatic synchronization
func (g *GitOpsManager) StopSync() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.syncEnabled {
		return
	}

	g.syncTicker.Stop()
	g.syncEnabled = false
}

// ValidatePipeline validates a pipeline definition
func (g *GitOpsManager) ValidatePipeline(pipeline *PipelineDefinition) error {
	if pipeline.Name == "" {
		return errors.New("pipeline name is required")
	}

	if pipeline.Mode == "" {
		pipeline.Mode = "serial"
	}

	if pipeline.Mode != "serial" && pipeline.Mode != "parallel" && pipeline.Mode != "hybrid" {
		return fmt.Errorf("invalid mode: %s", pipeline.Mode)
	}

	if len(pipeline.Steps) == 0 {
		return errors.New("pipeline must have at least one step")
	}

	// Validate step IDs are unique
	stepIDs := make(map[string]bool)
	for _, step := range pipeline.Steps {
		if step.ID == "" {
			return errors.New("step ID is required")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true

		if step.CLI == "" {
			return fmt.Errorf("step %s: CLI is required", step.ID)
		}

		// Validate depends_on references
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s: unknown dependency: %s", step.ID, dep)
			}
		}
	}

	return nil
}

// ParsePipelineDefinition parses a pipeline definition from YAML
func (g *GitOpsManager) ParsePipelineDefinition(data []byte) (*PipelineDefinition, error) {
	var pipeline PipelineDefinition
	if err := json.Unmarshal(data, &pipeline); err != nil {
		// Try YAML
		// In real implementation, would use yaml.Unmarshal
		return nil, fmt.Errorf("failed to parse pipeline definition: %w", err)
	}

	if err := g.ValidatePipeline(&pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// GetRepoStatus returns the status of a repository
func (g *GitOpsManager) GetRepoStatus(ctx context.Context, repoID string) (map[string]interface{}, error) {
	repo, err := g.GetRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":          repo.ID,
		"name":        repo.Name,
		"url":         repo.URL,
		"branch":      repo.Branch,
		"status":      repo.Status,
		"last_sync":   repo.LastSync,
		"last_commit": repo.LastCommit,
		"error":       repo.Error,
	}, nil
}

// generateRepoID generates a unique ID for a repository
func generateRepoID(url string) string {
	// Simple ID generation from URL
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "-" + parts[len(parts)-1]
	}
	return strings.ReplaceAll(url, "/", "-")
}