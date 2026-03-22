package api

import (
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient) {
	// Initialize WebSocket handlers
	wsHandler := NewWebSocketHandler(redis)
	taskSSEHandler := NewTaskSSEHandler(redis)

	api := r.Group("/api")
	{
		// Skills
		api.GET("/skills", ListSkills(skillSvc))
		api.GET("/skills/:id", GetSkill(skillSvc))
		api.POST("/skills/sync", SyncSkills(skillSvc))

		// Tasks (single CLI execution)
		api.POST("/tasks", CreateTask(taskSvc))
		api.GET("/tasks", ListTasks(taskSvc))
		api.GET("/tasks/:id", GetTask(taskSvc))
		api.GET("/tasks/:id/result", GetTaskResult(taskSvc))
		api.DELETE("/tasks/:id", CancelTask(executor))

		// Pipelines (multi-CLI orchestration)
		api.GET("/pipelines", ListPipelines(orch))
		api.POST("/pipelines", CreatePipeline(orch))
		api.GET("/pipelines/:id", GetPipeline(orch))
		api.DELETE("/pipelines/:id", DeletePipeline(orch))
		api.POST("/pipelines/:id/run", RunPipeline(orch))

		// Runs (pipeline executions)
		api.GET("/runs", ListRuns(orch))
		api.GET("/runs/:id", GetRun(orch))
		api.DELETE("/runs/:id", CancelRun(orch))

		// Status
		api.GET("/status", GetStatus(executor, skillSvc, orch))
	}

	// WebSocket endpoints (outside /api for easier client connection)
	r.GET("/ws/tasks/:id", wsHandler.HandleTaskWS)
	r.GET("/ws/runs/:id", wsHandler.HandleRunWS)
	r.GET("/ws", wsHandler.HandleGlobalWS)

	// SSE endpoints (alternative to WebSocket)
	r.GET("/sse/tasks/:id", taskSSEHandler.HandleTaskSSE)
	r.GET("/sse/runs/:id", taskSSEHandler.HandleRunSSE)
	r.GET("/sse", taskSSEHandler.HandleGlobalSSE)
}

// SetupRoutesWithScheduler sets up routes including scheduler endpoints
func SetupRoutesWithScheduler(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient, schedulerSvc *service.SchedulerService) {
	// Set up base routes
	SetupRoutes(r, skillSvc, taskSvc, executor, orch, redis)

	// Add scheduler routes
	api := r.Group("/api")
	{
		// Schedules
		api.GET("/schedules", ListSchedules(schedulerSvc))
		api.POST("/schedules", CreateSchedule(schedulerSvc))
		api.GET("/schedules/:id", GetSchedule(schedulerSvc))
		api.PUT("/schedules/:id", UpdateSchedule(schedulerSvc))
		api.DELETE("/schedules/:id", DeleteSchedule(schedulerSvc))
		api.POST("/schedules/:id/enable", EnableSchedule(schedulerSvc))
		api.POST("/schedules/:id/disable", DisableSchedule(schedulerSvc))
		api.POST("/schedules/:id/trigger", TriggerSchedule(schedulerSvc))
	}
}