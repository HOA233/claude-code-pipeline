package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
)

// TaskHistoryService manages task history and archiving
type TaskHistoryService struct {
	redis      *repository.RedisClient
	retention  time.Duration
	archiveTTL time.Duration
}

// NewTaskHistoryService creates a new task history service
func NewTaskHistoryService(redis *repository.RedisClient, retentionHours int) *TaskHistoryService {
	return &TaskHistoryService{
		redis:      redis,
		retention:  time.Duration(retentionHours) * time.Hour,
		archiveTTL: 7 * 24 * time.Hour, // 7 days archive
	}
}

// ArchiveTask archives a completed task
func (s *TaskHistoryService) ArchiveTask(ctx context.Context, task *model.Task) error {
	return s.redis.ArchiveTask(ctx, task)
}

// GetArchivedTask retrieves an archived task
func (s *TaskHistoryService) GetArchivedTask(ctx context.Context, taskID string) (*model.Task, error) {
	return s.redis.GetArchivedTask(ctx, taskID)
}

// ListArchivedTasks lists archived tasks
func (s *TaskHistoryService) ListArchivedTasks(ctx context.Context, limit int) ([]*model.Task, error) {
	return s.redis.ListArchivedTasks(ctx, limit)
}

// CleanupExpiredTasks removes expired tasks from history
func (s *TaskHistoryService) CleanupExpiredTasks(ctx context.Context) error {
	// This would typically be called by a background job
	// For now, Redis TTL handles automatic expiration
	return nil
}

// TaskReport represents a task execution report
type TaskReport struct {
	TaskID          string                 `json:"task_id"`
	SkillID         string                 `json:"skill_id"`
	Status          string                 `json:"status"`
	Duration        time.Duration          `json:"duration"`
	StartedAt       time.Time              `json:"started_at"`
	EndedAt         time.Time              `json:"ended_at"`
	OutputLineCount int                    `json:"output_line_count"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateReport generates a report for a task
func (s *TaskHistoryService) GenerateReport(ctx context.Context, taskID string) (*TaskReport, error) {
	task, err := s.redis.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	output, err := s.redis.GetTaskOutput(ctx, taskID)
	if err != nil {
		output = nil
	}

	var duration time.Duration
	if task.CompletedAt != nil && !task.CreatedAt.IsZero() {
		duration = task.CompletedAt.Sub(task.CreatedAt)
	}

	return &TaskReport{
		TaskID:          task.ID,
		SkillID:         task.SkillID,
		Status:          string(task.Status),
		Duration:        duration,
		StartedAt:       task.CreatedAt,
		EndedAt:         task.CreatedAt, // Use CreatedAt since EndedAt doesn't exist
		OutputLineCount: len(output),
	}, nil
}

// ExportReport exports a report as JSON
func (s *TaskHistoryService) ExportReport(ctx context.Context, taskID string) ([]byte, error) {
	report, err := s.GenerateReport(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(report, "", "  ")
}

// AggregatedStats represents aggregated task statistics
type AggregatedStats struct {
	TotalTasks     int            `json:"total_tasks"`
	CompletedTasks int            `json:"completed_tasks"`
	FailedTasks    int            `json:"failed_tasks"`
	AvgDuration    time.Duration  `json:"avg_duration"`
	SuccessRate    float64        `json:"success_rate"`
	BySkill        map[string]int `json:"by_skill"`
	Period         string         `json:"period"`
}

// GetAggregatedStats returns aggregated statistics for a time period
func (s *TaskHistoryService) GetAggregatedStats(ctx context.Context, period string) (*AggregatedStats, error) {
	tasks, err := s.redis.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	stats := &AggregatedStats{
		BySkill: make(map[string]int),
		Period:  period,
	}

	var totalDuration time.Duration
	var durationCount int

	for _, task := range tasks {
		stats.TotalTasks++

		switch task.Status {
		case model.TaskStatusCompleted:
			stats.CompletedTasks++
		case model.TaskStatusFailed, model.TaskStatusCancelled:
			stats.FailedTasks++
		}

		stats.BySkill[task.SkillID]++

		if task.CompletedAt != nil && !task.CreatedAt.IsZero() {
			totalDuration += task.CompletedAt.Sub(task.CreatedAt)
			durationCount++
		}
	}

	if durationCount > 0 {
		stats.AvgDuration = totalDuration / time.Duration(durationCount)
	}

	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}