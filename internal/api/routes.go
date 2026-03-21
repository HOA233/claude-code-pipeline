package api

import (
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator) {
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

		// WebSocket
		api.GET("/ws/tasks/:id/output", TaskOutputWS(taskSvc))
		api.GET("/ws/runs/:id/output", RunOutputWS(orch))

		// Status
		api.GET("/status", GetStatus(executor, skillSvc, orch))
	}
}