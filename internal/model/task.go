package model

import (
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID          string          `json:"id"`
	SkillID     string          `json:"skill_id"`
	Status      TaskStatus      `json:"status"`
	Parameters  json.RawMessage `json:"parameters"`
	Context     json.RawMessage `json:"context,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
	Duration    int64           `json:"duration,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type TaskCreateRequest struct {
	SkillID    string                 `json:"skill_id" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Options    *TaskOptions           `json:"options,omitempty"`
}

type TaskOptions struct {
	Timeout     int    `json:"timeout,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}