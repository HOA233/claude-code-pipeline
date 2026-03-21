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
)

// WebhookReceiver handles incoming webhooks from external services
type WebhookReceiver struct {
	secret   string
	handlers map[string]WebhookHandler
}

// WebhookHandler is a function that handles a webhook event
type WebhookHandler func(ctx context.Context, event WebhookEvent) error

// WebhookEvent represents an incoming webhook event
type WebhookEvent struct {
	Source    string                 `json:"source"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature,omitempty"`
}

// NewWebhookReceiver creates a new webhook receiver
func NewWebhookReceiver(secret string) *WebhookReceiver {
	return &WebhookReceiver{
		secret:   secret,
		handlers: make(map[string]WebhookHandler),
	}
}

// RegisterHandler registers a handler for a specific source
func (r *WebhookReceiver) RegisterHandler(source string, handler WebhookHandler) {
	r.handlers[source] = handler
}

// HandleWebhook processes an incoming webhook request
func (r *WebhookReceiver) HandleWebhook(ctx context.Context, source string, body []byte, signature string) error {
	// Verify signature if secret is configured
	if r.secret != "" {
		if !r.verifySignature(body, signature) {
			return fmt.Errorf("invalid webhook signature")
		}
	}

	// Parse the event
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to parse webhook body: %w", err)
	}

	event.Source = source
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Find and execute handler
	handler, ok := r.handlers[source]
	if !ok {
		return fmt.Errorf("no handler registered for source: %s", source)
	}

	return handler(ctx, event)
}

// verifySignature verifies the HMAC signature of the webhook
func (r *WebhookReceiver) verifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(r.secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// GitHubWebhookHandler handles GitHub webhooks
func GitHubWebhookHandler(taskSvc *TaskService, orch *Orchestrator) WebhookHandler {
	return func(ctx context.Context, event WebhookEvent) error {
		eventType, ok := event.Payload["action"].(string)
		if !ok {
			return fmt.Errorf("missing action in GitHub webhook")
		}

		switch eventType {
		case "push":
			return handleGitHubPush(ctx, event.Payload, taskSvc, orch)
		case "pull_request":
			return handleGitHubPR(ctx, event.Payload, taskSvc, orch)
		case "release":
			return handleGitHubRelease(ctx, event.Payload, taskSvc, orch)
		default:
			return nil // Ignore other events
		}
	}
}

func handleGitHubPush(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	repo, _ := payload["repository"].(map[string]interface{})
	ref, _ := payload["ref"].(string)

	if ref == "refs/heads/main" || ref == "refs/heads/master" {
		// Trigger CI/CD pipeline
		repoName, _ := repo["name"].(string)
		_, err := orch.CreatePipeline(ctx, &PipelineConfig{
			Name:        fmt.Sprintf("ci-%s-%d", repoName, time.Now().Unix()),
			Mode:        "serial",
			Description: "Triggered by GitHub push",
			Steps: []PipelineStep{
				{ID: "checkout", CLI: "git", Action: "clone"},
				{ID: "install", CLI: "npm", Action: "install"},
				{ID: "test", CLI: "npm", Action: "test"},
				{ID: "build", CLI: "npm", Action: "run", Command: "build"},
			},
		})
		return err
	}
	return nil
}

func handleGitHubPR(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	action, _ := payload["action"].(string)
	if action == "opened" || action == "synchronize" {
		// Trigger code review task
		pr, _ := payload["pull_request"].(map[string]interface{})
		prNumber, _ := pr["number"].(float64)

		_, err := taskSvc.CreateTask(ctx, &TaskConfig{
			SkillID: "code-review",
			Params: map[string]interface{}{
				"target":      fmt.Sprintf("pr-%d", int(prNumber)),
				"depth":       "standard",
				"output_file": fmt.Sprintf("review-pr-%d.json", int(prNumber)),
			},
		})
		return err
	}
	return nil
}

func handleGitHubRelease(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	action, _ := payload["action"].(string)
	if action == "published" {
		// Trigger deployment pipeline
		release, _ := payload["release"].(map[string]interface{})
		tag, _ := release["tag_name"].(string)

		_, err := orch.CreatePipeline(ctx, &PipelineConfig{
			Name:        fmt.Sprintf("deploy-%s", tag),
			Mode:        "serial",
			Description: "Triggered by GitHub release",
			Steps: []PipelineStep{
				{ID: "build-image", CLI: "docker", Action: "build"},
				{ID: "push-image", CLI: "docker", Action: "push"},
				{ID: "deploy", CLI: "kubectl", Action: "apply"},
			},
		})
		return err
	}
	return nil
}

// GitLabWebhookHandler handles GitLab webhooks
func GitLabWebhookHandler(taskSvc *TaskService, orch *Orchestrator) WebhookHandler {
	return func(ctx context.Context, event WebhookEvent) error {
		eventType, ok := event.Payload["object_kind"].(string)
		if !ok {
			return fmt.Errorf("missing object_kind in GitLab webhook")
		}

		switch eventType {
		case "push":
			return handleGitLabPush(ctx, event.Payload, taskSvc, orch)
		case "merge_request":
			return handleGitLabMR(ctx, event.Payload, taskSvc, orch)
		case "pipeline":
			return handleGitLabPipeline(ctx, event.Payload, taskSvc, orch)
		default:
			return nil
		}
	}
}

func handleGitLabPush(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	ref, _ := payload["ref"].(string)
	if ref == "refs/heads/main" || ref == "refs/heads/master" {
		project, _ := payload["project"].(map[string]interface{})
		projectName, _ := project["name"].(string)

		_, err := orch.CreatePipeline(ctx, &PipelineConfig{
			Name:        fmt.Sprintf("gitlab-ci-%s-%d", projectName, time.Now().Unix()),
			Mode:        "serial",
			Description: "Triggered by GitLab push",
			Steps: []PipelineStep{
				{ID: "sync-skills", CLI: "claude", Action: "sync"},
			},
		})
		return err
	}
	return nil
}

func handleGitLabMR(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	attrs, _ := payload["object_attributes"].(map[string]interface{})
	action, _ := attrs["action"].(string)

	if action == "open" || action == "update" {
		iid, _ := attrs["iid"].(float64)

		_, err := taskSvc.CreateTask(ctx, &TaskConfig{
			SkillID: "code-review",
			Params: map[string]interface{}{
				"target": fmt.Sprintf("mr-%d", int(iid)),
				"depth":  "standard",
			},
		})
		return err
	}
	return nil
}

func handleGitLabPipeline(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService, orch *Orchestrator) error {
	attrs, _ := payload["object_attributes"].(map[string]interface{})
	status, _ := attrs["status"].(string)

	if status == "failed" {
		// Create notification task
		_, err := taskSvc.CreateTask(ctx, &TaskConfig{
			SkillID: "notify",
			Params: map[string]interface{}{
				"type":    "pipeline_failure",
				"message": "GitLab pipeline failed",
			},
		})
		return err
	}
	return nil
}

// SlackWebhookHandler handles Slack slash commands and interactions
func SlackWebhookHandler(taskSvc *TaskService, orch *Orchestrator) WebhookHandler {
	return func(ctx context.Context, event WebhookEvent) error {
		command, _ := event.Payload["command"].(string)

		switch command {
		case "/pipeline":
			return handleSlackPipeline(ctx, event.Payload, orch)
		case "/task":
			return handleSlackTask(ctx, event.Payload, taskSvc)
		default:
			return nil
		}
	}
}

func handleSlackPipeline(ctx context.Context, payload map[string]interface{}, orch *Orchestrator) error {
	text, _ := payload["text"].(string)

	// Parse command: /pipeline run <name>
	// For simplicity, just trigger a pipeline
	if text == "run" || text == "" {
		_, err := orch.CreatePipeline(ctx, &PipelineConfig{
			Name:        fmt.Sprintf("slack-triggered-%d", time.Now().Unix()),
			Mode:        "serial",
			Description: "Triggered from Slack",
			Steps: []PipelineStep{
				{ID: "run", CLI: "claude", Action: "review"},
			},
		})
		return err
	}
	return nil
}

func handleSlackTask(ctx context.Context, payload map[string]interface{}, taskSvc *TaskService) error {
	text, _ := payload["text"].(string)

	_, err := taskSvc.CreateTask(ctx, &TaskConfig{
		SkillID: "code-review",
		Params: map[string]interface{}{
			"target": text,
		},
	})
	return err
}

// SendWebhook sends a webhook to an external URL
func SendWebhook(ctx context.Context, url string, secret string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add signature if secret is provided
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", signature)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}