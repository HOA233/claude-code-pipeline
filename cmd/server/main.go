package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/company/claude-pipeline/internal/api"
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/company/claude-pipeline/pkg/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init logger
	logger.Init(cfg.Log.Level)

	// Connect Redis
	redisClient := repository.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	// Verify Redis connection
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Info("Connected to Redis")

	// Init legacy services
	skillService := service.NewSkillService(redisClient, cfg.GitLab)
	taskService := service.NewTaskService(redisClient)
	cliExecutor := service.NewCLIExecutor(redisClient, cfg.CLI)
	orchestrator := service.NewOrchestrator(redisClient, cliExecutor)

	// Init new Agent orchestration services
	agentService := service.NewAgentService(redisClient)
	workflowService := service.NewWorkflowService(redisClient, agentService)
	scheduledJobService := service.NewScheduledJobService(redisClient, workflowService)
	metricsService := service.NewMetricsService(redisClient)
	logService := service.NewExecutionLogService(redisClient)
	webhookService := service.NewWebhookService(redisClient)

	// Sync default skills
	if _, err := skillService.SyncFromGitLab(context.Background()); err != nil {
		logger.Warn("Failed to sync skills: ", err)
	}

	// Load preset agents and workflows
	presetService := service.NewPresetService(redisClient, agentService, workflowService)
	if err := presetService.InitializePresets(context.Background()); err != nil {
		logger.Warn("Failed to load presets: ", err)
	}

	// Start consumers
	go cliExecutor.StartConsumer(context.Background())
	go orchestrator.StartConsumer(context.Background())

	// Start scheduled job scheduler
	scheduler := service.NewScheduler(redisClient, scheduledJobService)
	go scheduler.Start(context.Background())

	// Start HTTP server with all features
	server := api.NewServerWithAll(
		cfg,
		skillService,
		taskService,
		cliExecutor,
		orchestrator,
		redisClient,
		agentService,
		workflowService,
		scheduledJobService,
		metricsService,
		logService,
		webhookService,
	)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server...")
		cliExecutor.Stop()
		orchestrator.Stop()
		scheduler.Stop()
	}()

	logger.Info("Server starting on :8080")
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}