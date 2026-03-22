package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogHandler 审计日志处理器
type AuditLogHandler struct {
	logs []AuditLog
	mu   sync.RWMutex
}

// AuditLog 审计日志
type AuditLog struct {
	ID          string                 `json:"id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Actor       string                 `json:"actor"`
	ActorType   string                 `json:"actor_type"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewAuditLogHandler 创建审计日志处理器
func NewAuditLogHandler() *AuditLogHandler {
	now := time.Now()
	return &AuditLogHandler{
		logs: []AuditLog{
			{
				ID:         "log-001",
				Action:     "create",
				Resource:   "agent",
				ResourceID: "agent-001",
				Actor:      "admin",
				ActorType:  "user",
				Details:    map[string]interface{}{"name": "code-reviewer", "model": "claude-sonnet-4-6"},
				IPAddress:  "192.168.1.100",
				UserAgent:  "Mozilla/5.0",
				Status:     "success",
				Timestamp:  now.Add(-1 * time.Hour),
			},
			{
				ID:         "log-002",
				Action:     "execute",
				Resource:   "workflow",
				ResourceID: "workflow-001",
				Actor:      "scheduler",
				ActorType:  "system",
				Details:    map[string]interface{}{"name": "daily-scan", "triggered_by": "cron"},
				IPAddress:  "127.0.0.1",
				UserAgent:  "Claude-Agent-Platform/1.0",
				Status:     "success",
				Timestamp:  now.Add(-2 * time.Hour),
			},
			{
				ID:         "log-003",
				Action:     "update",
				Resource:   "schedule",
				ResourceID: "schedule-001",
				Actor:      "admin",
				ActorType:  "user",
				Details:    map[string]interface{}{"cron": "0 2 * * *", "enabled": true},
				IPAddress:  "192.168.1.100",
				UserAgent:  "Mozilla/5.0",
				Status:     "success",
				Timestamp:  now.Add(-3 * time.Hour),
			},
			{
				ID:         "log-004",
				Action:     "delete",
				Resource:   "agent",
				ResourceID: "agent-002",
				Actor:      "admin",
				ActorType:  "user",
				Details:    map[string]interface{}{"name": "old-agent"},
				IPAddress:  "192.168.1.101",
				UserAgent:  "Mozilla/5.0",
				Status:     "success",
				Timestamp:  now.Add(-5 * time.Hour),
			},
			{
				ID:         "log-005",
				Action:     "create",
				Resource:   "webhook",
				ResourceID: "webhook-001",
				Actor:      "admin",
				ActorType:  "user",
				Details:    map[string]interface{}{"url": "https://webhook.example.com/notify"},
				IPAddress:  "192.168.1.100",
				UserAgent:  "Mozilla/5.0",
				Status:     "success",
				Timestamp:  now.Add(-6 * time.Hour),
			},
			{
				ID:         "log-006",
				Action:     "execute",
				Resource:   "agent",
				ResourceID: "agent-003",
				Actor:      "api-key-001",
				ActorType:  "api",
				Details:    map[string]interface{}{"name": "test-generator", "duration": "45s"},
				IPAddress:  "10.0.0.50",
				UserAgent:  "curl/7.68.0",
				Status:     "failed",
				Timestamp:  now.Add(-8 * time.Hour),
			},
		},
	}
}

// ListAuditLogs 获取审计日志列表
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get filter params
	action := c.Query("action")
	resource := c.Query("resource")
	actor := c.Query("actor")

	var filteredLogs []AuditLog
	for _, log := range h.logs {
		if action != "" && log.Action != action {
			continue
		}
		if resource != "" && log.Resource != resource {
			continue
		}
		if actor != "" && log.Actor != actor {
			continue
		}
		filteredLogs = append(filteredLogs, log)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  filteredLogs,
		"total": len(filteredLogs),
	})
}

// GetAuditLog 获取审计日志详情
func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, log := range h.logs {
		if log.ID == id {
			c.JSON(http.StatusOK, log)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
}

// ExportAuditLogs 导出审计日志
func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	format := c.DefaultQuery("format", "json")

	if format == "csv" {
		// Return as CSV
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=audit-logs.csv")

		// Write CSV header
		c.Writer.WriteString("ID,Action,Resource,ResourceID,Actor,ActorType,Status,Timestamp\n")
		for _, log := range h.logs {
			c.Writer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
				log.ID, log.Action, log.Resource, log.ResourceID,
				log.Actor, log.ActorType, log.Status, log.Timestamp.Format(time.RFC3339)))
		}
		return
	}

	// Return as JSON
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=audit-logs.json")
	c.JSON(http.StatusOK, gin.H{
		"logs":        h.logs,
		"total":       len(h.logs),
		"exported_at": time.Now(),
	})
}

// GetAuditStats 获取审计统计
func (h *AuditLogHandler) GetAuditStats(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]int{
		"total":     len(h.logs),
		"create":    0,
		"update":    0,
		"delete":    0,
		"execute":   0,
		"success":   0,
		"failed":    0,
	}

	for _, log := range h.logs {
		stats[log.Action]++
		stats[log.Status]++
	}

	c.JSON(http.StatusOK, stats)
}