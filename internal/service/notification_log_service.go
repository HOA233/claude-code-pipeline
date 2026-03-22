package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
	"github.com/google/uuid"
)

// ExecutionLogService 执行日志服务
type ExecutionLogService struct {
	redis    *repository.RedisClient
	loggers  sync.Map
}

// NewExecutionLogService 创建日志服务
func NewExecutionLogService(redis *repository.RedisClient) *ExecutionLogService {
	return &ExecutionLogService{redis: redis}
}

// LogEntry 日志条目
type LogEntry struct {
	ID          string                 `json:"id"`
	ExecutionID string                 `json:"execution_id"`
	NodeID      string                 `json:"node_id,omitempty"`
	Level       string                 `json:"level"` // info, warn, error, debug
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WriteLog 写入日志
func (s *ExecutionLogService) WriteLog(ctx context.Context, executionID, nodeID, level, message string, metadata map[string]interface{}) error {
	entry := LogEntry{
		ID:          uuid.New().String(),
		ExecutionID: executionID,
		NodeID:      nodeID,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
		Metadata:    metadata,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// 保存到 Redis 列表
	return s.redis.RPush(ctx, "execution_logs:"+executionID, string(data))
}

// GetLogs 获取日志
func (s *ExecutionLogService) GetLogs(ctx context.Context, executionID string, limit int64) ([]LogEntry, error) {
	key := "execution_logs:" + executionID
	results, err := s.redis.LRange(ctx, key, 0, limit-1)
	if err != nil {
		return nil, err
	}

	logs := make([]LogEntry, 0, len(results))
	for _, result := range results {
		var entry LogEntry
		if err := json.Unmarshal([]byte(result), &entry); err != nil {
			continue
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

// StreamLogs 流式获取日志
func (s *ExecutionLogService) StreamLogs(ctx context.Context, executionID string) <-chan LogEntry {
	ch := make(chan LogEntry, 100)

	go func() {
		defer close(ch)
		key := "execution_logs:" + executionID
		index := int64(0)

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				results, err := s.redis.LRange(ctx, key, index, -1)
				if err != nil {
					continue
				}

				for _, result := range results {
					var entry LogEntry
					if err := json.Unmarshal([]byte(result), &entry); err != nil {
						continue
					}
					ch <- entry
				}

				index += int64(len(results))
			}
		}
	}()

	return ch
}

// Info 记录信息日志
func (s *ExecutionLogService) Info(ctx context.Context, executionID, nodeID, message string, metadata ...map[string]interface{}) {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	s.WriteLog(ctx, executionID, nodeID, "info", message, meta)
}

// Warn 记录警告日志
func (s *ExecutionLogService) Warn(ctx context.Context, executionID, nodeID, message string, metadata ...map[string]interface{}) {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	s.WriteLog(ctx, executionID, nodeID, "warn", message, meta)
}

// Error 记录错误日志
func (s *ExecutionLogService) Error(ctx context.Context, executionID, nodeID, message string, metadata ...map[string]interface{}) {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	s.WriteLog(ctx, executionID, nodeID, "error", message, meta)
}

// Debug 记录调试日志
func (s *ExecutionLogService) Debug(ctx context.Context, executionID, nodeID, message string, metadata ...map[string]interface{}) {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	s.WriteLog(ctx, executionID, nodeID, "debug", message, meta)
}

// WebhookNotificationService Webhook 通知服务
type WebhookNotificationService struct {
	redis   *repository.RedisClient
	history sync.Map
}

// NewWebhookNotificationService 创建 Webhook 通知服务
func NewWebhookNotificationService(redis *repository.RedisClient) *WebhookNotificationService {
	return &WebhookNotificationService{redis: redis}
}

// WebhookEvent Webhook 事件
type WebhookEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // execution.started, execution.completed, execution.failed, etc.
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Secret  string            `json:"secret,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Events  []string          `json:"events"` // 订阅的事件类型
}

// SendWebhook 发送 Webhook
func (s *WebhookNotificationService) SendWebhook(ctx context.Context, url string, event *WebhookEvent) error {
	// 这里简化实现，实际应该使用 HTTP 客户端发送
	logger.Infof("Sending webhook to %s: %s", url, event.Type)

	// 记录历史
	s.history.Store(event.ID, event)

	return nil
}

// NotifyExecutionUpdate 通知执行更新
func (s *WebhookNotificationService) NotifyExecutionUpdate(ctx context.Context, execution *model.Execution, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}

	event := &WebhookEvent{
		ID:        uuid.New().String(),
		Type:      fmt.Sprintf("execution.%s", execution.Status),
		Timestamp: time.Now(),
		Data:      execution,
	}

	return s.SendWebhook(ctx, webhookURL, event)
}

// NotifyScheduledJobTrigger 通知定时任务触发
func (s *WebhookNotificationService) NotifyScheduledJobTrigger(ctx context.Context, job *model.ScheduledJob, execution *model.Execution, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}

	event := &WebhookEvent{
		ID:        uuid.New().String(),
		Type:      "scheduled_job.triggered",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"job":       job,
			"execution": execution,
		},
	}

	return s.SendWebhook(ctx, webhookURL, event)
}

// NotifyError 通知错误
func (s *WebhookNotificationService) NotifyError(ctx context.Context, executionID, errorMessage, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}

	event := &WebhookEvent{
		ID:        uuid.New().String(),
		Type:      "execution.error",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id": executionID,
			"error":        errorMessage,
		},
	}

	return s.SendWebhook(ctx, webhookURL, event)
}

// NotificationService 统一通知服务
type NotificationService struct {
	webhook *WebhookNotificationService
	log     *ExecutionLogService
}

// NewNotificationService 创建通知服务
func NewNotificationService(redis *repository.RedisClient) *NotificationService {
	return &NotificationService{
		webhook: NewWebhookNotificationService(redis),
		log:     NewExecutionLogService(redis),
	}
}

// OnExecutionStart 执行开始通知
func (s *NotificationService) OnExecutionStart(ctx context.Context, execution *model.Execution, webhookURL string) {
	s.log.Info(ctx, execution.ID, "", "Execution started", map[string]interface{}{
		"workflow_id":   execution.WorkflowID,
		"workflow_name": execution.WorkflowName,
	})

	s.webhook.NotifyExecutionUpdate(ctx, execution, webhookURL)
}

// OnExecutionComplete 执行完成通知
func (s *NotificationService) OnExecutionComplete(ctx context.Context, execution *model.Execution, webhookURL string) {
	s.log.Info(ctx, execution.ID, "", "Execution completed", map[string]interface{}{
		"status":   execution.Status,
		"duration": execution.Duration,
	})

	s.webhook.NotifyExecutionUpdate(ctx, execution, webhookURL)
}

// OnExecutionFailed 执行失败通知
func (s *NotificationService) OnExecutionFailed(ctx context.Context, execution *model.Execution, webhookURL string) {
	s.log.Error(ctx, execution.ID, "", "Execution failed", map[string]interface{}{
		"error": execution.Error,
	})

	s.webhook.NotifyError(ctx, execution.ID, execution.Error, webhookURL)
}

// OnNodeStart 节点开始通知
func (s *NotificationService) OnNodeStart(ctx context.Context, executionID, nodeID, agentID string) {
	s.log.Info(ctx, executionID, nodeID, "Node started", map[string]interface{}{
		"agent_id": agentID,
	})
}

// OnNodeComplete 节点完成通知
func (s *NotificationService) OnNodeComplete(ctx context.Context, executionID, nodeID string, result *model.NodeResult) {
	s.log.Info(ctx, executionID, nodeID, "Node completed", map[string]interface{}{
		"status":   result.Status,
		"duration": result.Duration,
	})
}

// OnNodeFailed 节点失败通知
func (s *NotificationService) OnNodeFailed(ctx context.Context, executionID, nodeID, errorMessage string) {
	s.log.Error(ctx, executionID, nodeID, "Node failed", map[string]interface{}{
		"error": errorMessage,
	})
}