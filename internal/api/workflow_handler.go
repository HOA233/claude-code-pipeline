package api

import (
	"fmt"
	"net/http"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler Workflow API 处理器
type WorkflowHandler struct {
	workflowSvc *service.WorkflowService
}

// NewWorkflowHandler 创建 Workflow Handler
func NewWorkflowHandler(workflowSvc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowSvc: workflowSvc}
}

// CreateWorkflow 创建 Workflow
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req model.WorkflowCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.workflowSvc.CreateWorkflow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// GetWorkflow 获取 Workflow
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	workflow, err := h.workflowSvc.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// ListWorkflows 列出 Workflow
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	workflows, err := h.workflowSvc.ListWorkflows(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"total":     len(workflows),
	})
}

// UpdateWorkflow 更新 Workflow
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req model.WorkflowUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.workflowSvc.UpdateWorkflow(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DeleteWorkflow 删除 Workflow
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")

	if err := h.workflowSvc.DeleteWorkflow(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted"})
}

// ExecutionHandler Execution API 处理器
type ExecutionHandler struct {
	workflowSvc *service.WorkflowService
}

// NewExecutionHandler 创建 Execution Handler
func NewExecutionHandler(workflowSvc *service.WorkflowService) *ExecutionHandler {
	return &ExecutionHandler{workflowSvc: workflowSvc}
}

// ExecuteWorkflow 执行 Workflow
func (h *ExecutionHandler) ExecuteWorkflow(c *gin.Context) {
	var req model.ExecutionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.workflowSvc.ExecuteWorkflow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Async {
		c.JSON(http.StatusAccepted, execution)
	} else {
		c.JSON(http.StatusOK, execution)
	}
}

// GetExecution 获取 Execution
func (h *ExecutionHandler) GetExecution(c *gin.Context) {
	id := c.Param("id")

	execution, err := h.workflowSvc.GetExecution(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// ListExecutions 列出 Execution
func (h *ExecutionHandler) ListExecutions(c *gin.Context) {
	filter := &model.ExecutionFilter{
		Status:     model.ExecutionStatus(c.Query("status")),
		WorkflowID: c.Query("workflow_id"),
		TenantID:   c.Query("tenant_id"),
		Page:       1,
		PageSize:   20,
	}

	// Parse pagination
	if page := c.Query("page"); page != "" {
		if p, err := parseInt(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := parseInt(pageSize); err == nil && ps > 0 {
			filter.PageSize = ps
		}
	}

	result, err := h.workflowSvc.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelExecution 取消 Execution
func (h *ExecutionHandler) CancelExecution(c *gin.Context) {
	id := c.Param("id")

	if err := h.workflowSvc.CancelExecution(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "execution cancelled"})
}

// PauseExecution 暂停 Execution
func (h *ExecutionHandler) PauseExecution(c *gin.Context) {
	id := c.Param("id")

	if err := h.workflowSvc.PauseExecution(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "execution paused"})
}

// ResumeExecution 恢复 Execution
func (h *ExecutionHandler) ResumeExecution(c *gin.Context) {
	id := c.Param("id")

	if err := h.workflowSvc.ResumeExecution(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "execution resumed"})
}

// CancelAllExecutions 取消所有 Execution
func (h *ExecutionHandler) CancelAllExecutions(c *gin.Context) {
	status := model.ExecutionStatus(c.Query("status"))
	if status == "" {
		status = model.ExecutionStatusRunning
	}

	count, err := h.workflowSvc.CancelAllExecutions(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "executions cancelled",
		"count":   count,
	})
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}