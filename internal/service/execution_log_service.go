package service

import (
	"context"
	"time"
)

// ExecutionLog represents a log entry for an execution
type ExecutionLog struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"execution_id"`
	Level       string    `json:"level"` // info, warn, error
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source,omitempty"`
}

// ExecutionLogService handles execution logs
type ExecutionLogService struct{}

// NewExecutionLogService creates a new execution log service
func NewExecutionLogService(redis interface{}) *ExecutionLogService {
	return &ExecutionLogService{}
}

// GetLogs retrieves logs for an execution
func (s *ExecutionLogService) GetLogs(ctx context.Context, executionID string, limit int64) ([]ExecutionLog, error) {
	// Return mock logs for now
	return []ExecutionLog{
		{
			ID:          "log-1",
			ExecutionID: executionID,
			Level:       "info",
			Message:     "Execution started",
			Timestamp:   time.Now().Add(-5 * time.Minute),
			Source:      "scheduler",
		},
		{
			ID:          "log-2",
			ExecutionID: executionID,
			Level:       "info",
			Message:     "Running workflow steps",
			Timestamp:   time.Now().Add(-4 * time.Minute),
			Source:      "executor",
		},
		{
			ID:          "log-3",
			ExecutionID: executionID,
			Level:       "info",
			Message:     "Step 1 completed successfully",
			Timestamp:   time.Now().Add(-3 * time.Minute),
			Source:      "step-1",
		},
	}, nil
}

// StreamLogs streams logs for an execution
func (s *ExecutionLogService) StreamLogs(ctx context.Context, executionID string) <-chan ExecutionLog {
	ch := make(chan ExecutionLog, 10)

	go func() {
		defer close(ch)
		// Send initial log
		ch <- ExecutionLog{
			ID:          "stream-1",
			ExecutionID: executionID,
			Level:       "info",
			Message:     "Connected to log stream",
			Timestamp:   time.Now(),
		}
	}()

	return ch
}