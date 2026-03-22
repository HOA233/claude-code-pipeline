package model

import (
	"encoding/json"
	"time"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusPaused    ExecutionStatus = "paused"
)

// Execution 工作流执行实例
type Execution struct {
	ID          string                  `json:"id"`
	WorkflowID  string                  `json:"workflow_id"`
	WorkflowName string                 `json:"workflow_name"`
	SessionID   string                  `json:"session_id"`
	TenantID    string                  `json:"tenant_id,omitempty"`

	Status      ExecutionStatus         `json:"status"`
	Progress    int                     `json:"progress"` // 0-100

	// 步骤状态
	CurrentStep     string              `json:"current_step,omitempty"`
	TotalSteps      int                 `json:"total_steps"`
	CompletedSteps  int                 `json:"completed_steps"`
	NodeResults     map[string]NodeResult `json:"node_results"`

	// 结果
	FinalOutput json.RawMessage         `json:"final_output,omitempty"`

	// 元数据
	Duration    int64                   `json:"duration"`
	Error       string                  `json:"error,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	StartedAt   *time.Time              `json:"started_at,omitempty"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// NodeResult 节点执行结果
type NodeResult struct {
	NodeID      string          `json:"node_id"`
	AgentID     string          `json:"agent_id"`
	Status      ExecutionStatus `json:"status"`
	Output      json.RawMessage `json:"output,omitempty"`
	Error       string          `json:"error,omitempty"`
	Duration    int64           `json:"duration"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Retries     int             `json:"retries"`
}

// ExecutionCreateRequest 创建执行请求
type ExecutionCreateRequest struct {
	WorkflowID string                 `json:"workflow_id" binding:"required"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Async      bool                   `json:"async"`
	Callback   string                 `json:"callback,omitempty"`
}

// ExecutionListResponse 执行列表响应
type ExecutionListResponse struct {
	Executions []Execution `json:"executions"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// ExecutionFilter 执行过滤条件
type ExecutionFilter struct {
	Status     ExecutionStatus `json:"status,omitempty"`
	WorkflowID string          `json:"workflow_id,omitempty"`
	TenantID   string          `json:"tenant_id,omitempty"`
	StartDate  *time.Time      `json:"start_date,omitempty"`
	EndDate    *time.Time      `json:"end_date,omitempty"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

// SSEExecutionUpdate SSE 更新事件
type SSEExecutionUpdate struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// ExecutionProgressUpdate 进度更新
type ExecutionProgressUpdate struct {
	ExecutionID  string              `json:"execution_id"`
	Status       ExecutionStatus     `json:"status"`
	Progress     int                 `json:"progress"`
	CurrentStep  string              `json:"current_step,omitempty"`
	StepStatus   map[string]string   `json:"step_status,omitempty"`
	Error        string              `json:"error,omitempty"`
}