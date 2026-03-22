package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
)

// WebhookService Webhook 管理服务
type WebhookService struct {
	redis *repository.RedisClient
}

// NewWebhookService 创建 Webhook 服务
func NewWebhookService(redis *repository.RedisClient) *WebhookService {
	return &WebhookService{redis: redis}
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Secret    string            `json:"secret,omitempty"`
	Events    []string          `json:"events"`    // 订阅的事件类型
	Headers   map[string]string `json:"headers"`   // 自定义请求头
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	TenantID  string            `json:"tenant_id,omitempty"`
}

// WebhookDelivery Webhook 投递记录
type WebhookDelivery struct {
	ID           string          `json:"id"`
	WebhookID    string          `json:"webhook_id"`
	Event        string          `json:"event"`
	Payload      json.RawMessage `json:"payload"`
	ResponseCode int             `json:"response_code"`
	Error        string          `json:"error,omitempty"`
	Duration     int64           `json:"duration_ms"`
	DeliveredAt  time.Time       `json:"delivered_at"`
	Success      bool            `json:"success"`
}

// CreateWebhook 创建 Webhook
func (s *WebhookService) CreateWebhook(ctx context.Context, config *WebhookConfig) error {
	now := time.Now()
	config.ID = uuid.New().String()
	config.CreatedAt = now
	config.UpdatedAt = now
	config.Enabled = true

	return s.saveWebhook(ctx, config)
}

// GetWebhook 获取 Webhook
func (s *WebhookService) GetWebhook(ctx context.Context, id string) (*WebhookConfig, error) {
	data, err := s.redis.Get(ctx, "webhook:"+id)
	if err != nil {
		return nil, err
	}

	var config WebhookConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ListWebhooks 列出所有 Webhooks
func (s *WebhookService) ListWebhooks(ctx context.Context, tenantID string) ([]*WebhookConfig, error) {
	// 简化实现：扫描所有 webhook keys
	keys, err := s.redis.Keys(ctx, "webhook:*")
	if err != nil {
		return nil, err
	}

	webhooks := make([]*WebhookConfig, 0)
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var config WebhookConfig
		if err := json.Unmarshal([]byte(data), &config); err != nil {
			continue
		}

		if tenantID == "" || config.TenantID == tenantID {
			webhooks = append(webhooks, &config)
		}
	}

	return webhooks, nil
}

// UpdateWebhook 更新 Webhook
func (s *WebhookService) UpdateWebhook(ctx context.Context, id string, updates map[string]interface{}) error {
	config, err := s.GetWebhook(ctx, id)
	if err != nil {
		return err
	}

	if name, ok := updates["name"].(string); ok {
		config.Name = name
	}
	if url, ok := updates["url"].(string); ok {
		config.URL = url
	}
	if secret, ok := updates["secret"].(string); ok {
		config.Secret = secret
	}
	if events, ok := updates["events"].([]string); ok {
		config.Events = events
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		config.Enabled = enabled
	}

	config.UpdatedAt = time.Now()
	return s.saveWebhook(ctx, config)
}

// DeleteWebhook 删除 Webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, id string) error {
	return s.redis.Delete(ctx, "webhook:"+id)
}

// SaveDelivery 保存投递记录
func (s *WebhookService) SaveDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	data, err := json.Marshal(delivery)
	if err != nil {
		return err
	}

	// 保存到列表
	return s.redis.RPush(ctx, "webhook_deliveries:"+delivery.WebhookID, string(data))
}

// GetDeliveries 获取投递记录
func (s *WebhookService) GetDeliveries(ctx context.Context, webhookID string, limit int64) ([]*WebhookDelivery, error) {
	results, err := s.redis.LRange(ctx, "webhook_deliveries:"+webhookID, 0, limit-1)
	if err != nil {
		return nil, err
	}

	deliveries := make([]*WebhookDelivery, 0, len(results))
	for _, result := range results {
		var delivery WebhookDelivery
		if err := json.Unmarshal([]byte(result), &delivery); err != nil {
			continue
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

func (s *WebhookService) saveWebhook(ctx context.Context, config *WebhookConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, "webhook:"+config.ID, string(data), 0)
}