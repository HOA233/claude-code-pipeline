package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ListSkills returns all skills
func ListSkills(svc *service.SkillService) gin.HandlerFunc {
	return func(c *gin.Context) {
		skills, err := svc.GetAllSkills(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"skills": skills})
	}
}

// GetSkill returns a single skill
func GetSkill(svc *service.SkillService) gin.HandlerFunc {
	return func(c *gin.Context) {
		skillID := c.Param("id")
		skill, err := svc.GetSkill(c.Request.Context(), skillID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Skill not found"})
			return
		}
		c.JSON(http.StatusOK, skill)
	}
}

// SyncSkills syncs skills from GitLab
func SyncSkills(svc *service.SkillService) gin.HandlerFunc {
	return func(c *gin.Context) {
		skills, err := svc.SyncFromGitLab(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Skills synced successfully",
			"count":   len(skills),
		})
	}
}

// CreateTask creates a new task
func CreateTask(svc *service.TaskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.TaskCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task, err := svc.CreateTask(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, task)
	}
}

// ListTasks returns all tasks
func ListTasks(svc *service.TaskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tasks, err := svc.GetAllTasks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}

// GetTask returns a single task
func GetTask(svc *service.TaskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")
		task, err := svc.GetTask(c.Request.Context(), taskID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	}
}

// GetTaskResult returns task result
func GetTaskResult(svc *service.TaskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")
		result, err := svc.GetTaskResult(c.Request.Context(), taskID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// CancelTask cancels a running task
func CancelTask(executor *service.CLIExecutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")
		if err := executor.CancelTask(c.Request.Context(), taskID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"cancelled": true, "task_id": taskID})
	}
}

// TaskOutputWS handles WebSocket connections for task output
func TaskOutputWS(svc *service.TaskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		pubsub := svc.SubscribeTaskUpdates(c.Request.Context(), taskID)
		defer pubsub.Close()

		ch := pubsub.Channel()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					return
				}
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}