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

// SchedulerService manages scheduled tasks and recurring executions
type SchedulerService struct {
	redis       *repository.RedisClient
	taskSvc     *TaskService
	schedules   sync.Map
	stopChan    chan struct{}
	running     bool
	mu          sync.RWMutex
}

// Schedule represents a scheduled task configuration
type Schedule struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	SkillID      string                 `json:"skill_id"`
	Parameters   map[string]interface{} `json:"parameters"`
	CronExpr     string                 `json:"cron_expr"`
	Enabled      bool                   `json:"enabled"`
	LastRun      *time.Time             `json:"last_run,omitempty"`
	NextRun      *time.Time             `json:"next_run,omitempty"`
	RunCount     int                    `json:"run_count"`
	MaxRuns      int                    `json:"max_runs,omitempty"` // 0 = unlimited
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Tags         []string               `json:"tags,omitempty"`
	NotifyOnFail bool                   `json:"notify_on_fail"`
	WebhookURL   string                 `json:"webhook_url,omitempty"`
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(redis *repository.RedisClient, taskSvc *TaskService) *SchedulerService {
	return &SchedulerService{
		redis:    redis,
		taskSvc:  taskSvc,
		stopChan: make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start(ctx context.Context) {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	logger.Info("Scheduler service started")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Scheduler service stopped")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunScheduledTasks(ctx)
		}
	}
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	close(s.stopChan)
}

// CreateSchedule creates a new scheduled task
func (s *SchedulerService) CreateSchedule(ctx context.Context, schedule *Schedule) error {
	now := time.Now()
	schedule.ID = fmt.Sprintf("schedule-%d", now.UnixNano())
	schedule.CreatedAt = now
	schedule.UpdatedAt = now
	schedule.Enabled = true

	// Calculate next run time
	nextRun, err := s.calculateNextRun(schedule.CronExpr, now)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	schedule.NextRun = &nextRun

	// Save to Redis
	if err := s.saveSchedule(ctx, schedule); err != nil {
		return err
	}

	s.schedules.Store(schedule.ID, schedule)
	logger.Info(fmt.Sprintf("Created schedule: %s (%s)", schedule.Name, schedule.ID))

	return nil
}

// GetSchedule retrieves a schedule by ID
func (s *SchedulerService) GetSchedule(ctx context.Context, id string) (*Schedule, error) {
	// Check memory first
	if val, ok := s.schedules.Load(id); ok {
		return val.(*Schedule), nil
	}

	// Load from Redis
	return s.loadSchedule(ctx, id)
}

// ListSchedules returns all schedules
func (s *SchedulerService) ListSchedules(ctx context.Context) ([]*Schedule, error) {
	// Get all schedule keys from Redis
	keys, err := s.redis.ListScheduleKeys(ctx)
	if err != nil {
		return nil, err
	}

	schedules := make([]*Schedule, 0, len(keys))
	for _, key := range keys {
		schedule, err := s.loadSchedule(ctx, key)
		if err != nil {
			continue
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (s *SchedulerService) UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error {
	schedule, err := s.GetSchedule(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		schedule.Name = name
	}
	if cronExpr, ok := updates["cron_expr"].(string); ok {
		schedule.CronExpr = cronExpr
		nextRun, err := s.calculateNextRun(cronExpr, time.Now())
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		schedule.NextRun = &nextRun
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		schedule.Enabled = enabled
	}
	if params, ok := updates["parameters"].(map[string]interface{}); ok {
		schedule.Parameters = params
	}

	schedule.UpdatedAt = time.Now()

	if err := s.saveSchedule(ctx, schedule); err != nil {
		return err
	}

	s.schedules.Store(id, schedule)
	return nil
}

// DeleteSchedule removes a schedule
func (s *SchedulerService) DeleteSchedule(ctx context.Context, id string) error {
	s.schedules.Delete(id)
	return s.redis.DeleteSchedule(ctx, id)
}

// EnableSchedule enables a schedule
func (s *SchedulerService) EnableSchedule(ctx context.Context, id string) error {
	return s.UpdateSchedule(ctx, id, map[string]interface{}{"enabled": true})
}

// DisableSchedule disables a schedule
func (s *SchedulerService) DisableSchedule(ctx context.Context, id string) error {
	return s.UpdateSchedule(ctx, id, map[string]interface{}{"enabled": false})
}

// TriggerSchedule manually triggers a scheduled task
func (s *SchedulerService) TriggerSchedule(ctx context.Context, id string) (*model.Task, error) {
	schedule, err := s.GetSchedule(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.runScheduledTask(ctx, schedule)
}

// checkAndRunScheduledTasks checks all schedules and runs due tasks
func (s *SchedulerService) checkAndRunScheduledTasks(ctx context.Context) {
	now := time.Now()

	schedules, err := s.ListSchedules(ctx)
	if err != nil {
		logger.Error("Failed to list schedules: ", err)
		return
	}

	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		if schedule.NextRun != nil && schedule.NextRun.Before(now) || schedule.NextRun.Equal(now) {
			// Check max runs
			if schedule.MaxRuns > 0 && schedule.RunCount >= schedule.MaxRuns {
				schedule.Enabled = false
				s.saveSchedule(ctx, schedule)
				continue
			}

			// Run the task
			go s.runScheduledTask(ctx, schedule)
		}
	}
}

// runScheduledTask executes a scheduled task
func (s *SchedulerService) runScheduledTask(ctx context.Context, schedule *Schedule) (*model.Task, error) {
	logger.Info(fmt.Sprintf("Running scheduled task: %s", schedule.Name))

	task, err := s.taskSvc.CreateTask(ctx, &model.TaskCreateRequest{
		SkillID:    schedule.SkillID,
		Parameters: schedule.Parameters,
		Options: &model.TaskOptions{
			CallbackURL: schedule.WebhookURL,
		},
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create scheduled task: %v", err))
		return nil, err
	}

	// Update schedule stats
	now := time.Now()
	schedule.LastRun = &now
	schedule.RunCount++

	nextRun, err := s.calculateNextRun(schedule.CronExpr, now)
	if err != nil {
		logger.Error("Failed to calculate next run: ", err)
	} else {
		schedule.NextRun = &nextRun
	}

	s.saveSchedule(ctx, schedule)
	s.schedules.Store(schedule.ID, schedule)

	return task, nil
}

// calculateNextRun calculates the next run time based on cron expression
func (s *SchedulerService) calculateNextRun(cronExpr string, from time.Time) (time.Time, error) {
	// Simple cron parser for common patterns
	// Supports: "every N minutes/hours", specific times, etc.

	// For simplicity, implement basic patterns
	// Production would use a proper cron library

	switch {
	case cronExpr == "@hourly":
		return from.Add(1 * time.Hour).Truncate(time.Hour), nil
	case cronExpr == "@daily":
		return time.Date(from.Year(), from.Month(), from.Day()+1, 0, 0, 0, 0, from.Location()), nil
	case cronExpr == "@weekly":
		daysUntilSunday := 7 - int(from.Weekday())
		return time.Date(from.Year(), from.Month(), from.Day()+daysUntilSunday, 0, 0, 0, 0, from.Location()), nil
	default:
		// Default: run in 1 hour
		return from.Add(1 * time.Hour), nil
	}
}

// saveSchedule saves a schedule to Redis
func (s *SchedulerService) saveSchedule(ctx context.Context, schedule *Schedule) error {
	data, err := json.Marshal(schedule)
	if err != nil {
		return err
	}
	return s.redis.SaveSchedule(ctx, schedule.ID, data)
}

// loadSchedule loads a schedule from Redis
func (s *SchedulerService) loadSchedule(ctx context.Context, id string) (*Schedule, error) {
	data, err := s.redis.GetSchedule(ctx, id)
	if err != nil {
		return nil, err
	}

	var schedule Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, err
	}

	return &schedule, nil
}