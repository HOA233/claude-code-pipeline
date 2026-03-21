package model

import (
	"encoding/json"
	"time"
)

// Pipeline represents an isolated CLI orchestration pipeline
// Each pipeline has its own skills, sessions, and data isolation
type Pipeline struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Mode        ExecutionMode     `json:"mode"`
	Steps       []Step            `json:"steps"`

	// Isolation: each pipeline has its own skill set
	Skills      map[string]*Skill `json:"skills,omitempty"`

	// Session isolation
	SessionID   string            `json:"session_id"`
	Context     map[string]interface{} `json:"context,omitempty"`

	// Configuration
	ErrorConfig *ErrorConfig      `json:"error_handling,omitempty"`
	Output      *OutputConfig     `json:"output,omitempty"`

	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// Tenant/Owner for multi-tenant isolation
	TenantID    string            `json:"tenant_id,omitempty"`
	Owner       string            `json:"owner,omitempty"`
}

// PipelineSession represents an isolated execution session
type PipelineSession struct {
	ID          string                 `json:"id"`
	PipelineID  string                 `json:"pipeline_id"`
	TenantID    string                 `json:"tenant_id"`

	// Isolated data store for this session
	Data        map[string]interface{} `json:"data"`

	// Isolated skill registry for this session
	SkillRegistry map[string]*Skill    `json:"skill_registry"`

	// Execution history (isolated)
	StepHistory []StepResult           `json:"step_history"`

	// Cross-step data sharing (within this session only)
	SharedData  map[string]json.RawMessage `json:"shared_data"`

	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// Step represents a single CLI step in a pipeline
type Step struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	CLI         string                 `json:"cli"`
	Action      string                 `json:"action"`
	Command     string                 `json:"command,omitempty"`
	Params      map[string]interface{} `json:"params"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	OnError     ErrorStrategy          `json:"on_error"`
	RetryCount  int                    `json:"retry_count"`
	Timeout     int                    `json:"timeout"`
	Condition   string                 `json:"condition,omitempty"`

	// Skill reference (uses pipeline's isolated skill registry)
	SkillID     string                 `json:"skill_id,omitempty"`

	// Data input/output mapping
	InputFrom   []string               `json:"input_from,omitempty"`  // which steps' output to use
	OutputTo    string                 `json:"output_to,omitempty"`   // variable name to store output
}

// Run represents a pipeline execution instance
type Run struct {
	ID          string            `json:"id"`
	PipelineID  string            `json:"pipeline_id"`
	SessionID   string            `json:"session_id"`
	TenantID    string            `json:"tenant_id"`

	Status      RunStatus         `json:"status"`
	StepResults []StepResult      `json:"step_results"`

	// Isolated output for this run
	Output      json.RawMessage   `json:"output,omitempty"`

	// Session data snapshot at run time
	SessionData map[string]interface{} `json:"session_data,omitempty"`

	Error       string            `json:"error,omitempty"`
	Duration    int64             `json:"duration"`

	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// RunStatus represents execution status
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

// StepResult represents a single step execution result
type StepResult struct {
	StepID    string          `json:"step_id"`
	Status    RunStatus       `json:"status"`
	Output    json.RawMessage `json:"output,omitempty"`
	Error     string          `json:"error,omitempty"`
	Duration  int64           `json:"duration"`
	StartedAt *time.Time      `json:"started_at,omitempty"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	Retries   int             `json:"retries"`
}

// ExecutionMode defines how steps are executed
type ExecutionMode string

const (
	ModeSerial   ExecutionMode = "serial"
	ModeParallel ExecutionMode = "parallel"
	ModeHybrid   ExecutionMode = "hybrid"
)

// ErrorStrategy defines how to handle errors
type ErrorStrategy string

const (
	ErrorContinue ErrorStrategy = "continue"
	ErrorStop     ErrorStrategy = "stop"
	ErrorRetry    ErrorStrategy = "retry"
)

// ErrorConfig defines pipeline-level error handling
type ErrorConfig struct {
	Retry       int    `json:"retry"`
	OnFailure   string `json:"on_failure"`
	Webhook     string `json:"webhook,omitempty"`
	NotifyEmail string `json:"notify_email,omitempty"`
}

// OutputConfig defines output handling
type OutputConfig struct {
	Format        string `json:"format"`
	MergeStrategy string `json:"merge_strategy"`
	SavePath      string `json:"save_path,omitempty"`
}

// PipelineCreateRequest for creating pipelines
type PipelineCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Mode        ExecutionMode          `json:"mode"`
	Steps       []Step                 `json:"steps" binding:"required"`
	Skills      map[string]*Skill      `json:"skills,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	ErrorConfig *ErrorConfig           `json:"error_handling,omitempty"`
	Output      *OutputConfig          `json:"output,omitempty"`
}

// RunCreateRequest for creating runs
type RunCreateRequest struct {
	PipelineID string                 `json:"pipeline_id" binding:"required"`
	Params     map[string]interface{} `json:"params"`
	Context    map[string]interface{} `json:"context"`
}