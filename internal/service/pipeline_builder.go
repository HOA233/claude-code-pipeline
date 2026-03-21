package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// PipelineBuilder provides a fluent API for constructing pipelines
type PipelineBuilder struct {
	pipeline *PipelineDefinition
	errors   []error
}

// PipelineDefinition represents a pipeline configuration
type PipelineDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Mode        string                 `json:"mode"` // serial, parallel, hybrid
	Steps       []StepDefinition       `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Triggers    []TriggerDefinition    `json:"triggers,omitempty"`
	ErrorConfig *ErrorConfig           `json:"error_config,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StepDefinition represents a step in a pipeline
type StepDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	CLI         string                 `json:"cli"`   // claude, npm, git, etc.
	Action      string                 `json:"action"` // analyze, review, deploy, etc.
	Command     string                 `json:"command,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Condition   string                 `json:"condition,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	Retry       *RetryConfig           `json:"retry,omitempty"`
	OnError     string                 `json:"on_error,omitempty"` // stop, continue, rollback
	InputFrom   string                 `json:"input_from,omitempty"`
	OutputTo    string                 `json:"output_to,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	WorkDir     string                 `json:"work_dir,omitempty"`
}

// TriggerDefinition defines pipeline triggers
type TriggerDefinition struct {
	Type      string                 `json:"type"` // webhook, schedule, event, manual
	Config    map[string]interface{} `json:"config,omitempty"`
	Enabled   bool                   `json:"enabled"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts int    `json:"max_attempts"`
	Delay       int    `json:"delay"`
	Backoff     string `json:"backoff"` // linear, exponential
}

// ErrorConfig defines pipeline-level error handling
type ErrorConfig struct {
	StopOnFailure  bool   `json:"stop_on_failure"`
	RetryCount     int    `json:"retry_count"`
	RetryDelay     int    `json:"retry_delay"`
	FailureWebhook string `json:"failure_webhook,omitempty"`
	NotifyEmail    string `json:"notify_email,omitempty"`
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: &PipelineDefinition{
			Steps:     []StepDefinition{},
			Variables: make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
		},
		errors: []error{},
	}
}

// WithID sets the pipeline ID
func (b *PipelineBuilder) WithID(id string) *PipelineBuilder {
	if id == "" {
		b.errors = append(b.errors, errors.New("pipeline ID cannot be empty"))
		return b
	}
	b.pipeline.ID = id
	return b
}

// WithName sets the pipeline name
func (b *PipelineBuilder) WithName(name string) *PipelineBuilder {
	if name == "" {
		b.errors = append(b.errors, errors.New("pipeline name cannot be empty"))
		return b
	}
	b.pipeline.Name = name
	return b
}

// WithDescription sets the pipeline description
func (b *PipelineBuilder) WithDescription(desc string) *PipelineBuilder {
	b.pipeline.Description = desc
	return b
}

// WithVersion sets the pipeline version
func (b *PipelineBuilder) WithVersion(version string) *PipelineBuilder {
	b.pipeline.Version = version
	return b
}

// WithMode sets the execution mode (serial, parallel, hybrid)
func (b *PipelineBuilder) WithMode(mode string) *PipelineBuilder {
	validModes := map[string]bool{"serial": true, "parallel": true, "hybrid": true}
	if !validModes[mode] {
		b.errors = append(b.errors, fmt.Errorf("invalid mode: %s", mode))
		return b
	}
	b.pipeline.Mode = mode
	return b
}

// WithTimeout sets the pipeline timeout in seconds
func (b *PipelineBuilder) WithTimeout(seconds int) *PipelineBuilder {
	if seconds < 0 {
		b.errors = append(b.errors, errors.New("timeout cannot be negative"))
		return b
	}
	b.pipeline.Timeout = seconds
	return b
}

// AddStep adds a step to the pipeline
func (b *PipelineBuilder) AddStep(step StepDefinition) *PipelineBuilder {
	if step.ID == "" {
		b.errors = append(b.errors, errors.New("step ID is required"))
		return b
	}
	if step.CLI == "" {
		b.errors = append(b.errors, fmt.Errorf("step %s: CLI is required", step.ID))
		return b
	}
	b.pipeline.Steps = append(b.pipeline.Steps, step)
	return b
}

// AddClaudeStep adds a Claude CLI step
func (b *PipelineBuilder) AddClaudeStep(id, action string, params map[string]interface{}) *PipelineBuilder {
	return b.AddStep(StepDefinition{
		ID:     id,
		CLI:    "claude",
		Action: action,
		Params: params,
	})
}

// AddGitStep adds a Git CLI step
func (b *PipelineBuilder) AddGitStep(id, command string, params map[string]interface{}) *PipelineBuilder {
	return b.AddStep(StepDefinition{
		ID:      id,
		CLI:     "git",
		Command: command,
		Params:  params,
	})
}

// AddNPMStep adds an NPM CLI step
func (b *PipelineBuilder) AddNPMStep(id, command string, params map[string]interface{}) *PipelineBuilder {
	return b.AddStep(StepDefinition{
		ID:      id,
		CLI:     "npm",
		Command: command,
		Params:  params,
	})
}

// AddDockerStep adds a Docker CLI step
func (b *PipelineBuilder) AddDockerStep(id, command string, params map[string]interface{}) *PipelineBuilder {
	return b.AddStep(StepDefinition{
		ID:      id,
		CLI:     "docker",
		Command: command,
		Params:  params,
	})
}

// AddVariable adds a variable to the pipeline
func (b *PipelineBuilder) AddVariable(key string, value interface{}) *PipelineBuilder {
	b.pipeline.Variables[key] = value
	return b
}

// AddWebhookTrigger adds a webhook trigger
func (b *PipelineBuilder) AddWebhookTrigger(path string) *PipelineBuilder {
	b.pipeline.Triggers = append(b.pipeline.Triggers, TriggerDefinition{
		Type: "webhook",
		Config: map[string]interface{}{
			"path": path,
		},
		Enabled: true,
	})
	return b
}

// AddScheduleTrigger adds a schedule trigger (cron)
func (b *PipelineBuilder) AddScheduleTrigger(cronExpr string) *PipelineBuilder {
	b.pipeline.Triggers = append(b.pipeline.Triggers, TriggerDefinition{
		Type: "schedule",
		Config: map[string]interface{}{
			"cron": cronExpr,
		},
		Enabled: true,
	})
	return b
}

// WithErrorConfig sets the error configuration
func (b *PipelineBuilder) WithErrorConfig(cfg *ErrorConfig) *PipelineBuilder {
	b.pipeline.ErrorConfig = cfg
	return b
}

// WithMetadata adds metadata to the pipeline
func (b *PipelineBuilder) WithMetadata(key string, value interface{}) *PipelineBuilder {
	b.pipeline.Metadata[key] = value
	return b
}

// Build validates and returns the pipeline
func (b *PipelineBuilder) Build() (*PipelineDefinition, error) {
	// Check for builder errors
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("pipeline builder errors: %v", b.errors)
	}

	// Validate required fields
	if b.pipeline.ID == "" {
		return nil, errors.New("pipeline ID is required")
	}
	if b.pipeline.Name == "" {
		return nil, errors.New("pipeline name is required")
	}
	if len(b.pipeline.Steps) == 0 {
		return nil, errors.New("pipeline must have at least one step")
	}

	// Set defaults
	if b.pipeline.Mode == "" {
		b.pipeline.Mode = "serial"
	}
	if b.pipeline.Version == "" {
		b.pipeline.Version = "1.0.0"
	}

	// Validate step dependencies
	stepIDs := make(map[string]bool)
	for _, step := range b.pipeline.Steps {
		stepIDs[step.ID] = true
	}

	for _, step := range b.pipeline.Steps {
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				return nil, fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	return b.pipeline, nil
}

// PipelineRegistry manages pipeline definitions
type PipelineRegistry struct {
	mu        sync.RWMutex
	pipelines map[string]*PipelineDefinition
}

// NewPipelineRegistry creates a new pipeline registry
func NewPipelineRegistry() *PipelineRegistry {
	return &PipelineRegistry{
		pipelines: make(map[string]*PipelineDefinition),
	}
}

// Register registers a pipeline definition
func (r *PipelineRegistry) Register(pipeline *PipelineDefinition) error {
	if pipeline.ID == "" {
		return errors.New("pipeline ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.pipelines[pipeline.ID] = pipeline
	return nil
}

// Get retrieves a pipeline by ID
func (r *PipelineRegistry) Get(id string) (*PipelineDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pipeline, exists := r.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", id)
	}
	return pipeline, nil
}

// List returns all pipeline definitions
func (r *PipelineRegistry) List() []*PipelineDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pipelines := make([]*PipelineDefinition, 0, len(r.pipelines))
	for _, p := range r.pipelines {
		pipelines = append(pipelines, p)
	}
	return pipelines
}

// Delete removes a pipeline
func (r *PipelineRegistry) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.pipelines, id)
	return nil
}

// ToJSON serializes a pipeline to JSON
func (p *PipelineDefinition) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ParsePipeline parses a pipeline from JSON
func ParsePipeline(data []byte) (*PipelineDefinition, error) {
	var pipeline PipelineDefinition
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}
	return &pipeline, nil
}

// Validate validates a pipeline definition
func (p *PipelineDefinition) Validate() error {
	if p.ID == "" {
		return errors.New("pipeline ID is required")
	}
	if p.Name == "" {
		return errors.New("pipeline name is required")
	}
	if len(p.Steps) == 0 {
		return errors.New("pipeline must have at least one step")
	}

	// Validate each step
	stepIDs := make(map[string]bool)
	for _, step := range p.Steps {
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
	}

	// Validate dependencies
	for _, step := range p.Steps {
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	return nil
}

// PipelineTemplate provides reusable pipeline templates
type PipelineTemplate struct {
	ID          string
	Name        string
	Description string
	Builder     func(params map[string]interface{}) *PipelineBuilder
}

// Built-in templates
var templates = map[string]*PipelineTemplate{
	"code-review": {
		ID:          "code-review",
		Name:        "Code Review Pipeline",
		Description: "Analyzes code for quality, security, and performance issues",
		Builder: func(params map[string]interface{}) *PipelineBuilder {
			target, _ := params["target"].(string)
			if target == "" {
				target = "./"
			}
			depth, _ := params["depth"].(string)
			if depth == "" {
				depth = "standard"
			}

			return NewPipelineBuilder().
				WithID("code-review").
				WithName("Code Review Pipeline").
				WithMode("serial").
				AddClaudeStep("analyze", "analyze", map[string]interface{}{
					"target": target,
					"depth":  depth,
				}).
				AddClaudeStep("review", "review", map[string]interface{}{
					"target": target,
				}).
				AddClaudeStep("report", "generate-report", map[string]interface{}{
					"format": "markdown",
				})
		},
	},
	"ci-cd": {
		ID:          "ci-cd",
		Name:        "CI/CD Pipeline",
		Description: "Runs tests, builds, and deploys the application",
		Builder: func(params map[string]interface{}) *PipelineBuilder {
			return NewPipelineBuilder().
				WithID("ci-cd").
				WithName("CI/CD Pipeline").
				WithMode("hybrid").
				AddNPMStep("install", "install", map[string]interface{}{}).
				AddNPMStep("lint", "run lint", map[string]interface{}{}).
				AddNPMStep("test", "run test", map[string]interface{}{}).
				AddNPMStep("build", "run build", map[string]interface{}{}).
				AddDockerStep("push", "push", map[string]interface{}{
					"image": params["image"],
				})
		},
	},
	"full-analysis": {
		ID:          "full-analysis",
		Name:        "Full Project Analysis",
		Description: "Comprehensive analysis including security, performance, and documentation",
		Builder: func(params map[string]interface{}) *PipelineBuilder {
			target, _ := params["target"].(string)
			if target == "" {
				target = "./"
			}

			return NewPipelineBuilder().
				WithID("full-analysis").
				WithName("Full Project Analysis").
				WithMode("parallel").
				AddClaudeStep("security", "security-scan", map[string]interface{}{
					"target": target,
				}).
				AddClaudeStep("performance", "performance-analysis", map[string]interface{}{
					"target": target,
				}).
				AddClaudeStep("docs", "generate-docs", map[string]interface{}{
					"target": target,
				}).
				AddClaudeStep("report", "combine-reports", map[string]interface{}{})
		},
	},
}

// GetTemplate retrieves a template by ID
func GetTemplate(id string) (*PipelineTemplate, error) {
	template, exists := templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates returns all available templates
func ListTemplates() []*PipelineTemplate {
	result := make([]*PipelineTemplate, 0, len(templates))
	for _, t := range templates {
		result = append(result, t)
	}
	return result
}

// BuildFromTemplate creates a pipeline from a template
func BuildFromTemplate(templateID string, params map[string]interface{}) (*PipelineDefinition, error) {
	template, err := GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	builder := template.Builder(params)
	return builder.Build()
}

// PipelineExecutor executes pipelines
type PipelineExecutor struct {
	registry   *PipelineRegistry
	eventBus   *EventBus
	runService *RunService
}

// NewPipelineExecutor creates a new pipeline executor
func NewPipelineExecutor(registry *PipelineRegistry, eventBus *EventBus, runService *RunService) *PipelineExecutor {
	return &PipelineExecutor{
		registry:   registry,
		eventBus:   eventBus,
		runService: runService,
	}
}

// Execute executes a pipeline by ID
func (e *PipelineExecutor) Execute(ctx context.Context, pipelineID string, params map[string]interface{}) (string, error) {
	pipeline, err := e.registry.Get(pipelineID)
	if err != nil {
		return "", err
	}

	// Create run
	run, err := e.runService.CreateRun(ctx, &struct {
		PipelineID string
		Params     map[string]interface{}
		Context    map[string]interface{}
	}{
		PipelineID: pipelineID,
		Params:     params,
	})
	if err != nil {
		return "", err
	}

	// Start execution in background
	go e.runService.StartRun(ctx, run.ID)

	return run.ID, nil
}