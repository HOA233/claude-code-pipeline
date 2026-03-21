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

// Priority levels for tasks
type Priority int

const (
	PriorityLow      Priority = 1
	PriorityNormal   Priority = 5
	PriorityHigh     Priority = 10
	PriorityCritical Priority = 20
)

// TaskPriority extends Task with priority information
type TaskPriority struct {
	TaskID    string    `json:"task_id"`
	Priority  Priority  `json:"priority"`
	QueueName string    `json:"queue_name"`
	QueuedAt  time.Time `json:"queued_at"`
}

// PriorityQueue manages tasks with priority ordering
type PriorityQueue struct {
	redis    *repository.RedisClient
	mu       sync.RWMutex
	queues   map[string]*priorityHeap // queue_name -> heap
	stopChan chan struct{}
}

type priorityHeap struct {
	tasks []*TaskPriority
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(redis *repository.RedisClient) *PriorityQueue {
	pq := &PriorityQueue{
		redis:    redis,
		queues:   make(map[string]*priorityHeap),
		stopChan: make(chan struct{}),
	}

	// Initialize default queues
	pq.queues["default"] = &priorityHeap{}
	pq.queues["high"] = &priorityHeap{}
	pq.queues["low"] = &priorityHeap{}

	return pq
}

// Enqueue adds a task to the priority queue
func (pq *PriorityQueue) Enqueue(ctx context.Context, taskID string, priority Priority, queueName string) error {
	if queueName == "" {
		queueName = "default"
	}

	tp := &TaskPriority{
		TaskID:    taskID,
		Priority:  priority,
		QueueName: queueName,
		QueuedAt:  time.Now(),
	}

	// Save to Redis for persistence
	data, err := json.Marshal(tp)
	if err != nil {
		return err
	}

	// Use Redis sorted set for priority ordering
	// Score is negative priority so higher priority comes first
	score := float64(-priority) + float64(tp.QueuedAt.UnixNano())/1e18
	err = pq.redis.PushPriorityQueue(ctx, queueName, taskID, score)
	if err != nil {
		return err
	}

	// Also save task priority metadata
	pq.redis.SetTaskPriority(ctx, taskID, data)

	logger.Info(fmt.Sprintf("Task %s enqueued with priority %d in queue %s", taskID, priority, queueName))
	return nil
}

// Dequeue retrieves the highest priority task from a queue
func (pq *PriorityQueue) Dequeue(ctx context.Context, queueName string) (*TaskPriority, error) {
	if queueName == "" {
		queueName = "default"
	}

	taskID, err := pq.redis.PopPriorityQueue(ctx, queueName)
	if err != nil {
		return nil, err
	}

	if taskID == "" {
		return nil, nil
	}

	// Get task priority metadata
	data, err := pq.redis.GetTaskPriority(ctx, taskID)
	if err != nil {
		return nil, err
	}

	var tp TaskPriority
	if err := json.Unmarshal(data, &tp); err != nil {
		return nil, err
	}

	// Clean up metadata
	pq.redis.DeleteTaskPriority(ctx, taskID)

	return &tp, nil
}

// Peek returns the highest priority task without removing it
func (pq *PriorityQueue) Peek(ctx context.Context, queueName string) (*TaskPriority, error) {
	if queueName == "" {
		queueName = "default"
	}

	taskID, err := pq.redis.PeekPriorityQueue(ctx, queueName)
	if err != nil {
		return nil, err
	}

	if taskID == "" {
		return nil, nil
	}

	data, err := pq.redis.GetTaskPriority(ctx, taskID)
	if err != nil {
		return nil, err
	}

	var tp TaskPriority
	if err := json.Unmarshal(data, &tp); err != nil {
		return nil, err
	}

	return &tp, nil
}

// GetQueueLength returns the number of tasks in a queue
func (pq *PriorityQueue) GetQueueLength(ctx context.Context, queueName string) (int, error) {
	if queueName == "" {
		queueName = "default"
	}
	return pq.redis.GetPriorityQueueLength(ctx, queueName)
}

// GetAllQueueLengths returns the length of all queues
func (pq *PriorityQueue) GetAllQueueLengths(ctx context.Context) (map[string]int, error) {
	queues := []string{"default", "high", "low"}
	result := make(map[string]int)

	for _, q := range queues {
		length, err := pq.GetQueueLength(ctx, q)
		if err != nil {
			return nil, err
		}
		result[q] = length
	}

	return result, nil
}

// PromoteTask increases the priority of a queued task
func (pq *PriorityQueue) PromoteTask(ctx context.Context, taskID string, newPriority Priority) error {
	// Get current task priority
	data, err := pq.redis.GetTaskPriority(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found in queue: %w", err)
	}

	var tp TaskPriority
	if err := json.Unmarshal(data, &tp); err != nil {
		return err
	}

	// Remove from current position
	pq.redis.RemoveFromPriorityQueue(ctx, tp.QueueName, taskID)

	// Re-add with new priority
	tp.Priority = newPriority
	return pq.Enqueue(ctx, taskID, newPriority, tp.QueueName)
}

// MoveTask moves a task between queues
func (pq *PriorityQueue) MoveTask(ctx context.Context, taskID string, targetQueue string) error {
	data, err := pq.redis.GetTaskPriority(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found in queue: %w", err)
	}

	var tp TaskPriority
	if err := json.Unmarshal(data, &tp); err != nil {
		return err
	}

	// Remove from current queue
	pq.redis.RemoveFromPriorityQueue(ctx, tp.QueueName, taskID)

	// Add to target queue
	return pq.Enqueue(ctx, taskID, tp.Priority, targetQueue)
}

// ClearQueue removes all tasks from a queue
func (pq *PriorityQueue) ClearQueue(ctx context.Context, queueName string) error {
	return pq.redis.ClearPriorityQueue(ctx, queueName)
}

// GetQueueStats returns statistics about queues
func (pq *PriorityQueue) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	lengths, err := pq.GetAllQueueLengths(ctx)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, l := range lengths {
		total += l
	}

	return map[string]interface{}{
		"queues":      lengths,
		"total_tasks": total,
		"timestamp":   time.Now(),
	}, nil
}

// CreateTaskWithPriority creates a task with a specific priority
func (s *TaskService) CreateTaskWithPriority(ctx context.Context, req *model.TaskCreateRequest, priority Priority, queueName string) (*model.Task, error) {
	// Create the task normally
	task, err := s.CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	// Add to priority queue if priority queue is available
	// This is done separately to avoid circular dependency

	return task, nil
}

// String returns the string representation of a priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParsePriority parses a string to Priority
func ParsePriority(s string) Priority {
	switch s {
	case "low":
		return PriorityLow
	case "high":
		return PriorityHigh
	case "critical":
		return PriorityCritical
	default:
		return PriorityNormal
	}
}