package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/api"
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// Integration tests require running Redis

func setupIntegrationServer(t *testing.T) (*gin.Engine, *repository.RedisClient, func()) {
	gin.SetMode(gin.TestMode)

	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   2, // Use different DB for integration tests
	}

	redisClient := repository.NewRedisClient(cfg)

	ctx := t.Context()
	if err := redisClient.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	skillSvc := service.NewSkillService(redisClient, config.GitLabConfig{})
	taskSvc := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, config.CLIConfig{})
	orchestrator := service.NewOrchestrator(redisClient, executor)

	// Sync default skills
	skillSvc.SyncFromGitLab(ctx)

	router := gin.New()
	api.SetupRoutes(router, skillSvc, taskSvc, executor, orchestrator)

	cleanup := func() {
		// Clean up test data
		redisClient.Close()
	}

	return router, redisClient, cleanup
}

func TestIntegrationFullTaskWorkflow(t *testing.T) {
	router, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Step 1: List available skills
	t.Log("Step 1: Listing skills")
	req, _ := http.NewRequest("GET", "/api/skills", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to list skills: %d", w.Code)
	}

	var skillsResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &skillsResp)
	skills := skillsResp["skills"].([]interface{})
	if len(skills) == 0 {
		t.Fatal("No skills available")
	}

	// Step 2: Create a task
	t.Log("Step 2: Creating task")
	taskReq := map[string]interface{}{
		"skill_id": "code-review",
		"parameters": map[string]interface{}{
			"target": "src/",
			"depth":  "quick",
		},
	}

	body, _ := json.Marshal(taskReq)
	req, _ = http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Failed to create task: %d - %s", w.Code, w.Body.String())
	}

	var taskResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &taskResp)
	taskID := taskResp["id"].(string)
	t.Logf("Created task: %s", taskID)

	// Step 3: Get task status
	t.Log("Step 3: Getting task status")
	req, _ = http.NewRequest("GET", "/api/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get task: %d", w.Code)
	}

	// Step 4: List all tasks
	t.Log("Step 4: Listing all tasks")
	req, _ = http.NewRequest("GET", "/api/tasks", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to list tasks: %d", w.Code)
	}
}

func TestIntegrationFullPipelineWorkflow(t *testing.T) {
	router, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Step 1: Create a pipeline
	t.Log("Step 1: Creating pipeline")
	pipelineReq := map[string]interface{}{
		"name":        "integration-test-pipeline",
		"description": "Pipeline for integration testing",
		"mode":        "serial",
		"steps": []map[string]interface{}{
			{
				"id":      "step-1",
				"name":    "First Step",
				"cli":     "echo",
				"action":  "command",
				"command": "test",
				"params":  map[string]interface{}{},
			},
			{
				"id":        "step-2",
				"name":      "Second Step",
				"cli":       "echo",
				"action":    "command",
				"command":   "done",
				"depends_on": []string{"step-1"},
				"params":    map[string]interface{}{},
			},
		},
	}

	body, _ := json.Marshal(pipelineReq)
	req, _ := http.NewRequest("POST", "/api/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("Failed to create pipeline: %d - %s", w.Code, w.Body.String())
	}

	var pipelineResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &pipelineResp)
	pipelineID := pipelineResp["id"].(string)
	t.Logf("Created pipeline: %s", pipelineID)

	// Step 2: Get pipeline details
	t.Log("Step 2: Getting pipeline details")
	req, _ = http.NewRequest("GET", "/api/pipelines/"+pipelineID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get pipeline: %d", w.Code)
	}

	// Step 3: Run the pipeline
	t.Log("Step 3: Running pipeline")
	runReq := map[string]interface{}{
		"pipeline_id": pipelineID,
		"params":      map[string]interface{}{},
	}

	body, _ = json.Marshal(runReq)
	req, _ = http.NewRequest("POST", "/api/pipelines/"+pipelineID+"/run", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Run response: %s", w.Body.String())

	// Step 4: List pipelines
	t.Log("Step 4: Listing pipelines")
	req, _ = http.NewRequest("GET", "/api/pipelines", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to list pipelines: %d", w.Code)
	}

	// Step 5: Delete pipeline
	t.Log("Step 5: Deleting pipeline")
	req, _ = http.NewRequest("DELETE", "/api/pipelines/"+pipelineID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to delete pipeline: %d", w.Code)
	}
}

func TestIntegrationConcurrentRequests(t *testing.T) {
	router, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	const numRequests = 10
	done := make(chan bool, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			// Create task
			taskReq := map[string]interface{}{
				"skill_id": "code-review",
				"parameters": map[string]interface{}{
					"target": "src/",
					"depth":  "quick",
				},
			}

			body, _ := json.Marshal(taskReq)
			req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusAccepted {
				t.Errorf("Request %d failed: %d", idx, w.Code)
			}
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func TestIntegrationErrorHandling(t *testing.T) {
	router, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Test invalid skill ID
	t.Log("Testing invalid skill ID")
	taskReq := map[string]interface{}{
		"skill_id": "non-existent-skill",
		"parameters": map[string]interface{}{
			"target": "src/",
		},
	}

	body, _ := json.Marshal(taskReq)
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusAccepted {
		t.Error("Expected error for invalid skill ID")
	}

	// Test missing required parameters
	t.Log("Testing missing required parameters")
	taskReq = map[string]interface{}{
		"skill_id": "code-review",
		"parameters": map[string]interface{}{
			// Missing required "target" parameter
		},
	}

	body, _ = json.Marshal(taskReq)
	req, _ = http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusAccepted {
		t.Error("Expected error for missing required parameter")
	}

	// Test invalid pipeline
	t.Log("Testing invalid pipeline creation")
	pipelineReq := map[string]interface{}{
		"name": "", // Empty name should fail
		"steps": []map[string]interface{}{},
	}

	body, _ = json.Marshal(pipelineReq)
	req, _ = http.NewRequest("POST", "/api/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusCreated {
		t.Error("Expected error for invalid pipeline")
	}
}

func TestIntegrationStatusEndpoints(t *testing.T) {
	router, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint failed: %d", w.Code)
	}

	// Test status endpoint
	req, _ = http.NewRequest("GET", "/api/status", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status endpoint failed: %d", w.Code)
	}

	var statusResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &statusResp)

	if statusResp["status"] != "healthy" {
		t.Error("Expected healthy status")
	}
}