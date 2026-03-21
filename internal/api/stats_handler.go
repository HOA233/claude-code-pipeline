package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// Statistics handlers

// GetSystemStats returns comprehensive system statistics
func GetSystemStats(statsSvc *service.StatisticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := statsSvc.GetStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

// Batch handlers

// CreateBatch creates a new batch operation
func CreateBatch(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string             `json:"name" binding:"required"`
			Description string             `json:"description"`
			Operations  []service.BatchTask `json:"operations" binding:"required"`
			Options     service.BatchOptions `json:"options"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		batch, err := batchSvc.CreateBatch(c.Request.Context(), req.Name, req.Operations, req.Options)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, batch)
	}
}

// ExecuteBatch executes a batch operation
func ExecuteBatch(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		batchID := c.Param("id")

		batch, err := batchSvc.ExecuteBatch(c.Request.Context(), batchID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, batch)
	}
}

// GetBatch returns a batch operation
func GetBatch(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		batchID := c.Param("id")

		batch, err := batchSvc.GetBatch(c.Request.Context(), batchID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "batch not found"})
			return
		}

		c.JSON(http.StatusOK, batch)
	}
}

// ListBatches lists all batch operations
func ListBatches(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		batches, err := batchSvc.ListBatches(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"batches": batches})
	}
}

// CancelBatch cancels a running batch
func CancelBatch(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		batchID := c.Param("id")

		if err := batchSvc.CancelBatch(c.Request.Context(), batchID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"cancelled": true})
	}
}

// DeleteBatch deletes a batch
func DeleteBatch(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		batchID := c.Param("id")

		if err := batchSvc.DeleteBatch(c.Request.Context(), batchID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}

// GetBatchStats returns batch operation statistics
func GetBatchStats(batchSvc *service.BatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := batchSvc.GetBatchStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// Priority Queue handlers

// GetQueueStats returns priority queue statistics
func GetQueueStats(pq *service.PriorityQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := pq.GetQueueStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// PromoteTask promotes a task in the queue
func PromoteTask(pq *service.PriorityQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		var req struct {
			Priority service.Priority `json:"priority"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := pq.PromoteTask(c.Request.Context(), taskID, req.Priority); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"promoted": true})
	}
}

// MoveTask moves a task to another queue
func MoveTask(pq *service.PriorityQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		var req struct {
			TargetQueue string `json:"target_queue"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := pq.MoveTask(c.Request.Context(), taskID, req.TargetQueue); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"moved": true})
	}
}

// ClearQueue clears a queue
func ClearQueue(pq *service.PriorityQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Query("queue")
		if queueName == "" {
			queueName = "default"
		}

		if err := pq.ClearQueue(c.Request.Context(), queueName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"cleared": true, "queue": queueName})
	}
}