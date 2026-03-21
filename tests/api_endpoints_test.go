package tests

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
	"github.com/stretchr/testify/assert"
)

// TestServer holds test server and dependencies
type TestServer struct {
	Router   *gin.Engine
	Redis    *repository.RedisClient
	SkillSvc *service.SkillService
	TaskSvc  *service.TaskService
	Executor *service.CLIExecutor
	Orch     *service.Orchestrator
}

// SetupTestServer creates a test server with mocked dependencies
func SetupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	// Use test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		Redis: config.RedisConfig{
			Addr: "localhost:6379",
			DB:   1, // Use different DB for tests
		},
		CLI: config.CLIConfig{
			MaxConcurrency: 5,
			DefaultTimeout: 60,
			ClaudePath:     "claude",
		},
	}

	// Initialize Redis client (skip if not available)
	redisClient := repository.NewRedisClient(cfg.Redis)

	// Initialize services
	skillSvc := service.NewSkillService(redisClient)
	taskSvc := service.NewTaskService(redisClient, nil)
	executor := service.NewCLIExecutor(cfg.CLI, nil)
	orch := service.NewOrchestrator(redisClient, nil)

	// Setup router
	router := gin.New()
	api.SetupRoutes(router, skillSvc, taskSvc, executor, orch, redisClient)

	return &TestServer{
		Router:   router,
		Redis:    redisClient,
		SkillSvc: skillSvc,
		TaskSvc:  taskSvc,
		Executor: executor,
		Orch:     orch,
	}
}

func TestHealthEndpoint(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}

func TestReadyEndpoint(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListSkills(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/skills", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "skills")
}

func TestListTasks(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "tasks")
}

func TestListPipelines(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatePipeline(t *testing.T) {
	ts := SetupTestServer(t)

	pipeline := map[string]interface{}{
		"name": "test-pipeline",
		"mode": "serial",
		"steps": []map[string]interface{}{
			{
				"id":     "step1",
				"cli":    "echo",
				"action": "test",
				"params": map[string]interface{}{},
			},
		},
	}
	body, _ := json.Marshal(pipeline)

	req, _ := http.NewRequest("POST", "/api/pipelines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	// Should return 200 or 201
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated)
}

func TestListRuns(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/runs", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatusEndpoint(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateTaskValidation(t *testing.T) {
	ts := SetupTestServer(t)

	// Missing skill_id
	task := map[string]interface{}{
		"parameters": map[string]interface{}{
			"target": "src/",
		},
	}
	body, _ := json.Marshal(task)

	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	// Should return 400 for validation error
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnprocessableEntity)
}

func TestGetNonExistentTask(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/tasks/non-existent-id", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetNonExistentPipeline(t *testing.T) {
	ts := SetupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/pipelines/non-existent-id", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}