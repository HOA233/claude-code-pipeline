package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/pkg/logger"
)

// WebhookService handles webhook callbacks
type WebhookService struct {
	client  *http.Client
	timeout time.Duration
}

// NewWebhookService creates a new webhook service
func NewWebhookService() *WebhookService {
	return &WebhookService{
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