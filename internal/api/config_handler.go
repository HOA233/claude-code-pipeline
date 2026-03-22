package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 系统配置处理器
type ConfigHandler struct {
	config map[string]interface{}
	mu     sync.RWMutex
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		config: map[string]interface{}{
			"platform_name":     "Claude Code Agent Platform",
			"version":           "1.0.0",
			"default_model":     "claude-sonnet-4-6",
			"max_executions":    1000,
			"execution_timeout": 3600,
			"enable_webhooks":   true,
			"enable_scheduling": true,
			"log_level":         "info",
			"features": map[string]bool{
				"agent_creation":  true,
				"workflow_editor": true,
				"execution_logs":  true,
				"metrics":         true,
			},
		},
	}
}

// GetConfig 获取系统配置
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"config": h.config,
	})
}

// UpdateConfig 更新系统配置
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	for key, value := range updates {
		h.config[key] = value
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "config updated",
		"config":  h.config,
	})
}

// GetFeatures 获取功能开关
func (h *ConfigHandler) GetFeatures(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	features, ok := h.config["features"].(map[string]bool)
	if !ok {
		features = make(map[string]bool)
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
	})
}

// ToggleFeature 切换功能开关
func (h *ConfigHandler) ToggleFeature(c *gin.Context) {
	feature := c.Param("feature")

	h.mu.Lock()
	defer h.mu.Unlock()

	features, ok := h.config["features"].(map[string]bool)
	if !ok {
		features = make(map[string]bool)
		h.config["features"] = features
	}

	features[feature] = !features[feature]

	c.JSON(http.StatusOK, gin.H{
		"feature":  feature,
		"enabled":  features[feature],
		"features": features,
	})
}

// GetModels 获取可用模型列表
func (h *ConfigHandler) GetModels(c *gin.Context) {
	models := []map[string]interface{}{
		{
			"id":          "claude-sonnet-4-6",
			"name":        "Claude Sonnet 4.6",
			"description": "Best balance of speed and intelligence",
			"context":     200000,
			"max_output":  16384,
		},
		{
			"id":          "claude-opus-4-6",
			"name":        "Claude Opus 4.6",
			"description": "Most capable model for complex tasks",
			"context":     200000,
			"max_output":  16384,
		},
		{
			"id":          "claude-haiku-4-5",
			"name":        "Claude Haiku 4.5",
			"description": "Fastest model for simple tasks",
			"context":     200000,
			"max_output":  8192,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// GetCategories 获取 Agent 分类列表
func (h *ConfigHandler) GetCategories(c *gin.Context) {
	categories := []map[string]interface{}{
		{"id": "general", "name": "通用", "color": "#8c8c8c"},
		{"id": "code-review", "name": "代码审查", "color": "#52c41a"},
		{"id": "testing", "name": "测试", "color": "#1890ff"},
		{"id": "security", "name": "安全", "color": "#f5222d"},
		{"id": "documentation", "name": "文档", "color": "#722ed1"},
		{"id": "refactoring", "name": "重构", "color": "#faad14"},
		{"id": "performance", "name": "性能", "color": "#13c2c2"},
		{"id": "debugging", "name": "调试", "color": "#eb2f96"},
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}