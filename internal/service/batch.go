package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
)

// BatchOperation represents a batch operation request
type BatchOperation struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Operations  []BatchTask            `json:"operations"`
	Status      string                 `json:"status"` // pending, running, completed, failed, partial
	Results     []BatchResult          `json:"results,omitempty"`
	TotalCount  int                    `json:"total_count"`
	SuccessCount int                   `json:"success_count"`
	FailCount   int                    `json:"fail_count"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Options     BatchOptions           `json:"options"`
}

// BatchTask represents a single task in a batch
type BatchTask struct {
	ID         string                 `json:"id"`
	SkillID    string                 `json:"skill_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// BatchResult represents the result of a single batch task
type BatchResult struct {
	TaskID    string          `json:"task_id"`
	Status    string          `json:"status"`
	TaskRef   string          `json:"task_ref,omitempty"`
	Error     string          `json:"error,omitempty"`
	Duration  int64           `json:"duration,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
}

// BatchOptions contains options for batch operations
type BatchOptions struct {
	StopOnError    bool   `json:"stop_on_error"`
	MaxConcurrency int    `json:"max_concurrency"`
	Timeout        int    `json:"timeout"`
	NotifyOnComplete bool  `json:"notify_on_complete"`
	CallbackURL    string `json:"callback_url,omitempty"`
}

// BatchService handles batch operations
type BatchService struct {
	redis      *repository.RedisClient
	taskSvc    *TaskService
	activeOps  sync.Map
}

// NewBatchService creates a new batch service
func NewBatchService(redis *repository.RedisClient, taskSvc *TaskService) *BatchService {
	return &BatchService{
		redis:   redis,
		taskSvc: taskSvc,
	}
}

// CreateBatch creates a new batch operation
func (s *BatchService) CreateBatch(ctx context.Context, name string, operations []BatchTask, options BatchOptions) (*BatchOperation, error) {
	if len(operations) == 0 {
		return nil, fmt.Errorf("batch must contain at least one operation")
	}

	batch := &BatchOperation{
		ID:         fmt.Sprintf("batch-%d", time.Now().UnixNano()),
		Name:       name,
		Operations: operations,
		Status:     "pending",
		TotalCount: len(operations),
		CreatedAt:  time.Now(),
		Options:    options,
	}

	if err := s.saveBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// ExecuteBatch executes a batch operation
func (s *BatchService) ExecuteBatch(ctx context.Context, batchID string) (*BatchOperation, error) {
	batch, err := s.getBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	if batch.Status != "pending" {
		return nil, fmt.Errorf("batch is not in pending state")
	}

	batch.Status = "running"
	s.saveBatch(ctx, batch)
	s.activeOps.Store(batchID, true)

	// Execute based on concurrency setting
	maxConcurrency := batch.Options.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, op := range batch.Operations {
		// Check if we should stop
		if _, ok := s.activeOps.Load(batchID); !ok {
			break
		}

		wg.Add(1)
		go func(idx int, task BatchTask) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.executeSingleTask(ctx, task)
			result.TaskRef = task.ID

			mu.Lock()
			batch.Results = append(batch.Results, result)
			if result.Status == "success" {
				batch.SuccessCount++
			} else {
				batch.FailCount++
				if batch.Options.StopOnError {
					s.activeOps.Delete(batchID)
				}
			}
			mu.Unlock()
		}(i, op)
	}

	wg.Wait()

	// Update final status
	mu.Lock()
	if batch.FailCount == 0 {
		batch.Status = "completed"
	} else if batch.SuccessCount == 0 {
		batch.Status = "failed"
	} else {
		batch.Status = "partial"
	}
	now := time.Now()
	batch.CompletedAt = &now
	mu.Unlock()

	s.activeOps.Delete(batchID)
	s.saveBatch(ctx, batch)

	// Send notification if configured
	if batch.Options.CallbackURL != "" {
		go s.sendCallback(batch)
	}

	logger.Info(fmt.Sprintf("Batch %s completed: %d success, %d failed", batchID, batch.SuccessCount, batch.FailCount))

	return batch, nil
}

// executeSingleTask executes a single task within the batch
func (s *BatchService) executeSingleTask(ctx context.Context, task BatchTask) BatchResult {
	start := time.Now()
	result := BatchResult{
		TaskID: fmt.Sprintf("task-%d", time.Now().UnixNano()),
	}

	// Create the task
	createdTask, err := s.taskSvc.CreateTask(ctx, &model.TaskCreateRequest{
		SkillID:    task.SkillID,
		Parameters: task.Parameters,
		Context:    task.Context,
	})

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result
	}

	result.TaskID = createdTask.ID

	// Wait for task completion (simplified - in production would use callbacks/polling)
	timeout := 5 * time.Minute
	if s.taskSvc.redis != nil {
		// Poll for completion
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				result.Status = "failed"
				result.Error = "context cancelled"
				return result
			case <-time.After(timeout):
				result.Status = "failed"
				result.Error = "timeout waiting for task"
				return result
			case <-ticker.C:
				t, err := s.taskSvc.GetTask(ctx, createdTask.ID)
				if err != nil {
					continue
				}
				if t.Status == model.TaskStatusCompleted {
					result.Status = "success"
					result.Duration = time.Since(start).Milliseconds()
					result.Result = t.Result
					return result
				}
				if t.Status == model.TaskStatusFailed {
					result.Status = "failed"
					result.Error = t.Error
					result.Duration = time.Since(start).Milliseconds()
					return result
				}
			}
		}
	}

	result.Status = "success"
	result.Duration = time.Since(start).Milliseconds()
	return result
}

// GetBatch retrieves a batch operation
func (s *BatchService) GetBatch(ctx context.Context, batchID string) (*BatchOperation, error) {
	return s.getBatch(ctx, batchID)
}

// CancelBatch cancels a running batch operation
func (s *BatchService) CancelBatch(ctx context.Context, batchID string) error {
	s.activeOps.Delete(batchID)

	batch, err := s.getBatch(ctx, batchID)
	if err != nil {
		return err
	}

	if batch.Status != "running" {
		return fmt.Errorf("batch is not running")
	}

	batch.Status = "cancelled"
	now := time.Now()
	batch.CompletedAt = &now
	return s.saveBatch(ctx, batch)
}

// ListBatches lists all batch operations
func (s *BatchService) ListBatches(ctx context.Context) ([]*BatchOperation, error) {
	keys, err := s.redis.ListBatchKeys(ctx)
	if err != nil {
		return nil, err
	}

	batches := make([]*BatchOperation, 0, len(keys))
	for _, key := range keys {
		batch, err := s.getBatch(ctx, key)
		if err != nil {
			continue
		}
		batches = append(batches, batch)
	}

	return batches, nil
}

// DeleteBatch deletes a batch operation
func (s *BatchService) DeleteBatch(ctx context.Context, batchID string) error {
	return s.redis.DeleteBatch(ctx, batchID)
}

// GetBatchStats returns statistics about batch operations
func (s *BatchService) GetBatchStats(ctx context.Context) (map[string]interface{}, error) {
	batches, err := s.ListBatches(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":     len(batches),
		"pending":   0,
		"running":   0,
		"completed": 0,
		"failed":    0,
		"partial":   0,
	}

	for _, b := range batches {
		if count, ok := stats[b.Status].(int); ok {
			stats[b.Status] = count + 1
		}
	}

	return stats, nil
}

func (s *BatchService) saveBatch(ctx context.Context, batch *BatchOperation) error {
	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}
	return s.redis.SaveBatch(ctx, batch.ID, data)
}

func (s *BatchService) getBatch(ctx context.Context, batchID string) (*BatchOperation, error) {
	data, err := s.redis.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	var batch BatchOperation
	if err := json.Unmarshal(data, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

func (s *BatchService) sendCallback(batch *BatchOperation) {
	// Implementation would send HTTP POST to callback URL
	logger.Info(fmt.Sprintf("Sending callback for batch %s to %s", batch.ID, batch.Options.CallbackURL))
}