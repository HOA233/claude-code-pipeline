package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// GetStatus returns service status
func GetStatus(executor *service.CLIExecutor, skillSvc *service.SkillService, orch *service.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		skills, _ := skillSvc.GetAllSkills(c.Request.Context())
		pipelines, _ := orch.ListPipelines(c.Request.Context())
		runs, _ := orch.ListRuns(c.Request.Context())

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"cli":    executor.GetStatus(),
			"skills": gin.H{
				"loaded": len(skills),
			},
			"pipelines": gin.H{
				"total": len(pipelines),
			},
			"runs": gin.H{
				"total": len(runs),
			},
		})
	}
}