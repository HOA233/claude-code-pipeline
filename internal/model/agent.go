package model

import (
	"encoding/json"
	"time"
)

// Agent 配置好的 Claude Code CLI 实例
type Agent struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`

	// Claude Code CLI 配置
	Model       string          `json:"model"`
	SystemPrompt string         `json:"system_prompt"`
	MaxTokens   int             `json:"max_tokens"`

	// 技能选择
	Skills      []SkillRef      `json:"skills"`
	DefaultSkill string         `json:"default_skill,omitempty"`

	// 能力配置
	Tools       []Tool          `json:"tools"`
	Permissions []Permission    `json:"permissions"`

	// 输入/输出 Schema
	InputSchema  json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`

	// 行为配置
	Timeout     int             `json:"timeout"`
	RetryPolicy RetryPolicy     `json:"retry_policy,omitempty"`

	// 隔离配置
	Isolation   IsolationConfig `json:"isolation"`

	// 元数据
	Tags        []string        `json:"tags,omitempty"`
	Category    string          `json:"category,omitempty"`
	Version     string          `json:"version"`

	// 状态
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	TenantID    string          `json:"tenant_id,omitempty"`
}

// SkillRef 定义 Agent 可引用的技能
type SkillRef struct {
	SkillID      string            `json:"skill_id"`
	Alias        string            `json:"alias,omitempty"`
	InputMapping map[string]string `json:"input_mapping,omitempty"`
	OutputMapping map[string]string `json:"output_mapping,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// Permission 权限定义
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries int    `json:"max_retries"`
	Backoff    string `json:"backoff"` // "linear", "exponential"
	MaxDelay   int    `json:"max_delay"` // seconds
}

// IsolationConfig 隔离配置
type IsolationConfig struct {
	DataIsolation    bool   `json:"data_isolation"`
	SessionIsolation bool   `json:"session_isolation"`
	NetworkIsolation bool   `json:"network_isolation"`
	FileIsolation    bool   `json:"file_isolation"`
	Namespace        string `json:"namespace,omitempty"`
}

// AgentCreateRequest 创建 Agent 请求
type AgentCreateRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Model        string            `json:"model"`
	SystemPrompt string            `json:"system_prompt"`
	MaxTokens    int               `json:"max_tokens"`
	Skills       []SkillRef        `json:"skills"`
	Tools        []Tool            `json:"tools"`
	Permissions  []Permission      `json:"permissions"`
	InputSchema  json.RawMessage   `json:"input_schema"`
	OutputSchema json.RawMessage   `json:"output_schema"`
	Timeout      int               `json:"timeout"`
	Isolation    IsolationConfig   `json:"isolation"`
	Tags         []string          `json:"tags"`
	Category     string            `json:"category"`
}

// AgentUpdateRequest 更新 Agent 请求
type AgentUpdateRequest struct {
	Name         string            `json:"name,omitempty"`
	Description  string            `json:"description,omitempty"`
	Model        string            `json:"model,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Skills       []SkillRef        `json:"skills,omitempty"`
	Tools        []Tool            `json:"tools,omitempty"`
	Permissions  []Permission      `json:"permissions,omitempty"`
	Timeout      int               `json:"timeout,omitempty"`
	Isolation    IsolationConfig   `json:"isolation,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Enabled      *bool             `json:"enabled,omitempty"`
}

// AgentExecuteRequest 执行 Agent 请求
type AgentExecuteRequest struct {
	Input    map[string]interface{} `json:"input"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Async    bool                   `json:"async"`
	Callback string                 `json:"callback,omitempty"`
}