package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	metricsSvc *service.MetricsService
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(metricsSvc *service.MetricsService) *StatsHandler {
	return &StatsHandler{metricsSvc: metricsSvc}
}

// GetSystemMetrics 获取系统指标
func (h *StatsHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := h.metricsSvc.GetSystemMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetExecutionTrend 获取执行趋势
func (h *StatsHandler) GetExecutionTrend(c *gin.Context) {
	days := 7 // 默认 7 天
	if d := c.Query("days"); d != "" {
		// 解析天数
	}

	trends, err := h.metricsSvc.GetExecutionTrend(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"days":   days,
	})
}

// GetWorkflowStats 获取工作流统计
func (h *StatsHandler) GetWorkflowStats(c *gin.Context) {
	stats, err := h.metricsSvc.GetWorkflowStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": stats,
		"total":     len(stats),
	})
}

// GetHealthStatus 获取健康状态
func (h *StatsHandler) GetHealthStatus(c *gin.Context) {
	status := h.metricsSvc.GetHealthStatus(c.Request.Context())

	if status.Status == "healthy" {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}