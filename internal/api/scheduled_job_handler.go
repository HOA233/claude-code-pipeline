package api

import (
	"fmt"
	"net/http"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// ScheduledJobHandler 定时任务 API 处理器
type ScheduledJobHandler struct {
	jobSvc *service.ScheduledJobService
}

// NewScheduledJobHandler 创建处理器
func NewScheduledJobHandler(jobSvc *service.ScheduledJobService) *ScheduledJobHandler {
	return &ScheduledJobHandler{jobSvc: jobSvc}
}

// CreateJob 创建定时任务
func (h *ScheduledJobHandler) CreateJob(c *gin.Context) {
	var req model.ScheduledJobCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.jobSvc.CreateJob(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// GetJob 获取定时任务
func (h *ScheduledJobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.jobSvc.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs 列出定时任务
func (h *ScheduledJobHandler) ListJobs(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	jobs, err := h.jobSvc.ListJobs(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

// UpdateJob 更新定时任务
func (h *ScheduledJobHandler) UpdateJob(c *gin.Context) {
	id := c.Param("id")

	var req model.ScheduledJobUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.jobSvc.UpdateJob(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DeleteJob 删除定时任务
func (h *ScheduledJobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")

	if err := h.jobSvc.DeleteJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job deleted"})
}

// EnableJob 启用定时任务
func (h *ScheduledJobHandler) EnableJob(c *gin.Context) {
	id := c.Param("id")

	if err := h.jobSvc.EnableJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job enabled"})
}

// DisableJob 禁用定时任务
func (h *ScheduledJobHandler) DisableJob(c *gin.Context) {
	id := c.Param("id")

	if err := h.jobSvc.DisableJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job disabled"})
}

// TriggerJob 手动触发任务
func (h *ScheduledJobHandler) TriggerJob(c *gin.Context) {
	id := c.Param("id")

	execution, err := h.jobSvc.TriggerJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetJobHistory 获取执行历史
func (h *ScheduledJobHandler) GetJobHistory(c *gin.Context) {
	id := c.Param("id")

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if pi, err := parseIntParam(p); err == nil && pi > 0 {
			page = pi
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if psi, err := parseIntParam(ps); err == nil && psi > 0 {
			pageSize = psi
		}
	}

	result, err := h.jobSvc.GetJobHistory(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseIntParam(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}