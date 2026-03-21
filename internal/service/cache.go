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

// CacheService provides caching capabilities
type CacheService struct {
	redis   *repository.RedisClient
	local   *sync.Map
	ttl     time.Duration
	enabled bool
}

// NewCacheService creates a new cache service
func NewCacheService(redis *repository.RedisClient, ttl time.Duration) *CacheService {
	return &CacheService{
		redis:   redis,
		local:   &sync.Map{},
		ttl:     ttl,
		enabled: true,
	}
}

// cacheItem represents a cached item
type cacheItem struct {
	Value     []byte
	ExpiresAt time.Time
}

// Get retrieves a value from cache
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	// Try local cache first
	if item, ok := c.local.Load(key); ok {
		ci := item.(*cacheItem)
		if time.Now().Before(ci.ExpiresAt) {
			return json.Unmarshal(ci.Value, dest)
		}
		c.local.Delete(key)
	}

	// Try Redis
	data, err := c.redis.Get(ctx, "cache:"+key)
	if err != nil {
		return err
	}

	// Store in local cache
	c.local.Store(key, &cacheItem{
		Value:     data,
		ExpiresAt: time.Now().Add(c.ttl),
	})

	return json.Unmarshal(data, dest)
}

// Set stores a value in cache
func (c *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Store in Redis
	if err := c.redis.Set(ctx, "cache:"+key, data, c.ttl); err != nil {
		logger.Warnf("Failed to set Redis cache: %v", err)
	}

	// Store in local cache
	c.local.Store(key, &cacheItem{
		Value:     data,
		ExpiresAt: time.Now().Add(c.ttl),
	})

	return nil
}

// Delete removes a value from cache
func (c *CacheService) Delete(ctx context.Context, key string) error {
	c.local.Delete(key)
	return c.redis.Delete(ctx, "cache:"+key)
}

// Clear clears all cache
func (c *CacheService) Clear(ctx context.Context) error {
	c.local = &sync.Map{}
	return nil
}

// InvalidatePattern invalidates keys matching a pattern
func (c *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// Clear local cache for matching keys
	c.local.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if matchesPattern(k, pattern) {
				c.local.Delete(key)
			}
		}
		return true
	})

	return c.redis.DeleteByPattern(ctx, "cache:"+pattern)
}

func matchesPattern(s, pattern string) bool {
	// Simple prefix matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return len(s) >= len(pattern)-1 && s[:len(pattern)-1] == pattern[:len(pattern)-1]
	}
	return s == pattern
}

// GetOrSet returns cached value or computes and caches it
func (c *CacheService) GetOrSet(ctx context.Context, key string, dest interface{}, compute func() (interface{}, error)) error {
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	// Compute value
	value, err := compute()
	if err != nil {
		return err
	}

	// Cache it
	if err := c.Set(ctx, key, value); err != nil {
		logger.Warnf("Failed to cache computed value: %v", err)
	}

	// Set dest
	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

// SchedulerService handles scheduled tasks
type SchedulerService struct {
	redis    *repository.RedisClient
	tasks    map[string]*ScheduledTask
	mu       sync.RWMutex
	stopChan chan struct{}
}

// ScheduledTask represents a scheduled task
type ScheduledTask struct {
	ID        string
	PipelineID string
	Schedule  string
	Enabled   bool
	LastRun   *time.Time
	NextRun   *time.Time
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(redis *repository.RedisClient) *SchedulerService {
	return &SchedulerService{
		redis:    redis,
		tasks:    make(map[string]*ScheduledTask),
		stopChan: make(chan struct{}),
	}
}

// AddTask adds a scheduled task
func (s *SchedulerService) AddTask(ctx context.Context, task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate next run time
	nextRun, err := s.parseSchedule(task.Schedule)
	if err != nil {
		return err
	}
	task.NextRun = &nextRun

	s.tasks[task.ID] = task

	// Save to Redis
	return s.saveTask(ctx, task)
}

// RemoveTask removes a scheduled task
func (s *SchedulerService) RemoveTask(ctx context.Context, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, taskID)

	// Remove from Redis
	return s.redis.Delete(ctx, "schedule:"+taskID)
}

// Start starts the scheduler
func (s *SchedulerService) Start(ctx context.Context, orchestrator *Orchestrator) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	logger.Info("Scheduler started")

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRun(ctx, orchestrator)
		}
	}
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	close(s.stopChan)
}

func (s *SchedulerService) checkAndRun(ctx context.Context, orchestrator *Orchestrator) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	for _, task := range s.tasks {
		if !task.Enabled {
			continue
		}

		if task.NextRun != nil && now.After(*task.NextRun) {
			// Run the pipeline
			go func(t *ScheduledTask) {
				req := &model.RunCreateRequest{
					PipelineID: t.PipelineID,
				}
				_, err := orchestrator.RunPipeline(ctx, req)
				if err != nil {
					logger.Errorf("Failed to run scheduled pipeline %s: %v", t.PipelineID, err)
				}
			}(task)

			// Update last run and calculate next run
			task.LastRun = &now
			nextRun, _ := s.parseSchedule(task.Schedule)
			task.NextRun = &nextRun

			// Save updated task
			s.saveTask(ctx, task)
		}
	}
}

func (s *SchedulerService) parseSchedule(schedule string) (time.Time, error) {
	// Simple interval parsing (e.g., "1h", "30m", "24h")
	duration, err := time.ParseDuration(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid schedule: %w", err)
	}
	return time.Now().Add(duration), nil
}

func (s *SchedulerService) saveTask(ctx context.Context, task *ScheduledTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, "schedule:"+task.ID, data, 0)
}

// ListTasks lists all scheduled tasks
func (s *SchedulerService) ListTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}