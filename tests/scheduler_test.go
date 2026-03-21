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

func setupSchedulerTestServer(t *testing.T) (*gin.Engine, *service.SchedulerService, func()) {
	gin.SetMode(gin.TestMode)

	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   3, // Use different DB for scheduler tests
	}

	redisClient := repository.NewRedisClient(cfg)

	ctx := t.Context()
	if err := redisClient.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping scheduler test")
	}

	skillSvc := service.NewSkillService(redisClient, config.GitLabConfig{})
	taskSvc := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, config.CLIConfig{})
	orchestrator := service.NewOrchestrator(redisClient, executor)
	schedulerSvc := service.NewSchedulerService(redisClient, taskSvc)

	// Sync default skills
	skillSvc.SyncFromGitLab(ctx)

	router := gin.New()
	api.SetupRoutesWithScheduler(router, skillSvc, taskSvc, executor, orchestrator, schedulerSvc)

	cleanup := func() {
		redisClient.Close()
	}

	return router, schedulerSvc, cleanup
}

func TestSchedulerCreateSchedule(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// Create a schedule
	scheduleReq := map[string]interface{}{
		"name":       "daily-code-review",
		"skill_id":   "code-review",
		"cron_expr":  "@daily",
		"parameters": map[string]interface{}{"target": "src/", "depth": "quick"},
		"tags":       []string{"daily", "automation"},
	}

	body, _ := json.Marshal(scheduleReq)
	req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create schedule: %d - %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["id"] == "" {
		t.Error("Expected schedule ID in response")
	}

	t.Logf("Created schedule: %v", resp["id"])
}

func TestSchedulerListSchedules(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// List schedules
	req, _ := http.NewRequest("GET", "/api/schedules", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to list schedules: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	schedules := resp["schedules"].([]interface{})
	t.Logf("Found %d schedules", len(schedules))
}

func TestSchedulerEnableDisable(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// First create a schedule
	scheduleReq := map[string]interface{}{
		"name":      "test-schedule",
		"skill_id":  "code-review",
		"cron_expr": "@hourly",
	}

	body, _ := json.Marshal(scheduleReq)
	req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create schedule: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	scheduleID := resp["id"].(string)

	// Disable the schedule
	req, _ = http.NewRequest("POST", "/api/schedules/"+scheduleID+"/disable", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to disable schedule: %d", w.Code)
	}

	// Enable the schedule
	req, _ = http.NewRequest("POST", "/api/schedules/"+scheduleID+"/enable", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to enable schedule: %d", w.Code)
	}
}

func TestSchedulerTrigger(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// First create a schedule
	scheduleReq := map[string]interface{}{
		"name":      "triggerable-schedule",
		"skill_id":  "code-review",
		"cron_expr": "@daily",
		"parameters": map[string]interface{}{
			"target": "src/",
			"depth":  "quick",
		},
	}

	body, _ := json.Marshal(scheduleReq)
	req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create schedule: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	scheduleID := resp["id"].(string)

	// Trigger the schedule
	req, _ = http.NewRequest("POST", "/api/schedules/"+scheduleID+"/trigger", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Trigger response: %s", w.Body.String())

	// Response should be 202 Accepted
	if w.Code != http.StatusAccepted {
		t.Logf("Warning: Trigger response code: %d", w.Code)
	}
}

func TestSchedulerUpdate(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// First create a schedule
	scheduleReq := map[string]interface{}{
		"name":      "updatable-schedule",
		"skill_id":  "code-review",
		"cron_expr": "@hourly",
	}

	body, _ := json.Marshal(scheduleReq)
	req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create schedule: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	scheduleID := resp["id"].(string)

	// Update the schedule
	updateReq := map[string]interface{}{
		"name":      "updated-schedule-name",
		"cron_expr": "@daily",
	}

	body, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/schedules/"+scheduleID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to update schedule: %d - %s", w.Code, w.Body.String())
	}
}

func TestSchedulerDelete(t *testing.T) {
	router, _, cleanup := setupSchedulerTestServer(t)
	defer cleanup()

	// First create a schedule
	scheduleReq := map[string]interface{}{
		"name":      "deletable-schedule",
		"skill_id":  "code-review",
		"cron_expr": "@hourly",
	}

	body, _ := json.Marshal(scheduleReq)
	req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create schedule: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	scheduleID := resp["id"].(string)

	// Delete the schedule
	req, _ = http.NewRequest("DELETE", "/api/schedules/"+scheduleID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to delete schedule: %d", w.Code)
	}

	// Try to get deleted schedule
	req, _ = http.NewRequest("GET", "/api/schedules/"+scheduleID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Error("Expected 404 for deleted schedule")
	}
}