package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// AgentHandler Agent API 处理器
type AgentHandler struct {
	agentSvc *service.AgentService
}

// NewAgentHandler 创建 Agent Handler
func NewAgentHandler(agentSvc *service.AgentService) *AgentHandler {
	return &AgentHandler{agentSvc: agentSvc}
}

// CreateAgent 创建 Agent
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req model.AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, err := h.agentSvc.CreateAgent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// GetAgent 获取 Agent
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")

	agent, err := h.agentSvc.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// ListAgents 列出 Agent
func (h *AgentHandler) ListAgents(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	category := c.Query("category")

	agents, err := h.agentSvc.ListAgents(c.Request.Context(), tenantID, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"total":  len(agents),
	})
}

// UpdateAgent 更新 Agent
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")

	var req model.AgentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, err := h.agentSvc.UpdateAgent(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// DeleteAgent 删除 Agent
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	if err := h.agentSvc.DeleteAgent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// ExecuteAgent 执行 Agent
func (h *AgentHandler) ExecuteAgent(c *gin.Context) {
	id := c.Param("id")

	var req model.AgentExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.agentSvc.ExecuteAgent(c.Request.Context(), id, &req)
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

// TestAgent 测试 Agent
func (h *AgentHandler) TestAgent(c *gin.Context) {
	id := c.Param("id")

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.agentSvc.TestAgent(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}