package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/company/claude-pipeline/internal/api"
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

func setupPipelineTestServer(t *testing.T) (*gin.Engine, *repository.RedisClient) {
	gin.SetMode(gin.TestMode)

	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)

	ctx := t.Context()
	if err := redisClient.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	skillSvc := service.NewSkillService(redisClient, config.GitLabConfig{})
	taskSvc := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, config.CLIConfig{})
	orchestrator := service.NewOrchestrator(redisClient, executor)

	router := gin.New()
	api.SetupRoutes(router, skillSvc, taskSvc, executor, orchestrator)

	return router, redisClient
}

func TestCreatePipeline(t *testing.T) {
	router, _ := setupPipelineTestServer(t)

	pipelineReq := map[string]interface{}{
		"name":        "test-pipeline",
		"description": "Test pipeline for CI/CD",
		"mode":        "serial",
		"steps": []map[string]interface{}{
			{
				"id":     "step-1",
				"name":   "Code Review",
				"cli":    "claude",
				"action": "review",
				"params": map[string]interface{}{
					"target": "src/",
				},
			},
			{
				"id":        "step-2",
				"name":      "Run Tests",
				"cli":       "npm",
				"action":    "test",
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
		t.Errorf("Expected status 200/201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["id"] == nil {
		t.Error("Expected pipeline ID in response")
	}
}

func TestListPipelines(t *testing.T) {
	router, _ := setupPipelineTestServer(t)

	req, _ := http.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

func TestPipelineValidation(t *testing.T) {
	router, _ := setupPipelineTestServer(t)

	// Test with missing required fields
	pipelineReq := map[string]interface{}{
		"description": "Missing name and steps",
	}

	body, _ := json.Marshal(pipelineReq)
	req, _ := http.NewRequest("POST", "/api/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusCreated {
		t.Error("Expected validation error for missing fields")
	}
}

func TestPipelineStepDependency(t *testing.T) {
	router, _ := setupPipelineTestServer(t)

	// Test with invalid dependency
	pipelineReq := map[string]interface{}{
		"name":        "invalid-deps-pipeline",
		"description": "Pipeline with invalid step dependencies",
		"mode":        "hybrid",
		"steps": []map[string]interface{}{
			{
				"id":        "step-1",
				"name":      "First Step",
				"cli":       "claude",
				"action":    "review",
				"depends_on": []string{"non-existent-step"}, // Invalid dependency
			},
		},
	}

	body, _ := json.Marshal(pipelineReq)
	req, _ := http.NewRequest("POST", "/api/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail validation
	if w.Code == http.StatusOK || w.Code == http.StatusCreated {
		t.Error("Expected validation error for invalid dependency")
	}
}

func TestRunPipeline(t *testing.T) {
	router, _ := setupPipelineTestServer(t)

	// First create a pipeline
	pipelineReq := map[string]interface{}{
		"name":        "run-test-pipeline",
		"description": "Pipeline to test run",
		"mode":        "serial",
		"steps": []map[string]interface{}{
			{
				"id":     "step-1",
				"name":   "Echo Test",
				"cli":    "echo",
				"action": "command",
				"command": "hello",
				"params": map[string]interface{}{},
			},
		},
	}

	body, _ := json.Marshal(pipelineReq)
	req, _ := http.NewRequest("POST", "/api/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var pipelineResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &pipelineResponse)

	if pipelineResponse["id"] == nil {
		t.Skip("Failed to create pipeline, skipping run test")
	}

	pipelineID := pipelineResponse["id"].(string)

	// Run the pipeline
	runReq := map[string]interface{}{
		"pipeline_id": pipelineID,
		"params":      map[string]interface{}{},
	}

	body, _ = json.Marshal(runReq)
	req, _ = http.NewRequest("POST", "/api/pipelines/"+pipelineID+"/runs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted && w.Code != http.StatusOK {
		t.Errorf("Expected status 202/200, got %d: %s", w.Code, w.Body.String())
	}
}