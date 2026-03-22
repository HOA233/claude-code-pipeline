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

// SetupRoutesWithAgent sets up routes including Agent, Workflow, and Execution endpoints
func SetupRoutesWithAgent(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient, agentSvc *service.AgentService, workflowSvc *service.WorkflowService) {
	// Set up base routes
	SetupRoutes(r, skillSvc, taskSvc, executor, orch, redis)

	// Initialize handlers
	agentHandler := NewAgentHandler(agentSvc)
	workflowHandler := NewWorkflowHandler(workflowSvc)
	executionHandler := NewExecutionHandler(workflowSvc)

	api := r.Group("/api")
	{
		// Agents
		api.POST("/agents", agentHandler.CreateAgent)
		api.GET("/agents", agentHandler.ListAgents)
		api.GET("/agents/:id", agentHandler.GetAgent)
		api.PUT("/agents/:id", agentHandler.UpdateAgent)
		api.DELETE("/agents/:id", agentHandler.DeleteAgent)
		api.POST("/agents/:id/test", agentHandler.TestAgent)
		api.POST("/agents/:id/execute", agentHandler.ExecuteAgent)

		// Workflows
		api.POST("/workflows", workflowHandler.CreateWorkflow)
		api.GET("/workflows", workflowHandler.ListWorkflows)
		api.GET("/workflows/:id", workflowHandler.GetWorkflow)
		api.PUT("/workflows/:id", workflowHandler.UpdateWorkflow)
		api.DELETE("/workflows/:id", workflowHandler.DeleteWorkflow)

		// Executions
		api.POST("/executions", executionHandler.ExecuteWorkflow)
		api.GET("/executions", executionHandler.ListExecutions)
		api.GET("/executions/:id", executionHandler.GetExecution)
		api.POST("/executions/:id/cancel", executionHandler.CancelExecution)
		api.POST("/executions/:id/pause", executionHandler.PauseExecution)
		api.POST("/executions/:id/resume", executionHandler.ResumeExecution)
		api.POST("/executions/cancel-all", executionHandler.CancelAllExecutions)
	}
}

// SetupRoutesWithAll sets up all routes including scheduled jobs, stats, and execution details
func SetupRoutesWithAll(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient, agentSvc *service.AgentService, workflowSvc *service.WorkflowService, jobSvc *service.ScheduledJobService, metricsSvc *service.MetricsService, logSvc *service.ExecutionLogService, webhookSvc *service.WebhookService) {
	// Set up agent routes
	SetupRoutesWithAgent(r, skillSvc, taskSvc, executor, orch, redis, agentSvc, workflowSvc)

	// Initialize handlers
	jobHandler := NewScheduledJobHandler(jobSvc)
	statsHandler := NewStatsHandler(metricsSvc)
	execDetailHandler := NewExecutionDetailHandler(workflowSvc, logSvc)
	webhookHandler := NewWebhookHandler(webhookSvc)
	configHandler := NewConfigHandler()

	api := r.Group("/api")
	{
		// Scheduled Jobs
		api.POST("/schedules", jobHandler.CreateJob)
		api.GET("/schedules", jobHandler.ListJobs)
		api.GET("/schedules/:id", jobHandler.GetJob)
		api.PUT("/schedules/:id", jobHandler.UpdateJob)
		api.DELETE("/schedules/:id", jobHandler.DeleteJob)
		api.POST("/schedules/:id/enable", jobHandler.EnableJob)
		api.POST("/schedules/:id/disable", jobHandler.DisableJob)
		api.POST("/schedules/:id/trigger", jobHandler.TriggerJob)
		api.GET("/schedules/:id/history", jobHandler.GetJobHistory)

		// Stats & Metrics
		api.GET("/stats/system", statsHandler.GetSystemMetrics)
		api.GET("/stats/trends", statsHandler.GetExecutionTrend)
		api.GET("/stats/workflows", statsHandler.GetWorkflowStats)
		api.GET("/stats/health", statsHandler.GetHealthStatus)

		// Execution Details
		api.GET("/executions/:id/details", execDetailHandler.GetExecutionDetails)
		api.GET("/executions/:id/logs", execDetailHandler.GetExecutionLogs)
		api.GET("/executions/:id/stream", execDetailHandler.StreamExecutionLogs)
		api.POST("/executions/:id/retry", execDetailHandler.RetryExecution)
		api.GET("/metrics", execDetailHandler.GetExecutionMetrics)

		// Webhooks
		api.POST("/webhooks", webhookHandler.CreateWebhook)
		api.GET("/webhooks", webhookHandler.ListWebhooks)
		api.GET("/webhooks/:id", webhookHandler.GetWebhook)
		api.PUT("/webhooks/:id", webhookHandler.UpdateWebhook)
		api.DELETE("/webhooks/:id", webhookHandler.DeleteWebhook)
		api.GET("/webhooks/:id/deliveries", webhookHandler.GetWebhookDeliveries)

		// System Configuration
		api.GET("/config", configHandler.GetConfig)
		api.PUT("/config", configHandler.UpdateConfig)
		api.GET("/config/features", configHandler.GetFeatures)
		api.POST("/config/features/:feature/toggle", configHandler.ToggleFeature)
		api.GET("/models", configHandler.GetModels)
		api.GET("/categories", configHandler.GetCategories)
	}

	// SSE endpoints for real-time updates
	sseHandler := NewTaskSSEHandler(redis)
	r.GET("/sse/executions", sseHandler.HandleGlobalSSE)
	r.GET("/sse/executions/:id", sseHandler.HandleTaskSSE)
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