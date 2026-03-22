package service

import (
	"context"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
)

// MetricsService 指标服务
type MetricsService struct {
	redis *repository.RedisClient
}

// NewMetricsService 创建指标服务
func NewMetricsService(redis *repository.RedisClient) *MetricsService {
	return &MetricsService{redis: redis}
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	// Agent 指标
	TotalAgents    int `json:"total_agents"`
	EnabledAgents  int `json:"enabled_agents"`
	DisabledAgents int `json:"disabled_agents"`

	// Workflow 指标
	TotalWorkflows    int `json:"total_workflows"`
	EnabledWorkflows  int `json:"enabled_workflows"`
	DisabledWorkflows int `json:"disabled_workflows"`

	// Execution 指标
	TotalExecutions     int `json:"total_executions"`
	RunningExecutions   int `json:"running_executions"`
	PendingExecutions   int `json:"pending_executions"`
	CompletedExecutions int `json:"completed_executions"`
	FailedExecutions    int `json:"failed_executions"`

	// Scheduled Job 指标
	TotalJobs    int `json:"total_jobs"`
	EnabledJobs  int `json:"enabled_jobs"`
	DisabledJobs int `json:"disabled_jobs"`

	// 性能指标
	AverageExecutionDuration int64   `json:"avg_execution_duration_ms"`
	SuccessRate             float64 `json:"success_rate"`

	// 系统资源指标 (模拟值，实际应用中应从系统获取)
	CpuUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	Goroutines     int     `json:"goroutines"`
	ActiveExecutions int   `json:"active_executions"`
	QueuedTasks    int     `json:"queued_tasks"`
	RedisMemory    int64   `json:"redis_memory"`
	RedisKeys      int     `json:"redis_keys"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// GetSystemMetrics 获取系统指标
func (s *MetricsService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Agent 指标
	agents, _ := s.redis.ListAgents(ctx, "", "")
	metrics.TotalAgents = len(agents)
	for _, a := range agents {
		if a.Enabled {
			metrics.EnabledAgents++
		} else {
			metrics.DisabledAgents++
		}
	}

	// Workflow 指标
	workflows, _ := s.redis.ListWorkflows(ctx, "")
	metrics.TotalWorkflows = len(workflows)
	for _, w := range workflows {
		if w.Enabled {
			metrics.EnabledWorkflows++
		} else {
			metrics.DisabledWorkflows++
		}
	}

	// Execution 指标
	executions, _ := s.redis.ListExecutions(ctx, &model.ExecutionFilter{Page: 1, PageSize: 1000})
	metrics.TotalExecutions = executions.Total

	var totalDuration int64
	var completedCount int

	for _, exec := range executions.Executions {
		switch exec.Status {
		case model.ExecutionStatusRunning:
			metrics.RunningExecutions++
		case model.ExecutionStatusPending:
			metrics.PendingExecutions++
		case model.ExecutionStatusCompleted:
			metrics.CompletedExecutions++
			completedCount++
		case model.ExecutionStatusFailed:
			metrics.FailedExecutions++
		}
		totalDuration += exec.Duration
	}

	// 计算平均执行时间
	if metrics.TotalExecutions > 0 {
		metrics.AverageExecutionDuration = totalDuration / int64(metrics.TotalExecutions)
	}

	// 计算成功率
	totalFinished := metrics.CompletedExecutions + metrics.FailedExecutions
	if totalFinished > 0 {
		metrics.SuccessRate = float64(metrics.CompletedExecutions) / float64(totalFinished) * 100
	}

	// Scheduled Job 指标
	jobs, _ := s.redis.ListScheduledJobs(ctx, "")
	metrics.TotalJobs = len(jobs)
	for _, j := range jobs {
		if j.Enabled {
			metrics.EnabledJobs++
		} else {
			metrics.DisabledJobs++
		}
	}

	// 模拟系统资源指标 (实际应用中应使用 runtime 包获取)
	metrics.ActiveExecutions = metrics.RunningExecutions
	metrics.QueuedTasks = metrics.PendingExecutions
	metrics.Goroutines = 50 + metrics.RunningExecutions*2 // 模拟值
	metrics.CpuUsage = 15.0 + float64(metrics.RunningExecutions)*5.0
	metrics.MemoryUsage = 30.0 + float64(metrics.RunningExecutions)*3.0
	metrics.RedisMemory = int64(len(agents)*1024 + len(workflows)*2048 + executions.Total*512)
	metrics.RedisKeys = len(agents) + len(workflows) + executions.Total + len(jobs)

	return metrics, nil
}

// ExecutionTrend 执行趋势
type ExecutionTrend struct {
	Date      string `json:"date"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Failed    int    `json:"failed"`
}

// GetExecutionTrend 获取执行趋势
func (s *MetricsService) GetExecutionTrend(ctx context.Context, days int) ([]ExecutionTrend, error) {
	trends := make([]ExecutionTrend, 0)

	executions, _ := s.redis.ListExecutions(ctx, &model.ExecutionFilter{Page: 1, PageSize: 10000})

	// 按日期分组
	dateMap := make(map[string]*ExecutionTrend)

	for _, exec := range executions.Executions {
		date := exec.CreatedAt.Format("2006-01-02")
		if _, ok := dateMap[date]; !ok {
			dateMap[date] = &ExecutionTrend{Date: date}
		}
		dateMap[date].Total++
		if exec.Status == model.ExecutionStatusCompleted {
			dateMap[date].Completed++
		} else if exec.Status == model.ExecutionStatusFailed {
			dateMap[date].Failed++
		}
	}

	// 转换为列表
	for _, trend := range dateMap {
		trends = append(trends, *trend)
	}

	return trends, nil
}

// WorkflowStats 工作流统计
type WorkflowStats struct {
	WorkflowID    string  `json:"workflow_id"`
	WorkflowName  string  `json:"workflow_name"`
	TotalRuns     int     `json:"total_runs"`
	SuccessRate   float64 `json:"success_rate"`
	AvgDuration   int64   `json:"avg_duration_ms"`
	LastRunStatus string  `json:"last_run_status"`
}

// GetWorkflowStats 获取工作流统计
func (s *MetricsService) GetWorkflowStats(ctx context.Context) ([]WorkflowStats, error) {
	stats := make([]WorkflowStats, 0)

	workflows, _ := s.redis.ListWorkflows(ctx, "")
	executions, _ := s.redis.ListExecutions(ctx, &model.ExecutionFilter{Page: 1, PageSize: 10000})

	// 按工作流分组
	workflowMap := make(map[string]*WorkflowStats)

	for _, w := range workflows {
		workflowMap[w.ID] = &WorkflowStats{
			WorkflowID:   w.ID,
			WorkflowName: w.Name,
		}
	}

	for _, exec := range executions.Executions {
		if stat, ok := workflowMap[exec.WorkflowID]; ok {
			stat.TotalRuns++
			stat.AvgDuration += exec.Duration
			if exec.Status == model.ExecutionStatusCompleted {
				stat.SuccessRate++
			}
			stat.LastRunStatus = string(exec.Status)
		}
	}

	// 计算平均值
	for _, stat := range workflowMap {
		if stat.TotalRuns > 0 {
			stat.SuccessRate = stat.SuccessRate / float64(stat.TotalRuns) * 100
			stat.AvgDuration = stat.AvgDuration / int64(stat.TotalRuns)
		}
		stats = append(stats, *stat)
	}

	return stats, nil
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status     string                      `json:"status"`
	Components map[string]ComponentHealth  `json:"components"`
	Version    string                      `json:"version"`
	Uptime     int64                       `json:"uptime_seconds"`
	CheckedAt  time.Time                   `json:"checked_at"`
	LastCheck  string                      `json:"last_check"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status       string `json:"status"`
	Latency      int64  `json:"latency,omitempty"`
	RunningTasks int    `json:"running_tasks,omitempty"`
	ActiveJobs   int    `json:"active_jobs,omitempty"`
}

// GetHealthStatus 获取健康状态
func (s *MetricsService) GetHealthStatus(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Status:     "healthy",
		Components: make(map[string]ComponentHealth),
		Version:    "1.0.0",
		CheckedAt:  time.Now(),
	}

	// 检查 Redis 连接
	start := time.Now()
	if err := s.redis.Ping(ctx); err != nil {
		status.Components["redis"] = ComponentHealth{Status: "down"}
		status.Status = "unhealthy"
	} else {
		latency := time.Since(start).Milliseconds()
		status.Components["redis"] = ComponentHealth{
			Status:  "up",
			Latency: latency,
		}
	}

	// 获取运行中的执行数
	executions, _ := s.redis.ListExecutions(ctx, &model.ExecutionFilter{Page: 1, PageSize: 100})
	runningCount := 0
	for _, exec := range executions.Executions {
		if exec.Status == model.ExecutionStatusRunning {
			runningCount++
		}
	}
	status.Components["executor"] = ComponentHealth{
		Status:       "up",
		RunningTasks: runningCount,
	}

	// 获取活跃的定时任务数
	jobs, _ := s.redis.ListScheduledJobs(ctx, "")
	activeJobs := 0
	for _, job := range jobs {
		if job.Enabled {
			activeJobs++
		}
	}
	status.Components["scheduler"] = ComponentHealth{
		Status:     "up",
		ActiveJobs: activeJobs,
	}

	status.LastCheck = status.CheckedAt.Format(time.RFC3339)

	return status
}