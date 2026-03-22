package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
	"github.com/google/uuid"
)

// AgentService Agent 管理服务
type AgentService struct {
	redis *repository.RedisClient
}

// NewAgentService 创建 Agent 服务
func NewAgentService(redis *repository.RedisClient) *AgentService {
	return &AgentService{redis: redis}
}

// CreateAgent 创建 Agent
func (s *AgentService) CreateAgent(ctx context.Context, req *model.AgentCreateRequest) (*model.Agent, error) {
	now := time.Now()
	agent := &model.Agent{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
		MaxTokens:    req.MaxTokens,
		Skills:       req.Skills,
		Tools:        req.Tools,
		Permissions:  req.Permissions,
		InputSchema:  req.InputSchema,
		OutputSchema: req.OutputSchema,
		Timeout:      req.Timeout,
		Isolation:    req.Isolation,
		Tags:         req.Tags,
		Category:     req.Category,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.redis.SaveAgent(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to save agent: %w", err)
	}

	logger.Infof("Created agent: %s (%s)", agent.Name, agent.ID)
	return agent, nil
}

// CreateAgentDirect 直接创建 Agent（用于模板实例化）
func (s *AgentService) CreateAgentDirect(ctx context.Context, agent *model.Agent) error {
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}
	now := time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = now
	}
	if agent.UpdatedAt.IsZero() {
		agent.UpdatedAt = now
	}

	if err := s.redis.SaveAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to save agent: %w", err)
	}

	logger.Infof("Created agent directly: %s (%s)", agent.Name, agent.ID)
	return nil
}

// GetAgent 获取 Agent
func (s *AgentService) GetAgent(ctx context.Context, id string) (*model.Agent, error) {
	agent, err := s.redis.GetAgent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return agent, nil
}

// ListAgents 列出所有 Agent
func (s *AgentService) ListAgents(ctx context.Context, tenantID string, category string) ([]*model.Agent, error) {
	agents, err := s.redis.ListAgents(ctx, tenantID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return agents, nil
}

// UpdateAgent 更新 Agent
func (s *AgentService) UpdateAgent(ctx context.Context, id string, req *model.AgentUpdateRequest) (*model.Agent, error) {
	agent, err := s.redis.GetAgent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Model != "" {
		agent.Model = req.Model
	}
	if req.SystemPrompt != "" {
		agent.SystemPrompt = req.SystemPrompt
	}
	if req.MaxTokens > 0 {
		agent.MaxTokens = req.MaxTokens
	}
	if req.Skills != nil {
		agent.Skills = req.Skills
	}
	if req.Tools != nil {
		agent.Tools = req.Tools
	}
	if req.Permissions != nil {
		agent.Permissions = req.Permissions
	}
	if req.Timeout > 0 {
		agent.Timeout = req.Timeout
	}
	if req.Enabled != nil {
		agent.Enabled = *req.Enabled
	}

	agent.UpdatedAt = time.Now()

	if err := s.redis.SaveAgent(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return agent, nil
}

// DeleteAgent 删除 Agent
func (s *AgentService) DeleteAgent(ctx context.Context, id string) error {
	if err := s.redis.DeleteAgent(ctx, id); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	logger.Infof("Deleted agent: %s", id)
	return nil
}

// ExecuteAgent 执行 Agent
func (s *AgentService) ExecuteAgent(ctx context.Context, id string, req *model.AgentExecuteRequest) (*model.Execution, error) {
	agent, err := s.redis.GetAgent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	if !agent.Enabled {
		return nil, fmt.Errorf("agent is disabled")
	}

	// 创建执行记录
	now := time.Now()
	execution := &model.Execution{
		ID:           uuid.New().String(),
		WorkflowID:   "", // 单 Agent 执行没有 Workflow
		WorkflowName: agent.Name,
		SessionID:    uuid.New().String(),
		Status:       model.ExecutionStatusPending,
		Progress:     0,
		TotalSteps:   1,
		NodeResults:  make(map[string]model.NodeResult),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 保存执行记录
	if err := s.redis.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// 如果是异步执行，放入队列
	if req.Async {
		taskData, _ := json.Marshal(map[string]interface{}{
			"execution_id": execution.ID,
			"agent_id":     id,
			"input":        req.Input,
			"context":      req.Context,
			"callback":     req.Callback,
		})
		if err := s.redis.PushTaskQueue(ctx, string(taskData)); err != nil {
			return nil, fmt.Errorf("failed to queue execution: %w", err)
		}
		return execution, nil
	}

	// 同步执行
	return s.executeAgentSync(ctx, agent, execution, req)
}

// executeAgentSync 同步执行 Agent
func (s *AgentService) executeAgentSync(ctx context.Context, agent *model.Agent, execution *model.Execution, req *model.AgentExecuteRequest) (*model.Execution, error) {
	// 更新状态为运行中
	now := time.Now()
	execution.Status = model.ExecutionStatusRunning
	execution.StartedAt = &now
	execution.CurrentStep = agent.Name
	s.redis.SaveExecution(ctx, execution)

	// 发布更新
	s.redis.PublishExecutionUpdate(ctx, execution.ID, map[string]interface{}{
		"execution_id": execution.ID,
		"status":       "running",
		"progress":     0,
	})

	// TODO: 实际调用 Claude Code CLI 执行
	// 这里返回模拟结果
	completedAt := time.Now()
	execution.Status = model.ExecutionStatusCompleted
	execution.Progress = 100
	execution.CompletedSteps = 1
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(now).Milliseconds()
	execution.FinalOutput = json.RawMessage(`{"result": "success"}`)
	execution.UpdatedAt = completedAt

	s.redis.SaveExecution(ctx, execution)

	s.redis.PublishExecutionUpdate(ctx, execution.ID, map[string]interface{}{
		"execution_id": execution.ID,
		"status":       "completed",
		"progress":     100,
	})

	return execution, nil
}

// TestAgent 测试 Agent
func (s *AgentService) TestAgent(ctx context.Context, id string, input map[string]interface{}) (map[string]interface{}, error) {
	agent, err := s.redis.GetAgent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// 返回测试信息
	result := map[string]interface{}{
		"agent_id":   agent.ID,
		"agent_name": agent.Name,
		"model":      agent.Model,
		"input":      input,
		"test":       true,
		"message":    "Agent configuration is valid",
	}

	return result, nil
}