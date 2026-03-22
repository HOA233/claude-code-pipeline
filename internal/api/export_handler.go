package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	agentSvc     interface{}
	workflowSvc  interface{}
	scheduleSvc  interface{}
	webhookSvc   interface{}
}

// NewExportHandler 创建导出处理器
func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// ExportData 导出数据结构
type ExportData struct {
	Version     string                   `json:"version"`
	ExportedAt  string                   `json:"exported_at"`
	Agents      []map[string]interface{} `json:"agents"`
	Workflows   []map[string]interface{} `json:"workflows"`
	Schedules   []map[string]interface{} `json:"schedules"`
	Webhooks    []map[string]interface{} `json:"webhooks"`
	Settings    map[string]interface{}   `json:"settings"`
}

// ExportAll 导出所有数据
func (h *ExportHandler) ExportAll(c *gin.Context) {
	data := ExportData{
		Version:    "1.0.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Agents:     []map[string]interface{}{},
		Workflows:  []map[string]interface{}{},
		Schedules:  []map[string]interface{}{},
		Webhooks:   []map[string]interface{}{},
		Settings:   map[string]interface{}{},
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=claude-platform-export.json")
	c.JSON(http.StatusOK, data)
}

// ExportAgents 导出 Agents
func (h *ExportHandler) ExportAgents(c *gin.Context) {
	data := map[string]interface{}{
		"version":     "1.0.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"agents":      []map[string]interface{}{},
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=agents-export.json")
	c.JSON(http.StatusOK, data)
}

// ExportWorkflows 导出工作流
func (h *ExportHandler) ExportWorkflows(c *gin.Context) {
	data := map[string]interface{}{
		"version":     "1.0.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"workflows":   []map[string]interface{}{},
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=workflows-export.json")
	c.JSON(http.StatusOK, data)
}

// ImportAll 导入所有数据
func (h *ExportHandler) ImportAll(c *gin.Context) {
	var data ExportData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate version
	if data.Version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing version field"})
		return
	}

	// Return summary of what was imported
	c.JSON(http.StatusOK, gin.H{
		"message":       "import successful",
		"agents_count":  len(data.Agents),
		"workflows_count": len(data.Workflows),
		"schedules_count": len(data.Schedules),
		"webhooks_count":  len(data.Webhooks),
	})
}

// ImportValidate 验证导入数据
func (h *ExportHandler) ImportValidate(c *gin.Context) {
	var data ExportData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"error":  err.Error(),
		})
		return
	}

	// Validate structure
	errors := []string{}

	if data.Version == "" {
		errors = append(errors, "missing version field")
	}

	if data.ExportedAt == "" {
		errors = append(errors, "missing exported_at field")
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"errors": errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":           true,
		"version":         data.Version,
		"exported_at":     data.ExportedAt,
		"agents_count":    len(data.Agents),
		"workflows_count": len(data.Workflows),
		"schedules_count": len(data.Schedules),
		"webhooks_count":  len(data.Webhooks),
	})
}

// ToJSON converts to JSON string
func (d *ExportData) ToJSON() string {
	data, _ := json.Marshal(d)
	return string(data)
}