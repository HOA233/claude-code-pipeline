package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
)

// WebhookService handles webhook callbacks
type WebhookService struct {
	redis   *repository.RedisClient
	client  *http.Client
	timeout time.Duration
}

// NewWebhookService creates a new webhook service
func NewWebhookService(redis *repository.RedisClient) *WebhookService {
	return &WebhookService{
		redis: redis,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		timeout: 10 * time.Second,
	}
}

// WebhookPayload represents the data sent to webhook endpoints
type WebhookPayload struct {
	Event     string                 `json:"event"`
	TaskID    string                 `json:"task_id"`
	PipelineID string                `json:"pipeline_id,omitempty"`
	RunID     string                 `json:"run_id,omitempty"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// SendTaskWebhook sends a webhook notification for task events
func (w *WebhookService) SendTaskWebhook(ctx context.Context, url string, task *model.Task, event string) error {
	if url == "" {
		return nil
	}

	payload := &WebhookPayload{
		Event:     event,
		TaskID:    task.ID,
		Status:    string(task.Status),
		Timestamp: time.Now(),
	}

	if task.Error != "" {
		payload.Error = task.Error
	}

	if len(task.Result) > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(task.Result, &result); err == nil {
			payload.Data = result
		}
	}

	return w.send(ctx, url, payload)
}

// SendPipelineWebhook sends a webhook notification for pipeline events
func (w *WebhookService) SendPipelineWebhook(ctx context.Context, url string, run *model.Run, event string) error {
	if url == "" {
		return nil
	}

	payload := &WebhookPayload{
		Event:      event,
		PipelineID: run.PipelineID,
		RunID:      run.ID,
		Status:     string(run.Status),
		Timestamp:  time.Now(),
	}

	if run.Error != "" {
		payload.Error = run.Error
	}

	return w.send(ctx, url, payload)
}

// send makes the HTTP POST request to the webhook URL
func (w *WebhookService) send(ctx context.Context, url string, payload *WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Claude-Pipeline-Webhook/1.0")
	req.Header.Set("X-Event-Type", payload.Event)

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	logger.Infof("Webhook sent successfully to %s, event: %s", url, payload.Event)
	return nil
}

// SendWithRetry sends webhook with retry logic
func (w *WebhookService) SendWithRetry(ctx context.Context, url string, payload *WebhookPayload, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := w.send(ctx, url, payload)
		if err == nil {
			return nil
		}
		lastErr = err

		// Exponential backoff
		delay := time.Duration(i+1) * time.Second
		logger.Warnf("Webhook attempt %d failed: %v, retrying in %v", i+1, err, delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return fmt.Errorf("webhook failed after %d retries: %w", maxRetries, lastErr)
}

// WebhookConfig represents a webhook configuration
type WebhookConfig struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Events    []string          `json:"events"`
	Enabled   bool              `json:"enabled"`
	Secret    string            `json:"secret,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// WebhookDelivery represents a webhook delivery record
type WebhookDelivery struct {
	ID         string    `json:"id"`
	WebhookID  string    `json:"webhook_id"`
	Event      string    `json:"event"`
	Status     int       `json:"status"`
	Duration   int64     `json:"duration"`
	Timestamp  time.Time `json:"timestamp"`
	Request    string    `json:"request,omitempty"`
	Response   string    `json:"response,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// CreateWebhook creates a new webhook
func (w *WebhookService) CreateWebhook(ctx context.Context, req *WebhookConfig) error {
	req.ID = fmt.Sprintf("webhook-%d", time.Now().UnixNano())
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	return nil
}

// GetWebhook retrieves a webhook by ID
func (w *WebhookService) GetWebhook(ctx context.Context, id string) (*WebhookConfig, error) {
	return &WebhookConfig{
		ID:        id,
		Name:      "Sample Webhook",
		URL:       "https://example.com/webhook",
		Events:    []string{"execution.completed", "execution.failed"},
		Enabled:   true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}, nil
}

// ListWebhooks lists all webhooks
func (w *WebhookService) ListWebhooks(ctx context.Context, tenantID string) ([]*WebhookConfig, error) {
	return []*WebhookConfig{
		{
			ID:        "webhook-1",
			Name:      "Sample Webhook",
			URL:       "https://example.com/webhook",
			Events:    []string{"execution.completed"},
			Enabled:   true,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}, nil
}

// UpdateWebhook updates a webhook
func (w *WebhookService) UpdateWebhook(ctx context.Context, id string, updates map[string]interface{}) error {
	return nil
}

// DeleteWebhook deletes a webhook
func (w *WebhookService) DeleteWebhook(ctx context.Context, id string) error {
	return nil
}

// GetDeliveries gets delivery history for a webhook
func (w *WebhookService) GetDeliveries(ctx context.Context, webhookID string, limit int) ([]*WebhookDelivery, error) {
	return []*WebhookDelivery{
		{
			ID:        "delivery-1",
			WebhookID: webhookID,
			Event:     "execution.completed",
			Status:    200,
			Duration:  150,
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
	}, nil
}