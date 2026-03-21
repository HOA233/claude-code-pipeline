package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// ListPipelines returns all pipelines
func ListPipelines(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		pipelines, err := orch.ListPipelines(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"pipelines": pipelines})
	}
}

// GetPipeline returns a single pipeline
func GetPipeline(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pipeline, err := orch.GetPipeline(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
			return
		}
		c.JSON(http.StatusOK, pipeline)
	}
}

// CreatePipeline creates a new pipeline
func CreatePipeline(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.PipelineCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pipeline, err := orch.CreatePipeline(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, pipeline)
	}
}

// DeletePipeline deletes a pipeline
func DeletePipeline(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := orch.DeletePipeline(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}

// ListRuns returns all runs
func ListRuns(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		runs, err := orch.ListRuns(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"runs": runs})
	}
}

// GetRun returns a single run
func GetRun(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		run, err := orch.GetRun(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
			return
		}
		c.JSON(http.StatusOK, run)
	}
}

// RunPipeline executes a pipeline
func RunPipeline(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req model.RunCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			req = model.RunCreateRequest{PipelineID: id}
		} else {
			req.PipelineID = id
		}

		run, err := orch.RunPipeline(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, run)
	}
}

// CancelRun cancels a running pipeline
func CancelRun(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := orch.CancelRun(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"cancelled": true})
	}
}

// RunOutputWS handles WebSocket for run output
func RunOutputWS(orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Get updates from Redis pub/sub
		pubsub := orch.SubscribeRunUpdates(c.Request.Context(), id)
		defer pubsub.Close()

		ch := pubsub.Channel()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if err := conn.WriteMessage(1, []byte(msg.Payload)); err != nil {
					return
				}
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}