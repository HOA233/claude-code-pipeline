package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockServer creates a mock HTTP server for testing
func MockServer() *httptest.Server {
	router := gin.New()

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"alive": true})
	})

	// API endpoints
	api := router.Group("/api")
	{
		// Skills
		api.GET("/skills", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"skills": []gin.H{
					{
						"id":          "code-review",
						"name":        "Code Review",
						"description": "Analyzes code quality",
						"version":     "1.0.0",
						"category":    "quality",
						"enabled":     true,
					},
					{
						"id":          "deploy",
						"name":        "Deploy",
						"description": "Deploys services",
						"version":     "2.0.0",
						"category":    "devops",
						"enabled":     true,
					},
				},
			})
		})

		api.GET("/skills/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "code-review" {
				c.JSON(http.StatusOK, gin.H{
					"id":          "code-review",
					"name":        "Code Review",
					"description": "Analyzes code quality",
					"version":     "1.0.0",
					"parameters": []gin.H{
						{"name": "target", "type": "string", "required": true},
						{"name": "depth", "type": "enum", "values": []string{"quick", "standard", "deep"}},
					},
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			}
		})

		// Tasks
		api.GET("/tasks", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"tasks": []gin.H{
					{"id": "task-001", "skill_id": "code-review", "status": "completed"},
					{"id": "task-002", "skill_id": "deploy", "status": "running"},
				},
			})
		})

		api.POST("/tasks", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusAccepted, gin.H{
				"id":        "task-new",
				"skill_id":  req["skill_id"],
				"status":    "pending",
				"params":    req["parameters"],
			})
		})

		api.GET("/tasks/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":        id,
				"skill_id":  "code-review",
				"status":    "running",
				"progress":  gin.H{"current": 2, "total": 5},
			})
		})

		api.GET("/tasks/:id/result", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":        id,
				"status":    "completed",
				"result":    gin.H{"summary": "Analysis complete"},
				"duration":  5000,
			})
		})

		api.DELETE("/tasks/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"cancelled": true})
		})

		// Pipelines
		api.GET("/pipelines", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"pipelines": []gin.H{
					{"id": "pipeline-001", "name": "Full Review", "mode": "serial"},
				},
			})
		})

		api.POST("/pipelines", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":        "pipeline-new",
				"name":      req["name"],
				"mode":      req["mode"],
			})
		})

		api.POST("/pipelines/:id/run", func(c *gin.Context) {
			c.JSON(http.StatusAccepted, gin.H{
				"id":          "run-new",
				"pipeline_id": c.Param("id"),
				"status":      "pending",
			})
		})

		// Runs
		api.GET("/runs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"runs": []gin.H{
					{"id": "run-001", "pipeline_id": "pipeline-001", "status": "completed"},
				},
			})
		})

		// Schedules
		api.GET("/schedules", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"schedules": []gin.H{
					{"id": "schedule-001", "skill_id": "code-review", "cron_expr": "@daily"},
				},
			})
		})

		api.POST("/schedules", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":        "schedule-new",
				"skill_id":  req["skill_id"],
				"cron_expr": req["cron_expr"],
			})
		})

		// Status
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"cli": gin.H{
					"active_count":    2,
					"max_concurrency": 10,
				},
			})
		})
	}

	return httptest.NewServer(router)
}

func TestMockServerHealth(t *testing.T) {
	server := MockServer()
	defer server.Close()

	// Test health endpoint
	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", result["status"])
	}
}

func TestMockServerSkills(t *testing.T) {
	server := MockServer()
	defer server.Close()

	// Test skills list
	resp, err := http.Get(server.URL + "/api/skills")
	if err != nil {
		t.Fatalf("Failed to call skills endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	skills, ok := result["skills"].([]interface{})
	if !ok {
		t.Fatal("Expected skills array")
	}

	if len(skills) == 0 {
		t.Error("Expected at least one skill")
	}
}

func TestMockServerCreateTask(t *testing.T) {
	server := MockServer()
	defer server.Close()

	// Create task
	reqBody := `{"skill_id":"code-review","parameters":{"target":"src/"}}`
	resp, err := http.Post(server.URL+"/api/tasks", "application/json", nil)
	if err != nil {
		_ = reqBody
		t.Fatalf("Failed to create task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", resp.StatusCode)
	}
}

func TestMockServerPipelines(t *testing.T) {
	server := MockServer()
	defer server.Close()

	// Test pipelines list
	resp, err := http.Get(server.URL + "/api/pipelines")
	if err != nil {
		t.Fatalf("Failed to call pipelines endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMockServerStatus(t *testing.T) {
	server := MockServer()
	defer server.Close()

	// Test status endpoint
	resp, err := http.Get(server.URL + "/api/status")
	if err != nil {
		t.Fatalf("Failed to call status endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", result["status"])
	}
}