package api

import (
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// Metrics holds application metrics
type Metrics struct {
	StartTime        time.Time
	RequestsTotal    atomic.Int64
	RequestsActive   atomic.Int64
	TasksCreated     atomic.Int64
	TasksCompleted   atomic.Int64
	TasksFailed      atomic.Int64
	PipelinesCreated atomic.Int64
	PipelinesRun     atomic.Int64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// IncrementRequests increments request counter
func (m *Metrics) IncrementRequests() {
	m.RequestsTotal.Add(1)
}

// IncrementActiveRequests increments active request counter
func (m *Metrics) IncrementActiveRequests() {
	m.RequestsActive.Add(1)
}

// DecrementActiveRequests decrements active request counter
func (m *Metrics) DecrementActiveRequests() {
	m.RequestsActive.Add(-1)
}

// IncrementTasksCreated increments tasks created counter
func (m *Metrics) IncrementTasksCreated() {
	m.TasksCreated.Add(1)
}

// IncrementTasksCompleted increments tasks completed counter
func (m *Metrics) IncrementTasksCompleted() {
	m.TasksCompleted.Add(1)
}

// IncrementTasksFailed increments tasks failed counter
func (m *Metrics) IncrementTasksFailed() {
	m.TasksFailed.Add(1)
}

// MetricsMiddleware records metrics for each request
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.IncrementRequests()
		metrics.IncrementActiveRequests()

		start := time.Now()

		c.Next()

		metrics.DecrementActiveRequests()

		// Record latency
		latency := time.Since(start)
		c.Set("latency", latency)
	}
}

// MetricsHandler returns metrics in Prometheus format
func MetricsHandler(metrics *Metrics, executor *service.CLIExecutor, orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		uptime := time.Since(metrics.StartTime)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Prometheus format
		metricsText := `# HELP claude_pipeline_uptime_seconds Service uptime in seconds
# TYPE claude_pipeline_uptime_seconds gauge
claude_pipeline_uptime_seconds %.0f

# HELP claude_pipeline_requests_total Total number of requests
# TYPE claude_pipeline_requests_total counter
claude_pipeline_requests_total %d

# HELP claude_pipeline_requests_active Current active requests
# TYPE claude_pipeline_requests_active gauge
claude_pipeline_requests_active %d

# HELP claude_pipeline_tasks_created_total Total tasks created
# TYPE claude_pipeline_tasks_created_total counter
claude_pipeline_tasks_created_total %d

# HELP claude_pipeline_tasks_completed_total Total tasks completed
# TYPE claude_pipeline_tasks_completed_total counter
claude_pipeline_tasks_completed_total %d

# HELP claude_pipeline_tasks_failed_total Total tasks failed
# TYPE claude_pipeline_tasks_failed_total counter
claude_pipeline_tasks_failed_total %d

# HELP claude_pipeline_pipelines_created_total Total pipelines created
# TYPE claude_pipeline_pipelines_created_total counter
claude_pipeline_pipelines_created_total %d

# HELP claude_pipeline_pipelines_run_total Total pipeline runs
# TYPE claude_pipeline_pipelines_run_total counter
claude_pipeline_pipelines_run_total %d

# HELP claude_pipeline_memory_alloc_bytes Current memory allocation
# TYPE claude_pipeline_memory_alloc_bytes gauge
claude_pipeline_memory_alloc_bytes %d

# HELP claude_pipeline_goroutines Current number of goroutines
# TYPE claude_pipeline_goroutines gauge
claude_pipeline_goroutines %d
`

		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(
			formatMetrics(metricsText,
				uptime.Seconds(),
				metrics.RequestsTotal.Load(),
				metrics.RequestsActive.Load(),
				metrics.TasksCreated.Load(),
				metrics.TasksCompleted.Load(),
				metrics.TasksFailed.Load(),
				metrics.PipelinesCreated.Load(),
				metrics.PipelinesRun.Load(),
				memStats.Alloc,
				runtime.NumGoroutine(),
			),
		))
	}
}

func formatMetrics(format string, args ...interface{}) string {
	return sprintf(format, args...)
}

func sprintf(format string, args ...interface{}) string {
	return gin.H{}.String() // placeholder
}

// HealthHandler returns detailed health status
func HealthHandler(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		uptime := time.Since(metrics.StartTime)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"uptime": uptime.String(),
			"metrics": gin.H{
				"requests_total":    metrics.RequestsTotal.Load(),
				"requests_active":   metrics.RequestsActive.Load(),
				"tasks_created":     metrics.TasksCreated.Load(),
				"tasks_completed":   metrics.TasksCompleted.Load(),
				"tasks_failed":      metrics.TasksFailed.Load(),
				"pipelines_created": metrics.PipelinesCreated.Load(),
				"pipelines_run":     metrics.PipelinesRun.Load(),
			},
			"runtime": gin.H{
				"goroutines":   runtime.NumGoroutine(),
				"memory_alloc": memStats.Alloc,
				"memory_sys":   memStats.Sys,
				"go_version":   runtime.Version(),
			},
		})
	}
}