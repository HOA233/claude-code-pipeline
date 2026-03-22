package model

import (
	"time"
)

// ScheduledJob 定时任务
type ScheduledJob struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`

	// 执行目标
	TargetType  string                 `json:"target_type"` // "agent" 或 "workflow"
	TargetID    string                 `json:"target_id"`
	TargetName  string                 `json:"target_name,omitempty"`

	// 调度配置
	Cron        string                 `json:"cron"`
	Timezone    string                 `json:"timezone"`
	Enabled     bool                   `json:"enabled"`

	// 执行参数
	Input       map[string]interface{} `json:"input,omitempty"`

	// 执行历史
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	RunCount    int                    `json:"run_count"`
	LastStatus  string                 `json:"last_status,omitempty"`

	// 失败处理
	OnFailure   string                 `json:"on_failure"` // "notify", "retry", "disable"
	NotifyEmail string                 `json:"notify_email,omitempty"`
	RetryCount  int                    `json:"retry_count,omitempty"`

	// 元数据
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	TenantID    string                 `json:"tenant_id,omitempty"`

	// 状态
	LastError   string                 `json:"last_error,omitempty"`
}

// ScheduledJobCreateRequest 创建定时任务请求
type ScheduledJobCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	TargetType  string                 `json:"target_type" binding:"required"`
	TargetID    string                 `json:"target_id" binding:"required"`
	Cron        string                 `json:"cron" binding:"required"`
	Timezone    string                 `json:"timezone"`
	Input       map[string]interface{} `json:"input,omitempty"`
	OnFailure   string                 `json:"on_failure"`
	NotifyEmail string                 `json:"notify_email,omitempty"`
}

// ScheduledJobUpdateRequest 更新定时任务请求
type ScheduledJobUpdateRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Cron        string                 `json:"cron,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	OnFailure   string                 `json:"on_failure,omitempty"`
	NotifyEmail string                 `json:"notify_email,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
}

// ScheduledJobListResponse 定时任务列表响应
type ScheduledJobListResponse struct {
	Jobs     []ScheduledJob `json:"jobs"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// JobExecutionHistory 任务执行历史
type JobExecutionHistory struct {
	ID           string    `json:"id"`
	JobID        string    `json:"job_id"`
	ExecutionID  string    `json:"execution_id"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string    `json:"error,omitempty"`
	Duration     int64     `json:"duration"`
}

// JobExecutionHistoryListResponse 执行历史列表响应
type JobExecutionHistoryListResponse struct {
	History  []JobExecutionHistory `json:"history"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}