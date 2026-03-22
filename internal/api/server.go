package api

import (
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
	cfg *config.Config
}

func NewServer(cfg *config.Config, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	SetupRoutes(r, skillSvc, taskSvc, executor, orch, redis)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{Engine: r, cfg: cfg}
}

// NewServerWithAgent creates server with Agent orchestration support
func NewServerWithAgent(cfg *config.Config, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient, agentSvc *service.AgentService, workflowSvc *service.WorkflowService) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	SetupRoutesWithAgent(r, skillSvc, taskSvc, executor, orch, redis, agentSvc, workflowSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{Engine: r, cfg: cfg}
}

// NewServerWithAll creates server with all features including scheduled jobs and metrics
func NewServerWithAll(cfg *config.Config, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator, redis *repository.RedisClient, agentSvc *service.AgentService, workflowSvc *service.WorkflowService, jobSvc *service.ScheduledJobService, metricsSvc *service.MetricsService, logSvc *service.ExecutionLogService, webhookSvc *service.WebhookService, templateSvc *service.TemplateService) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	SetupRoutesWithAll(r, skillSvc, taskSvc, executor, orch, redis, agentSvc, workflowSvc, jobSvc, metricsSvc, logSvc, webhookSvc, templateSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{Engine: r, cfg: cfg}
}

func (s *Server) Run() error {
	return s.Engine.Run(":8080")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}