package api

import (
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor) {
	api := r.Group("/api")
	{
		// Skills
		api.GET("/skills", ListSkills(skillSvc))
		api.GET("/skills/:id", GetSkill(skillSvc))
		api.POST("/skills/sync", SyncSkills(skillSvc))

		// Tasks
		api.POST("/tasks", CreateTask(taskSvc))
		api.GET("/tasks", ListTasks(taskSvc))
		api.GET("/tasks/:id", GetTask(taskSvc))
		api.GET("/tasks/:id/result", GetTaskResult(taskSvc))
		api.DELETE("/tasks/:id", CancelTask(executor))

		// WebSocket
		api.GET("/ws/tasks/:id/output", TaskOutputWS(taskSvc))

		// Status
		api.GET("/status", GetStatus(executor, skillSvc))
	}
}