package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// Scheduler handlers

// ListSchedules returns all scheduled tasks
func ListSchedules(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		schedules, err := svc.ListSchedules(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"schedules": schedules})
	}
}

// GetSchedule returns a specific schedule
func GetSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		schedule, err := svc.GetSchedule(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}
		c.JSON(http.StatusOK, schedule)
	}
}

// CreateSchedule creates a new scheduled task
func CreateSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.Schedule
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.SkillID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "skill_id is required"})
			return
		}

		if req.CronExpr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cron_expr is required"})
			return
		}

		if err := svc.CreateSchedule(c.Request.Context(), &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, req)
	}
}

// UpdateSchedule updates an existing schedule
func UpdateSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.UpdateSchedule(c.Request.Context(), id, updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"updated": true})
	}
}

// DeleteSchedule deletes a schedule
func DeleteSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := svc.DeleteSchedule(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}

// EnableSchedule enables a schedule
func EnableSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := svc.EnableSchedule(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enabled": true})
	}
}

// DisableSchedule disables a schedule
func DisableSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := svc.DisableSchedule(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enabled": false})
	}
}

// TriggerSchedule manually triggers a scheduled task
func TriggerSchedule(svc *service.SchedulerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		task, err := svc.TriggerSchedule(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"triggered": true,
			"task":      task,
		})
	}
}