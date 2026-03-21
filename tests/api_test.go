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
	"github.com/redis/go-redis/v9"
)

func setupTestServer(t *testing.T) (*gin.Engine, *repository.RedisClient) {
	gin.SetMode(gin.TestMode)

	// Use mock Redis for testing
	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for tests
	}

	redisClient := repository.NewRedisClient(cfg)

	// Try to connect, skip test if Redis not available
	ctx := t.Context()
	if err := redisClient.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	skillSvc := service.NewSkillService(redisClient, config.GitLabConfig{})
	taskSvc := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, config.CLIConfig{})

	// Sync skills for tests
	skillSvc.SyncFromGitLab(ctx)

	router := gin.New()
	api.SetupRoutes(router, skillSvc, taskSvc, executor)

	return router, redisClient
}

func TestListSkills(t *testing.T) {
	router, _ := setupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/skills", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	skills, ok := response["skills"].([]interface{})
	if !ok {
		t.Error("Expected skills array in response")
	}

	if len(skills) == 0 {
		t.Error("Expected at least one skill")
	}
}

func TestCreateTask(t *testing.T) {
	router, _ := setupTestServer(t)

	taskReq := map[string]interface{}{
		"skill_id": "code-review",
		"parameters": map[string]interface{}{
			"target": "src/",
			"depth":  "standard",
		},
	}

	body, _ := json.Marshal(taskReq)
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["id"] == nil {
		t.Error("Expected task ID in response")
	}
}

func TestGetTaskNotFound(t *testing.T) {
	router, _ := setupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestStatusEndpoint(t *testing.T) {
	router, _ := setupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Error("Expected healthy status")
	}
}