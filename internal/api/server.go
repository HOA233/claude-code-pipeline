package api

import (
	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
	cfg *config.Config
}

func NewServer(cfg *config.Config, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor, orch *service.Orchestrator) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	SetupRoutes(r, skillSvc, taskSvc, executor, orch)

	// Health check
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