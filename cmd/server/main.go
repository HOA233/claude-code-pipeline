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

	// Init services
	skillService := service.NewSkillService(redisClient, cfg.GitLab)
	taskService := service.NewTaskService(redisClient)
	executor := service.NewCLIExecutor(redisClient, cfg.CLI)

	// Sync default skills
	if _, err := skillService.SyncFromGitLab(context.Background()); err != nil {
		logger.Warn("Failed to sync skills: ", err)
	}

	// Start task consumer
	go executor.StartConsumer(context.Background())

	// Start HTTP server
	server := api.NewServer(cfg, skillService, taskService, executor)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server...")
		executor.Stop()
		server.Engine.Context().
			Done()
	}()

	logger.Info("Server starting on :8080")
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}