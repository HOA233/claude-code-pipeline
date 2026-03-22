package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// ExecutionDetailHandler 执行详情处理器
type ExecutionDetailHandler struct {
	workflowSvc *service.WorkflowService
	logSvc      *service.ExecutionLogService
}

// NewExecutionDetailHandler 创建处理器
func NewExecutionDetailHandler(workflowSvc *service.WorkflowService, logSvc *service.ExecutionLogService) *ExecutionDetailHandler {
	return &ExecutionDetailHandler{
		workflowSvc: workflowSvc,
		logSvc:      logSvc,
	}
}

// GetExecutionLogs 获取执行日志
func (h *ExecutionDetailHandler) GetExecutionLogs(c *gin.Context) {
	id := c.Param("id")

	// 验证执行存在
	_, err := h.workflowSvc.GetExecution(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// 获取日志
	limit := int64(100)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := h.logSvc.GetLogs(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": id,
		"logs":         logs,
		"total":        len(logs),
	})
}

// StreamExecutionLogs SSE 流式日志
func (h *ExecutionDetailHandler) StreamExecutionLogs(c *gin.Context) {
	id := c.Param("id")

	// 验证执行存在
	_, err := h.workflowSvc.GetExecution(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// 设置 SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 获取日志流
	logChan := h.logSvc.StreamLogs(c.Request.Context(), id)

	for log := range logChan {
		data, _ := jsonMarshal(log)
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", data))
		c.Writer.Flush()
	}
}

// RetryExecution 重试执行
func (h *ExecutionDetailHandler) RetryExecution(c *gin.Context) {
	id := c.Param("id")

	exec, err := h.workflowSvc.GetExecution(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// 只能重试失败的执行
	if exec.Status != model.ExecutionStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only retry failed executions"})
		return
	}

	// 创建新的执行
	newExec, err := h.workflowSvc.ExecuteWorkflow(c.Request.Context(), &model.ExecutionCreateRequest{
		WorkflowID: exec.WorkflowID,
		Async:      true,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "execution retry started",
		"original_id":    id,
		"new_execution":  newExec,
	})
}

// GetExecutionDetails 获取执行详情（包含步骤详情）
func (h *ExecutionDetailHandler) GetExecutionDetails(c *gin.Context) {
	id := c.Param("id")

	exec, err := h.workflowSvc.GetExecution(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// 构建详情响应
	details := map[string]interface{}{
		"execution":    exec,
		"node_details": make([]map[string]interface{}, 0),
	}

	// 添加节点详情
	for nodeID, result := range exec.NodeResults {
		nodeDetail := map[string]interface{}{
			"node_id":    nodeID,
			"agent_id":   result.AgentID,
			"status":     result.Status,
			"duration":   result.Duration,
			"retries":    result.Retries,
			"error":      result.Error,
			"started_at": result.StartedAt,
			"completed_at": result.CompletedAt,
		}

		if result.Output != nil {
			nodeDetail["output"] = result.Output
		}

		details["node_details"] = append(details["node_details"].([]map[string]interface{}), nodeDetail)
	}

	c.JSON(http.StatusOK, details)
}

// GetExecutionMetrics 获取执行指标
func (h *ExecutionDetailHandler) GetExecutionMetrics(c *gin.Context) {
	filter := &model.ExecutionFilter{
		Page:     1,
		PageSize: 1000, // 获取足够多的数据计算指标
	}

	result, err := h.workflowSvc.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 计算指标
	metrics := calculateMetrics(result.Executions)

	c.JSON(http.StatusOK, metrics)
}

func calculateMetrics(executions []model.Execution) map[string]interface{} {
	total := len(executions)
	if total == 0 {
		return map[string]interface{}{
			"total":          0,
			"success_rate":   0,
			"avg_duration":   0,
			"by_status":      map[string]int{},
			"by_workflow":    map[string]int{},
		}
	}

	statusCounts := make(map[string]int)
	workflowCounts := make(map[string]int)
	var totalDuration int64
	var successCount int

	for _, exec := range executions {
		statusCounts[string(exec.Status)]++
		workflowCounts[exec.WorkflowName]++
		totalDuration += exec.Duration

		if exec.Status == model.ExecutionStatusCompleted {
			successCount++
		}
	}

	avgDuration := totalDuration / int64(total)
	successRate := float64(successCount) / float64(total) * 100

	return map[string]interface{}{
		"total":          total,
		"success_rate":   successRate,
		"avg_duration":   avgDuration,
		"by_status":      statusCounts,
		"by_workflow":    workflowCounts,
		"success_count":  successCount,
		"failed_count":   statusCounts["failed"],
		"running_count":  statusCounts["running"],
	}
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}