package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/company/claude-pipeline/pkg/logger"
)

// NotificationService handles sending notifications for task events
type NotificationService struct {
	httpClient *http.Client
	config     NotificationConfig
}

// NotificationConfig contains notification settings
type NotificationConfig struct {
	// Slack configuration
	SlackWebhookURL string `json:"slack_webhook_url,omitempty"`
	SlackChannel    string `json:"slack_channel,omitempty"`

	// Email configuration
	SMTPHost     string `json:"smtp_host,omitempty"`
	SMTPPort     int    `json:"smtp_port,omitempty"`
	SMTPUser     string `json:"smtp_user,omitempty"`
	SMTPPassword string `json:"smtp_password,omitempty"`
	FromEmail    string `json:"from_email,omitempty"`

	// Generic webhook
	DefaultWebhookURL string `json:"default_webhook_url,omitempty"`
	WebhookSecret     string `json:"webhook_secret,omitempty"`

	// Settings
	Enabled       bool     `json:"enabled"`
	RetryCount    int      `json:"retry_count"`
	RetryDelay    int      `json:"retry_delay"` // seconds
	NotifyOn      []string `json:"notify_on"`   // events: task.completed, task.failed, etc.
}

// Notification represents a notification message
type Notification struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"` // info, warning, error, success
	Timestamp   time.Time              `json:"timestamp"`
	TaskID      string                 `json:"task_id,omitempty"`
	SkillID     string                 `json:"skill_id,omitempty"`
	PipelineID  string                 `json:"pipeline_id,omitempty"`
	Duration    int64                  `json:"duration,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(config NotificationConfig) *NotificationService {
	return &NotificationService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// Notify sends a notification through configured channels
func (n *NotificationService) Notify(ctx context.Context, notif *Notification) error {
	if !n.config.Enabled {
		return nil
	}

	// Check if this event type should trigger notification
	if !n.shouldNotify(notif.Type) {
		return nil
	}

	notif.Timestamp = time.Now()

	// Send to all configured channels
	var errors []error

	if n.config.SlackWebhookURL != "" {
		if err := n.sendSlackNotification(ctx, notif); err != nil {
			errors = append(errors, fmt.Errorf("slack: %w", err))
		}
	}

	if n.config.DefaultWebhookURL != "" {
		if err := n.sendWebhookNotification(ctx, notif); err != nil {
			errors = append(errors, fmt.Errorf("webhook: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}

	return nil
}

// shouldNotify checks if the event type should trigger notification
func (n *NotificationService) shouldNotify(eventType string) bool {
	if len(n.config.NotifyOn) == 0 {
		// Default: notify on all events
		return true
	}

	for _, e := range n.config.NotifyOn {
		if e == eventType || e == "*" {
			return true
		}
	}

	return false
}

// sendSlackNotification sends a notification to Slack
func (n *NotificationService) sendSlackNotification(ctx context.Context, notif *Notification) error {
	// Build Slack message
	color := n.getSlackColor(notif.Severity)

	payload := map[string]interface{}{
		"channel": n.config.SlackChannel,
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  notif.Title,
				"text":   notif.Message,
				"footer": fmt.Sprintf("Task: %s | Skill: %s", notif.TaskID, notif.SkillID),
				"ts":     notif.Timestamp.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Severity",
						"value": notif.Severity,
						"short": true,
					},
					{
						"title": "Type",
						"value": notif.Type,
						"short": true,
					},
				},
			},
		},
	}

	if notif.Duration > 0 {
		payload["attachments"].([]map[string]interface{})[0]["fields"] = append(
			payload["attachments"].([]map[string]interface{})[0]["fields"].([]map[string]interface{}),
			map[string]interface{}{
				"title": "Duration",
				"value": fmt.Sprintf("%dms", notif.Duration),
				"short": true,
			},
		)
	}

	return n.sendWithRetry(ctx, n.config.SlackWebhookURL, payload, n.config.RetryCount)
}

// sendWebhookNotification sends a notification to a generic webhook
func (n *NotificationService) sendWebhookNotification(ctx context.Context, notif *Notification) error {
	payload := map[string]interface{}{
		"event":     notif.Type,
		"timestamp": notif.Timestamp,
		"data":      notif,
	}

	// Add signature if secret is configured
	if n.config.WebhookSecret != "" {
		body, _ := json.Marshal(payload)
		signature := n.generateSignature(body, n.config.WebhookSecret)
		// Would need custom request handling for signature
		// For now, include in payload
		payload["signature"] = signature
	}

	return n.sendWithRetry(ctx, n.config.DefaultWebhookURL, payload, n.config.RetryCount)
}

// sendWithRetry sends HTTP POST with retry logic
func (n *NotificationService) sendWithRetry(ctx context.Context, url string, payload interface{}, maxRetries int) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	retries := maxRetries
	if retries == 0 {
		retries = 3
	}

	for i := 0; i < retries; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := n.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(n.config.RetryDelay) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.Info(fmt.Sprintf("Notification sent successfully to %s", url))
			return nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
		time.Sleep(time.Duration(n.config.RetryDelay) * time.Second)
	}

	return lastErr
}

// getSlackColor returns the Slack attachment color for a severity
func (n *NotificationService) getSlackColor(severity string) string {
	switch severity {
	case "error":
		return "#DC2626"
	case "warning":
		return "#F59E0B"
	case "success":
		return "#10B981"
	default:
		return "#3B82F6"
	}
}

// generateSignature generates an HMAC signature for webhook security
func (n *NotificationService) generateSignature(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// NotifyTaskCreated sends notification for task creation
func (n *NotificationService) NotifyTaskCreated(ctx context.Context, taskID, skillID string) error {
	return n.Notify(ctx, &Notification{
		Type:     "task.created",
		Title:    "Task Created",
		Message:  fmt.Sprintf("New task %s created with skill %s", taskID, skillID),
		Severity: "info",
		TaskID:   taskID,
		SkillID:  skillID,
	})
}

// NotifyTaskCompleted sends notification for task completion
func (n *NotificationService) NotifyTaskCompleted(ctx context.Context, taskID, skillID string, duration int64) error {
	return n.Notify(ctx, &Notification{
		Type:     "task.completed",
		Title:    "Task Completed",
		Message:  fmt.Sprintf("Task %s completed successfully", taskID),
		Severity: "success",
		TaskID:   taskID,
		SkillID:  skillID,
		Duration: duration,
	})
}

// NotifyTaskFailed sends notification for task failure
func (n *NotificationService) NotifyTaskFailed(ctx context.Context, taskID, skillID string, errMsg string) error {
	return n.Notify(ctx, &Notification{
		Type:     "task.failed",
		Title:    "Task Failed",
		Message:  fmt.Sprintf("Task %s failed: %s", taskID, errMsg),
		Severity: "error",
		TaskID:   taskID,
		SkillID:  skillID,
		Metadata: map[string]interface{}{"error": errMsg},
	})
}

// NotifyPipelineCompleted sends notification for pipeline completion
func (n *NotificationService) NotifyPipelineCompleted(ctx context.Context, pipelineID, pipelineName string, success bool, duration int64) error {
	severity := "success"
	if !success {
		severity = "error"
	}

	return n.Notify(ctx, &Notification{
		Type:       "pipeline.completed",
		Title:      fmt.Sprintf("Pipeline %s", pipelineName),
		Message:    fmt.Sprintf("Pipeline %s execution %s", pipelineName, map[bool]string{true: "succeeded", false: "failed"}[success]),
		Severity:   severity,
		PipelineID: pipelineID,
		Duration:   duration,
	})
}