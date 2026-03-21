package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// EnvironmentService manages deployment environments
type EnvironmentService struct {
	mu           sync.RWMutex
	environments map[string]*Environment
	deployments  map[string][]*Deployment
	variables    map[string]map[string]EnvironmentVariable
}

// Environment represents a deployment environment
type Environment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	TenantID    string            `json:"tenant_id"`
	Type        EnvironmentType   `json:"type"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Config      map[string]interface{} `json:"config"`
	Variables   map[string]string `json:"variables"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastDeployedAt *time.Time     `json:"last_deployed_at,omitempty"`
}

// EnvironmentType represents environment type
type EnvironmentType string

const (
	EnvTypeDevelopment EnvironmentType = "development"
	EnvTypeStaging     EnvironmentType = "staging"
	EnvTypeProduction  EnvironmentType = "production"
	EnvTypePreview     EnvironmentType = "preview"
)

// Deployment represents a deployment to an environment
type Deployment struct {
	ID            string    `json:"id"`
	EnvironmentID string    `json:"environment_id"`
	Version       string    `json:"version"`
	Status        string    `json:"status"` // pending, deploying, succeeded, failed, rolled_back
	TriggeredBy   string    `json:"triggered_by"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Duration      int64     `json:"duration_ms"`
	Changes       []string  `json:"changes"`
	RollbackFrom  string    `json:"rollback_from,omitempty"`
	Logs          []string  `json:"logs"`
}

// EnvironmentVariable represents an environment variable
type EnvironmentVariable struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Secret    bool      `json:"secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeploymentStatus constants
const (
	DeploymentStatusPending     = "pending"
	DeploymentStatusDeploying   = "deploying"
	DeploymentStatusSucceeded   = "succeeded"
	DeploymentStatusFailed      = "failed"
	DeploymentStatusRolledBack  = "rolled_back"
)

// NewEnvironmentService creates a new environment service
func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{
		environments: make(map[string]*Environment),
		deployments:  make(map[string][]*Deployment),
		variables:    make(map[string]map[string]EnvironmentVariable),
	}
}

// CreateEnvironment creates a new environment
func (s *EnvironmentService) CreateEnvironment(ctx context.Context, env *Environment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env.Name == "" {
		return fmt.Errorf("name is required")
	}
	if env.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	now := time.Now()
	if env.ID == "" {
		env.ID = generateID()
	}
	env.CreatedAt = now
	env.UpdatedAt = now
	env.Status = "active"

	if env.Config == nil {
		env.Config = make(map[string]interface{})
	}
	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}

	s.environments[env.ID] = env
	s.variables[env.ID] = make(map[string]EnvironmentVariable)

	return nil
}

// GetEnvironment gets an environment by ID
func (s *EnvironmentService) GetEnvironment(id string) (*Environment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	env, ok := s.environments[id]
	if !ok {
		return nil, fmt.Errorf("environment not found: %s", id)
	}
	return env, nil
}

// GetEnvironmentByName gets an environment by name and tenant
func (s *EnvironmentService) GetEnvironmentByName(tenantID, name string) (*Environment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, env := range s.environments {
		if env.TenantID == tenantID && env.Name == name {
			return env, nil
		}
	}
	return nil, fmt.Errorf("environment not found: %s", name)
}

// ListEnvironments lists environments for a tenant
func (s *EnvironmentService) ListEnvironments(tenantID string) []*Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var envs []*Environment
	for _, env := range s.environments {
		if env.TenantID == tenantID {
			envs = append(envs, env)
		}
	}
	return envs
}

// UpdateEnvironment updates an environment
func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, id string, updates map[string]interface{}) (*Environment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	env, ok := s.environments[id]
	if !ok {
		return nil, fmt.Errorf("environment not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		env.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		env.Description = desc
	}
	if status, ok := updates["status"].(string); ok {
		env.Status = status
	}
	if config, ok := updates["config"].(map[string]interface{}); ok {
		env.Config = config
	}
	if tags, ok := updates["tags"].([]string); ok {
		env.Tags = tags
	}

	env.UpdatedAt = time.Now()

	return env, nil
}

// DeleteEnvironment deletes an environment
func (s *EnvironmentService) DeleteEnvironment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.environments[id]; !ok {
		return fmt.Errorf("environment not found: %s", id)
	}

	delete(s.environments, id)
	delete(s.deployments, id)
	delete(s.variables, id)

	return nil
}

// SetVariable sets an environment variable
func (s *EnvironmentService) SetVariable(envID, key, value string, secret bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.environments[envID]; !ok {
		return fmt.Errorf("environment not found: %s", envID)
	}

	vars, ok := s.variables[envID]
	if !ok {
		vars = make(map[string]EnvironmentVariable)
		s.variables[envID] = vars
	}

	now := time.Now()
	vars[key] = EnvironmentVariable{
		Key:       key,
		Value:     value,
		Secret:    secret,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Also update the environment's variables map (non-secret only)
	if !secret {
		s.environments[envID].Variables[key] = value
	}

	return nil
}

// GetVariable gets an environment variable
func (s *EnvironmentService) GetVariable(envID, key string) (EnvironmentVariable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vars, ok := s.variables[envID]
	if !ok {
		return EnvironmentVariable{}, fmt.Errorf("environment not found: %s", envID)
	}

	v, ok := vars[key]
	if !ok {
		return EnvironmentVariable{}, fmt.Errorf("variable not found: %s", key)
	}

	return v, nil
}

// ListVariables lists all variables for an environment
func (s *EnvironmentService) ListVariables(envID string, includeSecrets bool) ([]EnvironmentVariable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vars, ok := s.variables[envID]
	if !ok {
		return nil, fmt.Errorf("environment not found: %s", envID)
	}

	var result []EnvironmentVariable
	for _, v := range vars {
		if !v.Secret || includeSecrets {
			result = append(result, v)
		}
	}
	return result, nil
}

// DeleteVariable deletes an environment variable
func (s *EnvironmentService) DeleteVariable(envID, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vars, ok := s.variables[envID]
	if !ok {
		return fmt.Errorf("environment not found: %s", envID)
	}

	delete(vars, key)
	delete(s.environments[envID].Variables, key)

	return nil
}

// CreateDeployment creates a new deployment
func (s *EnvironmentService) CreateDeployment(ctx context.Context, deployment *Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if deployment.EnvironmentID == "" {
		return fmt.Errorf("environment_id is required")
	}

	if _, ok := s.environments[deployment.EnvironmentID]; !ok {
		return fmt.Errorf("environment not found: %s", deployment.EnvironmentID)
	}

	now := time.Now()
	if deployment.ID == "" {
		deployment.ID = generateID()
	}
	deployment.Status = DeploymentStatusPending
	deployment.StartedAt = now

	s.deployments[deployment.EnvironmentID] = append(s.deployments[deployment.EnvironmentID], deployment)

	return nil
}

// StartDeployment marks a deployment as started
func (s *EnvironmentService) StartDeployment(deploymentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for envID, deployments := range s.deployments {
		for _, d := range deployments {
			if d.ID == deploymentID {
				d.Status = DeploymentStatusDeploying
				d.StartedAt = time.Now()
				s.environments[envID].Status = "deploying"
				return nil
			}
		}
	}
	return fmt.Errorf("deployment not found: %s", deploymentID)
}

// CompleteDeployment marks a deployment as completed
func (s *EnvironmentService) CompleteDeployment(deploymentID string, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for envID, deployments := range s.deployments {
		for _, d := range deployments {
			if d.ID == deploymentID {
				now := time.Now()
				d.CompletedAt = &now
				d.Duration = now.Sub(d.StartedAt).Milliseconds()

				if success {
					d.Status = DeploymentStatusSucceeded
					s.environments[envID].Status = "active"
					s.environments[envID].LastDeployedAt = &now
				} else {
					d.Status = DeploymentStatusFailed
					s.environments[envID].Status = "error"
				}
				return nil
			}
		}
	}
	return fmt.Errorf("deployment not found: %s", deploymentID)
}

// RollbackDeployment rolls back a deployment
func (s *EnvironmentService) RollbackDeployment(deploymentID, rollbackFrom string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for envID, deployments := range s.deployments {
		for _, d := range deployments {
			if d.ID == deploymentID {
				d.Status = DeploymentStatusRolledBack
				d.RollbackFrom = rollbackFrom

				// Find the deployment to roll back to
				for _, prev := range deployments {
					if prev.ID == rollbackFrom {
						s.environments[envID].Status = "active"
						s.environments[envID].LastDeployedAt = &prev.CompletedAt
						break
					}
				}
				return nil
			}
		}
	}
	return fmt.Errorf("deployment not found: %s", deploymentID)
}

// GetDeployments gets deployments for an environment
func (s *EnvironmentService) GetDeployments(envID string, limit int) []*Deployment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deployments := s.deployments[envID]
	if limit > 0 && len(deployments) > limit {
		deployments = deployments[:limit]
	}
	return deployments
}

// GetDeployment gets a specific deployment
func (s *EnvironmentService) GetDeployment(deploymentID string) (*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, deployments := range s.deployments {
		for _, d := range deployments {
			if d.ID == deploymentID {
				return d, nil
			}
		}
	}
	return nil, fmt.Errorf("deployment not found: %s", deploymentID)
}

// AddDeploymentLog adds a log entry to a deployment
func (s *EnvironmentService) AddDeploymentLog(deploymentID, log string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, deployments := range s.deployments {
		for _, d := range deployments {
			if d.ID == deploymentID {
				d.Logs = append(d.Logs, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), log))
				return nil
			}
		}
	}
	return fmt.Errorf("deployment not found: %s", deploymentID)
}

// PromoteEnvironment promotes from one environment to another
func (s *EnvironmentService) PromoteEnvironment(ctx context.Context, fromEnvID, toEnvID string) (*Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromEnv, ok := s.environments[fromEnvID]
	if !ok {
		return nil, fmt.Errorf("source environment not found: %s", fromEnvID)
	}

	toEnv, ok := s.environments[toEnvID]
	if !ok {
		return nil, fmt.Errorf("target environment not found: %s", toEnvID)
	}

	now := time.Now()
	deployment := &Deployment{
		ID:            generateID(),
		EnvironmentID: toEnvID,
		Status:        DeploymentStatusPending,
		StartedAt:     now,
		Changes:       []string{fmt.Sprintf("Promoted from %s", fromEnv.Name)},
	}

	s.deployments[toEnvID] = append(s.deployments[toEnvID], deployment)
	toEnv.Status = "deploying"

	return deployment, nil
}

// ToJSON serializes environment to JSON
func (e *Environment) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ToJSON serializes deployment to JSON
func (d *Deployment) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}