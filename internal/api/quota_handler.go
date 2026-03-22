package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// QuotaHandler 配额处理器
type QuotaHandler struct {
	quotas    []Quota
	costs     []CostRecord
	mu        sync.RWMutex
}

// Quota 配额定义
type Quota struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ResourceType string    `json:"resource_type"`
	Limit        int64     `json:"limit"`
	Used         int64     `json:"used"`
	Unit         string    `json:"unit"`
	Period       string    `json:"period"`
	ResetAt      time.Time `json:"reset_at"`
}

// CostRecord 成本记录
type CostRecord struct {
	ID            string    `json:"id"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	ResourceName  string    `json:"resource_name"`
	TokensInput   int64     `json:"tokens_input"`
	TokensOutput  int64     `json:"tokens_output"`
	Cost          float64   `json:"cost"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewQuotaHandler 创建配额处理器
func NewQuotaHandler() *QuotaHandler {
	now := time.Now()
	return &QuotaHandler{
		quotas: []Quota{
			{
				ID:           "quota-1",
				Name:         "每日执行次数",
				ResourceType: "execution",
				Limit:        1000,
				Used:         456,
				Unit:         "次",
				Period:       "daily",
				ResetAt:      now.Add(24 * time.Hour),
			},
			{
				ID:           "quota-2",
				Name:         "每月 Token 配额",
				ResourceType: "tokens",
				Limit:        10000000,
				Used:         3456789,
				Unit:         "tokens",
				Period:       "monthly",
				ResetAt:      now.AddDate(0, 1, 0),
			},
			{
				ID:           "quota-3",
				Name:         "并发执行数",
				ResourceType: "concurrent",
				Limit:        10,
				Used:         3,
				Unit:         "个",
				Period:       "daily",
				ResetAt:      now.Add(24 * time.Hour),
			},
			{
				ID:           "quota-4",
				Name:         "存储空间",
				ResourceType: "storage",
				Limit:        100,
				Used:         45,
				Unit:         "GB",
				Period:       "monthly",
				ResetAt:      now.AddDate(0, 1, 0),
			},
		},
		costs: []CostRecord{
			{
				ID:            "cost-1",
				ResourceType:  "agent",
				ResourceID:    "agent-001",
				ResourceName:  "code-reviewer",
				TokensInput:   1500,
				TokensOutput:  3200,
				Cost:          0.089,
				Timestamp:     now.Add(-1 * time.Hour),
			},
			{
				ID:            "cost-2",
				ResourceType:  "workflow",
				ResourceID:    "workflow-001",
				ResourceName:  "code-review-workflow",
				TokensInput:   5200,
				TokensOutput:  8400,
				Cost:          0.234,
				Timestamp:     now.Add(-2 * time.Hour),
			},
			{
				ID:            "cost-3",
				ResourceType:  "agent",
				ResourceID:    "agent-002",
				ResourceName:  "test-generator",
				TokensInput:   2300,
				TokensOutput:  4100,
				Cost:          0.112,
				Timestamp:     now.Add(-4 * time.Hour),
			},
		},
	}
}

// GetQuotas 获取配额列表
func (h *QuotaHandler) GetQuotas(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"quotas": h.quotas,
		"total":  len(h.quotas),
	})
}

// GetCosts 获取成本记录
func (h *QuotaHandler) GetCosts(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get time range from query
	timeRange := c.DefaultQuery("range", "week")

	// Filter costs by time range
	var filteredCosts []CostRecord
	now := time.Now()
	cutoff := now

	switch timeRange {
	case "today":
		cutoff = now.Add(-24 * time.Hour)
	case "week":
		cutoff = now.Add(-7 * 24 * time.Hour)
	case "month":
		cutoff = now.AddDate(0, -1, 0)
	}

	for _, cost := range h.costs {
		if cost.Timestamp.After(cutoff) {
			filteredCosts = append(filteredCosts, cost)
		}
	}

	// Calculate totals
	var totalCost float64
	var totalTokensInput, totalTokensOutput int64
	for _, cost := range filteredCosts {
		totalCost += cost.Cost
		totalTokensInput += cost.TokensInput
		totalTokensOutput += cost.TokensOutput
	}

	c.JSON(http.StatusOK, gin.H{
		"costs":              filteredCosts,
		"total":              len(filteredCosts),
		"summary": gin.H{
			"total_cost":         totalCost,
			"total_tokens_input": totalTokensInput,
			"total_tokens_output": totalTokensOutput,
		},
	})
}

// UpdateQuota 更新配额
func (h *QuotaHandler) UpdateQuota(c *gin.Context) {
	quotaID := c.Param("id")

	var req struct {
		Limit int64 `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for i, quota := range h.quotas {
		if quota.ID == quotaID {
			h.quotas[i].Limit = req.Limit
			c.JSON(http.StatusOK, gin.H{
				"quota": h.quotas[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
}

// GetQuotaSummary 获取配额摘要
func (h *QuotaHandler) GetQuotaSummary(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Calculate totals
	var totalCost float64
	var totalTokensInput, totalTokensOutput int64

	for _, cost := range h.costs {
		totalCost += cost.Cost
		totalTokensInput += cost.TokensInput
		totalTokensOutput += cost.TokensOutput
	}

	// Get usage percentages
	usageStats := make(map[string]float64)
	for _, quota := range h.quotas {
		if quota.Limit > 0 {
			usageStats[quota.ResourceType] = float64(quota.Used) / float64(quota.Limit) * 100
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"quotas":            h.quotas,
		"total_cost":        totalCost,
		"total_tokens":      totalTokensInput + totalTokensOutput,
		"total_tokens_in":   totalTokensInput,
		"total_tokens_out":  totalTokensOutput,
		"execution_count":   len(h.costs),
		"usage_percentages": usageStats,
	})
}