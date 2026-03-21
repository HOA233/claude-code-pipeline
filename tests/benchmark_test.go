package main

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/api"
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// setupBenchmarkServer sets up a test server for benchmarking
func setupBenchmarkServer(b *testing.B) (*gin.Engine, *repository.RedisClient) {
	gin.SetMode(gin.TestMode)

	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)

	ctx := b.Context()
	if err := redisClient.Ping(ctx); err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	skillSvc := service.NewSkillService(redisClient, config.GitLabConfig{})
	taskSvc := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, config.CLIConfig{})
	orchestrator := service.NewOrchestrator(redisClient, executor)

	// Sync skills for tests
	skillSvc.SyncFromGitLab(ctx)

	router := gin.New()
	api.SetupRoutes(router, skillSvc, taskSvc, executor, orchestrator)

	return router, redisClient
}

// BenchmarkCreateTask benchmarks task creation
func BenchmarkCreateTask(b *testing.B) {
	router, _ := setupBenchmarkServer(b)
	ctx := b.Context()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &model.TaskCreateRequest{
			SkillID: "code-review",
			Parameters: map[string]interface{}{
				"target": "src/",
				"depth":  "quick",
			},
		}
		_, err := taskSvc.CreateTask(ctx, req)
		if err != nil {
			b.Errorf("Failed to create task: %v", err)
		}
	}
}

// BenchmarkPipelineCreation benchmarks pipeline creation
func BenchmarkPipelineCreation(b *testing.B) {
	router, _ := setupBenchmarkServer(b)
	ctx := b.Context()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &model.PipelineCreateRequest{
			Name:        "benchmark-pipeline",
			Description: "Benchmark test pipeline",
			Mode:        model.ModeSerial,
			Steps: []model.Step{
				{
					ID:      "step-1",
					Name:    "Test Step",
					CLI:     "echo",
					Action:  "command",
					Command: "hello",
					Params:  map[string]interface{}{},
				},
			},
		}
		_, err := orchestrator.CreatePipeline(ctx, req)
		if err != nil {
			b.Errorf("Failed to create pipeline: %v", err)
		}
	}
}

// BenchmarkRedisSaveSkill benchmarks saving a skill to Redis
func BenchmarkRedisSaveSkill(b *testing.B) {
	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)
	ctx := b.Context()

	if err := redisClient.Ping(ctx); err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	skill := &model.Skill{
		ID:          "bench-skill",
		Name:        "Benchmark Skill",
		Description: "Test skill for benchmarking",
		Version:     "1.0.0",
		Category:    "test",
		Enabled:     true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := redisClient.SaveSkill(ctx, skill)
		if err != nil {
			b.Errorf("Failed to save skill: %v", err)
		}
	}
}

// BenchmarkRedisGetSkill benchmarks getting a skill from Redis
func BenchmarkRedisGetSkill(b *testing.B) {
	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)
	ctx := b.Context()

	if err := redisClient.Ping(ctx); err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := redisClient.GetSkill(ctx, "bench-skill")
		if err != nil {
			b.Errorf("Failed to get skill: %v", err)
		}
	}
}

// BenchmarkRedisGetAllSkills benchmarks getting all skills
func BenchmarkRedisGetAllSkills(b *testing.B) {
	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)
	ctx := b.Context()

	if err := redisClient.Ping(ctx); err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := redisClient.GetAllSkills(ctx)
		if err != nil {
			b.Errorf("Failed to get all skills: %v", err)
		}
	}
}

// BenchmarkAPISkillsEndpoint benchmarks the skills API endpoint
func BenchmarkAPISkillsEndpoint(b *testing.B) {
	router, _ := setupBenchmarkServer(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/skills", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIStatusEndpoint benchmarks the status API endpoint
func BenchmarkAPIStatusEndpoint(b *testing.B) {
	router, _ := setupBenchmarkServer(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIPipelinesEndpoint benchmarks the pipelines API endpoint
func BenchmarkAPIPipelinesEndpoint(b *testing.B) {
	router, _ := setupBenchmarkServer(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/pipelines", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRateLimiter benchmarks the rate limiter
func BenchmarkRateLimiter(b *testing.B) {
	rl := api.NewRateLimiter(1000, time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rl.Allow("test-client")
	}
}

// BenchmarkParallelRequests benchmarks parallel request handling
func BenchmarkParallelRequests(b *testing.B) {
	router, _ := setupBenchmarkServer(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/status", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}