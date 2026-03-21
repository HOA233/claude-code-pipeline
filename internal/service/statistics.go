package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
)

// StatisticsService provides system statistics and metrics
type StatisticsService struct {
	redis      *repository.RedisClient
	startTime  time.Time
	mu         sync.RWMutex
	counters   map[string]int64
}

// SystemStats represents comprehensive system statistics
type SystemStats struct {
	Uptime           string                 `json:"uptime"`
	UptimeSeconds    int64                  `json:"uptime_seconds"`
	StartTime        time.Time              `json:"start_time"`
	Tasks            TaskStats              `json:"tasks"`
	Pipelines        PipelineStats          `json:"pipelines"`
	Skills           SkillStats             `json:"skills"`
	Queues           map[string]interface{} `json:"queues"`
	Performance      PerformanceStats       `json:"performance"`
	System           SystemInfo             `json:"system"`
}

// TaskStats contains task-related statistics
type TaskStats struct {
	Total       int64 `json:"total"`
	Pending     int64 `json:"pending"`
	Running     int64 `json:"running"`
	Completed   int64 `json:"completed"`
	Failed      int64 `json:"failed"`
	Cancelled   int64 `json:"cancelled"`
	AvgDuration int64 `json:"avg_duration_ms"`
}

// PipelineStats contains pipeline-related statistics
type PipelineStats struct {
	Total      int64 `json:"total"`
	Runs       int64 `json:"runs"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

// SkillStats contains skill-related statistics
type SkillStats struct {
	Total     int   `json:"total"`
	Enabled   int   `json:"enabled"`
	ByCategory map[string]int `json:"by_category"`
}

// PerformanceStats contains performance metrics
type PerformanceStats struct {
	TasksPerMinute    float64 `json:"tasks_per_minute"`
	AvgResponseTime   int64   `json:"avg_response_time_ms"`
	QueueWaitTime     int64   `json:"queue_wait_time_ms"`
	SuccessRate       float64 `json:"success_rate"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	CPUUsage     string `json:"cpu_usage"`
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(redis *repository.RedisClient) *StatisticsService {
	return &StatisticsService{
		redis:     redis,
		startTime: time.Now(),
		counters:  make(map[string]int64),
	}
}

// GetStats returns comprehensive system statistics
func (s *StatisticsService) GetStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{
		StartTime: s.startTime,
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
		Uptime: formatDuration(time.Since(s.startTime)),
	}

	// Get task stats
	taskStats, err := s.getTaskStats(ctx)
	if err == nil {
		stats.Tasks = *taskStats
	}

	// Get pipeline stats
	pipelineStats, err := s.getPipelineStats(ctx)
	if err == nil {
		stats.Pipelines = *pipelineStats
	}

	// Get skill stats
	skillStats, err := s.getSkillStats(ctx)
	if err == nil {
		stats.Skills = *skillStats
	}

	// Get queue stats
	queueStats, err := s.getQueueStats(ctx)
	if err == nil {
		stats.Queues = queueStats
	}

	// Get performance stats
	perfStats := s.getPerformanceStats()
	stats.Performance = *perfStats

	// Get system info
	stats.System = getSystemInfo()

	return stats, nil
}

// getTaskStats retrieves task statistics
func (s *StatisticsService) getTaskStats(ctx context.Context) (*TaskStats, error) {
	stats := &TaskStats{}

	// Get tasks from Redis
	tasks, err := s.redis.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	var totalDuration int64
	var durationCount int64

	for _, task := range tasks {
		stats.Total++
		switch task.Status {
		case "pending":
			stats.Pending++
		case "running":
			stats.Running++
		case "completed":
			stats.Completed++
			if task.Duration > 0 {
				totalDuration += task.Duration
				durationCount++
			}
		case "failed":
			stats.Failed++
		case "cancelled":
			stats.Cancelled++
		}
	}

	if durationCount > 0 {
		stats.AvgDuration = totalDuration / durationCount
	}

	return stats, nil
}

// getPipelineStats retrieves pipeline statistics
func (s *StatisticsService) getPipelineStats(ctx context.Context) (*PipelineStats, error) {
	stats := &PipelineStats{}

	pipelines, err := s.redis.GetAllPipelines(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = int64(len(pipelines))

	runs, err := s.redis.GetAllRuns(ctx)
	if err != nil {
		return nil, err
	}

	for _, run := range runs {
		stats.Runs++
		if run.Status == "completed" {
			stats.Successful++
		} else if run.Status == "failed" {
			stats.Failed++
		}
	}

	return stats, nil
}

// getSkillStats retrieves skill statistics
func (s *StatisticsService) getSkillStats(ctx context.Context) (*SkillStats, error) {
	stats := &SkillStats{
		ByCategory: make(map[string]int),
	}

	skills, err := s.redis.GetAllSkills(ctx)
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		stats.Total++
		if skill.Enabled {
			stats.Enabled++
		}
		if skill.Category != "" {
			stats.ByCategory[skill.Category]++
		}
	}

	return stats, nil
}

// getQueueStats retrieves queue statistics
func (s *StatisticsService) getQueueStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Task queue
	taskQueueLen, err := s.redis.GetQueueLength(ctx)
	if err == nil {
		stats["task_queue"] = taskQueueLen
	}

	// Run queue
	runQueueLen, err := s.redis.GetRunQueueLength(ctx)
	if err == nil {
		stats["run_queue"] = runQueueLen
	}

	// Priority queues
	pq := NewPriorityQueue(s.redis)
	pqStats, err := pq.GetAllQueueLengths(ctx)
	if err == nil {
		stats["priority_queues"] = pqStats
	}

	return stats, nil
}

// getPerformanceStats calculates performance metrics
func (s *StatisticsService) getPerformanceStats() *PerformanceStats {
	stats := &PerformanceStats{}

	// Calculate tasks per minute
	if s.counters["tasks_total"] > 0 {
		uptimeMinutes := float64(time.Since(s.startTime).Minutes())
		if uptimeMinutes > 0 {
			stats.TasksPerMinute = float64(s.counters["tasks_total"]) / uptimeMinutes
		}
	}

	// Calculate success rate
	if s.counters["tasks_total"] > 0 {
		stats.SuccessRate = float64(s.counters["tasks_completed"]) / float64(s.counters["tasks_total"]) * 100
	}

	stats.AvgResponseTime = s.counters["avg_response_time"]
	stats.QueueWaitTime = s.counters["avg_queue_wait"]

	return stats
}

// RecordTaskCreated records a task creation event
func (s *StatisticsService) RecordTaskCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters["tasks_total"]++
}

// RecordTaskCompleted records a task completion event
func (s *StatisticsService) RecordTaskCompleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters["tasks_completed"]++
}

// RecordTaskFailed records a task failure event
func (s *StatisticsService) RecordTaskFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters["tasks_failed"]++
}

// RecordDuration records a task duration
func (s *StatisticsService) RecordDuration(durationMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Running average
	count := s.counters["duration_count"]
	total := s.counters["duration_total"]
	count++
	total += durationMs
	s.counters["duration_count"] = count
	s.counters["duration_total"] = total
	s.counters["avg_response_time"] = total / count
}

// GetUptime returns the service uptime
func (s *StatisticsService) GetUptime() string {
	return formatDuration(time.Since(s.startTime))
}

// GetStartTime returns the service start time
func (s *StatisticsService) GetStartTime() time.Time {
	return s.startTime
}

// Helper functions

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func getSystemInfo() SystemInfo {
	// In a real implementation, use runtime package
	return SystemInfo{
		GoVersion:    "go1.22",
		NumGoroutine: 10, // runtime.NumGoroutine()
		MemAllocMB:   64,
		MemTotalMB:   128,
		CPUUsage:     "low",
	}
}