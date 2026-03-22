package api

import (
	"net/http"
	"strconv"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// WebhookHandler Webhook API 处理器
type WebhookHandler struct {
	webhookSvc *service.WebhookService
}

// NewWebhookHandler 创建 Webhook Handler
func NewWebhookHandler(webhookSvc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookSvc: webhookSvc}
}

// CreateWebhook 创建 Webhook
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var req service.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	if err := h.webhookSvc.CreateWebhook(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetWebhook 获取 Webhook
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id := c.Param("id")

	webhook, err := h.webhookSvc.GetWebhook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// ListWebhooks 列出 Webhooks
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	webhooks, err := h.webhookSvc.ListWebhooks(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"webhooks": webhooks,
		"total":    len(webhooks),
	})
}

// UpdateWebhook 更新 Webhook
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.webhookSvc.UpdateWebhook(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	webhook, _ := h.webhookSvc.GetWebhook(c.Request.Context(), id)
	c.JSON(http.StatusOK, webhook)
}

// DeleteWebhook 删除 Webhook
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	id := c.Param("id")

	if err := h.webhookSvc.DeleteWebhook(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook deleted"})
}

// GetWebhookDeliveries 获取 Webhook 投递记录
func (h *WebhookHandler) GetWebhookDeliveries(c *gin.Context) {
	id := c.Param("id")

	limit := int64(100)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	deliveries, err := h.webhookSvc.GetDeliveries(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deliveries": deliveries,
		"total":      len(deliveries),
	})
}