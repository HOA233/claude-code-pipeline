package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/company/claude-pipeline/pkg/logger"
)

// AuditLogger handles audit logging
type AuditLogger struct {
	mu      sync.Mutex
	entries []AuditEntry
	redis   AuditStorage
	enabled bool
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	User        string                 `json:"user,omitempty"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Status      string                 `json:"status"` // success, failure
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
}

// AuditStorage defines the storage interface for audit logs
type AuditStorage interface {
	SaveAuditEntry(ctx context.Context, entry *AuditEntry) error
	GetAuditEntries(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// AuditFilter defines filters for querying audit logs
type AuditFilter struct {
	User       string
	Resource   string
	ResourceID string
	TenantID   string
	Action     string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(storage AuditStorage) *AuditLogger {
	return &AuditLogger{
		entries: make([]AuditEntry, 0),
		redis:   storage,
		enabled: true,
	}
}

// Log logs an audit entry
func (a *AuditLogger) Log(ctx context.Context, entry *AuditEntry) {
	if !a.enabled {
		return
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Store locally
	a.mu.Lock()
	a.entries = append(a.entries, *entry)
	// Keep last 1000 entries in memory
	if len(a.entries) > 1000 {
		a.entries = a.entries[len(a.entries)-1000:]
	}
	a.mu.Unlock()

	// Store in persistent storage
	if a.redis != nil {
		if err := a.redis.SaveAuditEntry(ctx, entry); err != nil {
			logger.Warnf("Failed to save audit entry: %v", err)
		}
	}

	// Log to application logs
	logAuditEntry(entry)
}

// LogTask logs a task-related action
func (a *AuditLogger) LogTask(ctx context.Context, action, taskID, status string, details map[string]interface{}) {
	a.Log(ctx, &AuditEntry{
		Action:     action,
		Resource:   "task",
		ResourceID: taskID,
		Status:     status,
		Details:    details,
	})
}

// LogPipeline logs a pipeline-related action
func (a *AuditLogger) LogPipeline(ctx context.Context, action, pipelineID, status string, details map[string]interface{}) {
	a.Log(ctx, &AuditEntry{
		Action:     action,
		Resource:   "pipeline",
		ResourceID: pipelineID,
		Status:     status,
		Details:    details,
	})
}

// LogRun logs a run-related action
func (a *AuditLogger) LogRun(ctx context.Context, action, runID, status string, duration int64, err error) {
	entry := &AuditEntry{
		Action:     action,
		Resource:   "run",
		ResourceID: runID,
		Status:     status,
		Duration:   duration,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	a.Log(ctx, entry)
}

// GetEntries returns audit entries from memory
func (a *AuditLogger) GetEntries(limit int) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	if limit <= 0 || limit > len(a.entries) {
		limit = len(a.entries)
	}

	// Return last N entries
	start := len(a.entries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]AuditEntry, limit)
	copy(result, a.entries[start:])
	return result
}

// Query queries audit entries from storage
func (a *AuditLogger) Query(ctx context.Context, filter AuditFilter) ([]AuditEntry, error) {
	if a.redis != nil {
		return a.redis.GetAuditEntries(ctx, filter)
	}
	return a.entries, nil
}

func logAuditEntry(entry *AuditEntry) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[AUDIT] %s ", entry.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("action=%s resource=%s/%s", entry.Action, entry.Resource, entry.ResourceID))

	if entry.User != "" {
		sb.WriteString(fmt.Sprintf(" user=%s", entry.User))
	}

	if entry.TenantID != "" {
		sb.WriteString(fmt.Sprintf(" tenant=%s", entry.TenantID))
	}

	sb.WriteString(fmt.Sprintf(" status=%s", entry.Status))

	if entry.Duration > 0 {
		sb.WriteString(fmt.Sprintf(" duration=%dms", entry.Duration))
	}

	if entry.Error != "" {
		sb.WriteString(fmt.Sprintf(" error=%s", entry.Error))
	}

	if entry.Status == "success" {
		logger.Info(sb.String())
	} else {
		logger.Warn(sb.String())
	}
}

// Audit actions
const (
	ActionTaskCreate   = "task.create"
	ActionTaskStart    = "task.start"
	ActionTaskComplete = "task.complete"
	ActionTaskFail     = "task.fail"
	ActionTaskCancel   = "task.cancel"

	ActionPipelineCreate = "pipeline.create"
	ActionPipelineUpdate = "pipeline.update"
	ActionPipelineDelete = "pipeline.delete"
	ActionPipelineRun    = "pipeline.run"

	ActionRunStart    = "run.start"
	ActionRunComplete = "run.complete"
	ActionRunFail     = "run.fail"
	ActionRunCancel   = "run.cancel"

	ActionSkillSync = "skill.sync"
	ActionSkillGet  = "skill.get"
)