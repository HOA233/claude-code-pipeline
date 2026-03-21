package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// HealthStatus represents detailed health status
type HealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	Version     string            `json:"version"`
	Components  ComponentHealth   `json:"components"`
	System      SystemHealth      `json:"system"`
}

// ComponentHealth represents health of individual components
type ComponentHealth struct {
	Redis   ComponentStatus `json:"redis"`
	API     ComponentStatus `json:"api"`
	Queue   ComponentStatus `json:"queue"`
	Executor ComponentStatus `json:"executor"`
}

// ComponentStatus represents status of a single component
type ComponentStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SystemHealth represents system-level health
type SystemHealth struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	MemoryMB     uint64 `json:"memory_mb"`
	CPUCount     int    `json:"cpu_count"`
}

// HealthHandler returns a detailed health check handler
func HealthHandler(redis *repository.RedisClient, startTime time.Time, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
			Uptime:    time.Since(startTime).String(),
			Version:   version,
			Components: ComponentHealth{
				API: ComponentStatus{Status: "up"},
			},
			System: SystemHealth{
				GoVersion:    runtime.Version(),
				NumGoroutine: runtime.NumGoroutine(),
				CPUCount:     runtime.NumCPU(),
			},
		}

		// Check Redis
		if redis != nil {
			start := time.Now()
			err := redis.Ping(c.Request.Context())
			latency := time.Since(start)

			if err != nil {
				health.Components.Redis = ComponentStatus{
					Status: "down",
					Error:  err.Error(),
				}
				health.Status = "degraded"
			} else {
				health.Components.Redis = ComponentStatus{
					Status:  "up",
					Latency: latency.String(),
				}
			}
		}

		// Check queue
		if redis != nil {
			length, err := redis.GetQueueLength(c.Request.Context())
			if err != nil {
				health.Components.Queue = ComponentStatus{
					Status: "unknown",
					Error:  err.Error(),
				}
			} else {
				health.Components.Queue = ComponentStatus{
					Status: "up",
				}
				_ = length
			}
		}

		// Get memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		health.System.MemoryMB = m.Alloc / 1024 / 1024

		// Determine HTTP status
		httpStatus := http.StatusOK
		if health.Status == "degraded" {
			httpStatus = http.StatusOK // Still return 200 but show degraded
		}

		c.JSON(httpStatus, health)
	}
}

// ReadinessHandler checks if the service is ready to accept traffic
func ReadinessHandler(redis *repository.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check all dependencies
		checks := make(map[string]string)

		// Check Redis
		if redis != nil {
			if err := redis.Ping(c.Request.Context()); err != nil {
				checks["redis"] = "not ready: " + err.Error()
			} else {
				checks["redis"] = "ready"
			}
		}

		// If any check failed, return 503
		allReady := true
		for _, status := range checks {
			if status != "ready" {
				allReady = false
				break
			}
		}

		if allReady {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
				"checks": checks,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"checks": checks,
			})
		}
	}
}

// LivenessHandler returns a simple liveness check
func LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	}
}

// DetailedStatusHandler returns comprehensive system status
func DetailedStatusHandler(
	redis *repository.RedisClient,
	executor *service.CLIExecutor,
	skillSvc *service.SkillService,
	orch *service.Orchestrator,
	statsSvc *service.StatisticsService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{
			"timestamp": time.Now(),
		}

		// CLI Executor status
		if executor != nil {
			status["executor"] = executor.GetStatus()
		}

		// Skills status
		if skillSvc != nil {
			skills, _ := skillSvc.GetAllSkills(c.Request.Context())
			status["skills"] = gin.H{
				"count":   len(skills),
				"enabled": countEnabledSkills(skills),
			}
		}

		// Pipeline status
		if orch != nil {
			pipelines, _ := redis.GetAllPipelines(c.Request.Context())
			runs, _ := redis.GetAllRuns(c.Request.Context())
			status["pipelines"] = gin.H{
				"count": len(pipelines),
			}
			status["runs"] = gin.H{
				"count": len(runs),
			}
		}

		// Statistics
		if statsSvc != nil {
			stats, _ := statsSvc.GetStats(c.Request.Context())
			status["statistics"] = stats
		}

		// Queue status
		queueLen, _ := redis.GetQueueLength(c.Request.Context())
		runQueueLen, _ := redis.GetRunQueueLength(c.Request.Context())
		status["queues"] = gin.H{
			"task_queue": queueLen,
			"run_queue":  runQueueLen,
		}

		c.JSON(http.StatusOK, status)
	}
}

func countEnabledSkills(skills []*service.Skill) int {
	count := 0
	for _, s := range skills {
		if s.Enabled {
			count++
		}
	}
	return count
}