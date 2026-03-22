package model

import (
	"encoding/json"
	"time"
)

// Workflow 定义多个 Agent 如何协同工作
type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`

	// Agent 组合
	Agents      []AgentNode       `json:"agents"`

	// 执行流程
	Connections []Connection      `json:"connections,omitempty"`
	Mode        ExecutionMode     `json:"mode"`

	// 隔离配置
	SessionID   string            `json:"session_id,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`

	// 配置
	Context     map[string]interface{} `json:"context,omitempty"`
	ErrorHandling *ErrorConfig       `json:"error_handling,omitempty"`
	Output      *OutputConfig       `json:"output,omitempty"`

	// 状态
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AgentNode 代表 Agent 在工作流中的角色
type AgentNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	AgentID     string                 `json:"agent_id"`

	// 输入配置
	Input       map[string]interface{} `json:"input,omitempty"`
	InputFrom   map[string]string      `json:"input_from,omitempty"`

	// 输出配置
	OutputAs    string                 `json:"output_as,omitempty"`

	// 执行控制
	DependsOn   []string               `json:"depends_on,omitempty"`
	Condition   string                 `json:"condition,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	OnError     ErrorStrategy          `json:"on_error,omitempty"`
	RetryCount  int                    `json:"retry_count,omitempty"`

	// CLI 配置 (可选，直接执行)
	CLI         string                 `json:"cli,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Command     string                 `json:"command,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// Connection Agent 之间的数据流连接
type Connection struct {
	FromNode    string `json:"from_node"`
	FromOutput  string `json:"from_output"`
	ToNode      string `json:"to_node"`
	ToInput     string `json:"to_input"`
}

// WorkflowCreateRequest 创建 Workflow 请求
type WorkflowCreateRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	Agents        []AgentNode       `json:"agents" binding:"required"`
	Connections   []Connection      `json:"connections,omitempty"`
	Mode          ExecutionMode     `json:"mode"`
	Context       map[string]interface{} `json:"context,omitempty"`
	ErrorHandling *ErrorConfig      `json:"error_handling,omitempty"`
	Output        *OutputConfig     `json:"output,omitempty"`
}

// WorkflowUpdateRequest 更新 Workflow 请求
type WorkflowUpdateRequest struct {
	Name          string            `json:"name,omitempty"`
	Description   string            `json:"description,omitempty"`
	Agents        []AgentNode       `json:"agents,omitempty"`
	Connections   []Connection      `json:"connections,omitempty"`
	Mode          ExecutionMode     `json:"mode,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	ErrorHandling *ErrorConfig      `json:"error_handling,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
}

// WorkflowListResponse Workflow 列表响应
type WorkflowListResponse struct {
	Workflows []Workflow `json:"workflows"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

// WorkflowSession 工作流会话
type WorkflowSession struct {
	ID           string                     `json:"id"`
	WorkflowID   string                     `json:"workflow_id"`
	TenantID     string                     `json:"tenant_id"`

	// 隔离数据存储
	Data         map[string]interface{}     `json:"data"`

	// Agent 注册表
	AgentRegistry map[string]*Agent         `json:"agent_registry"`

	// 执行历史
	NodeHistory  []NodeResult               `json:"node_history"`

	// 跨节点数据共享
	SharedData   map[string]json.RawMessage `json:"shared_data"`

	CreatedAt    time.Time                  `json:"created_at"`
	ExpiresAt    *time.Time                 `json:"expires_at,omitempty"`
}