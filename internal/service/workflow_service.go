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

// WorkflowService 工作流管理服务
type WorkflowService struct {
	redis      *repository.RedisClient
	agentSvc   *AgentService
	executor   *CLIExecutor
	activeRuns sync.Map
	stopChan   chan struct{}
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(redis *repository.RedisClient, agentSvc *AgentService, executor *CLIExecutor) *WorkflowService {
	return &WorkflowService{
		redis:    redis,
		agentSvc: agentSvc,
		executor: executor,
		stopChan: make(chan struct{}),
	}
}

// CreateWorkflow 创建工作流
func (s *WorkflowService) CreateWorkflow(ctx context.Context, req *model.WorkflowCreateRequest) (*model.Workflow, error) {
	// 验证 Agent 引用
	for _, agent := range req.Agents {
		if agent.AgentID != "" {
			if _, err := s.agentSvc.GetAgent(ctx, agent.AgentID); err != nil {
				return nil, fmt.Errorf("agent %s not found: %w", agent.AgentID, err)
			}
		}
	}

	now := time.Now()
	workflow := &model.Workflow{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Description:   req.Description,
		Agents:        req.Agents,
		Connections:   req.Connections,
		Mode:          req.Mode,
		Context:       req.Context,
		ErrorHandling: req.ErrorHandling,
		Output:        req.Output,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if workflow.Mode == "" {
		workflow.Mode = model.ModeSerial
	}

	if err := s.redis.SaveWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	logger.Infof("Created workflow: %s (%s)", workflow.Name, workflow.ID)
	return workflow, nil
}

// GetWorkflow 获取工作流
func (s *WorkflowService) GetWorkflow(ctx context.Context, id string) (*model.Workflow, error) {
	workflow, err := s.redis.GetWorkflow(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	return workflow, nil
}

// ListWorkflows 列出工作流
func (s *WorkflowService) ListWorkflows(ctx context.Context, tenantID string) ([]*model.Workflow, error) {
	workflows, err := s.redis.ListWorkflows(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	return workflows, nil
}

// UpdateWorkflow 更新工作流
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id string, req *model.WorkflowUpdateRequest) (*model.Workflow, error) {
	workflow, err := s.redis.GetWorkflow(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	if req.Name != "" {
		workflow.Name = req.Name
	}
	if req.Description != "" {
		workflow.Description = req.Description
	}
	if req.Agents != nil {
		workflow.Agents = req.Agents
	}
	if req.Connections != nil {
		workflow.Connections = req.Connections
	}
	if req.Mode != "" {
		workflow.Mode = req.Mode
	}
	if req.Context != nil {
		workflow.Context = req.Context
	}
	if req.ErrorHandling != nil {
		workflow.ErrorHandling = req.ErrorHandling
	}
	if req.Enabled != nil {
		workflow.Enabled = *req.Enabled
	}

	workflow.UpdatedAt = time.Now()

	if err := s.redis.SaveWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return workflow, nil
}

// DeleteWorkflow 删除工作流
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id string) error {
	if err := s.redis.DeleteWorkflow(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	logger.Infof("Deleted workflow: %s", id)
	return nil
}

// ExecuteWorkflow 执行工作流
func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, req *model.ExecutionCreateRequest) (*model.Execution, error) {
	workflow, err := s.redis.GetWorkflow(ctx, req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	if !workflow.Enabled {
		return nil, fmt.Errorf("workflow is disabled")
	}

	now := time.Now()
	execution := &model.Execution{
		ID:             uuid.New().String(),
		WorkflowID:     workflow.ID,
		WorkflowName:   workflow.Name,
		SessionID:      uuid.New().String(),
		Status:         model.ExecutionStatusPending,
		Progress:       0,
		TotalSteps:     len(workflow.Agents),
		CompletedSteps: 0,
		NodeResults:    make(map[string]model.NodeResult),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.redis.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// 异步执行
	if req.Async {
		go s.runWorkflow(ctx, workflow, execution, req.Input)
		return execution, nil
	}

	// 同步执行
	return s.runWorkflowSync(ctx, workflow, execution, req.Input)
}

// runWorkflow 异步执行工作流
func (s *WorkflowService) runWorkflow(ctx context.Context, workflow *model.Workflow, execution *model.Execution, input map[string]interface{}) {
	s.runWorkflowSync(ctx, workflow, execution, input)
}

// runWorkflowSync 同步执行工作流
func (s *WorkflowService) runWorkflowSync(ctx context.Context, workflow *model.Workflow, execution *model.Execution, input map[string]interface{}) (*model.Execution, error) {
	// 更新状态为运行中
	now := time.Now()
	execution.Status = model.ExecutionStatusRunning
	execution.StartedAt = &now
	execution.UpdatedAt = now
	s.redis.SaveExecution(ctx, execution)

	s.redis.PublishExecutionUpdate(ctx, execution.ID, map[string]interface{}{
		"execution_id": execution.ID,
		"status":       "running",
		"progress":     0,
	})

	// 根据执行模式运行
	var err error
	switch workflow.Mode {
	case model.ModeSerial:
		err = s.executeSerial(ctx, workflow, execution, input)
	case model.ModeParallel:
		err = s.executeParallel(ctx, workflow, execution, input)
	case model.ModeHybrid:
		err = s.executeHybrid(ctx, workflow, execution, input)
	}

	// 更新最终状态
	completedAt := time.Now()
	if err != nil {
		execution.Status = model.ExecutionStatusFailed
		execution.Error = err.Error()
	} else {
		execution.Status = model.ExecutionStatusCompleted
		execution.Progress = 100
	}
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(now).Milliseconds()
	execution.UpdatedAt = completedAt
	s.redis.SaveExecution(ctx, execution)

	s.redis.PublishExecutionUpdate(ctx, execution.ID, map[string]interface{}{
		"execution_id": execution.ID,
		"status":       string(execution.Status),
		"progress":     execution.Progress,
		"error":        execution.Error,
	})

	return execution, err
}

// executeSerial 串行执行
func (s *WorkflowService) executeSerial(ctx context.Context, workflow *model.Workflow, execution *model.Execution, input map[string]interface{}) error {
	sharedData := make(map[string]json.RawMessage)

	for i, agentNode := range workflow.Agents {
		// 检查依赖
		if len(agentNode.DependsOn) > 0 {
			for _, depID := range agentNode.DependsOn {
				if result, ok := execution.NodeResults[depID]; ok {
					if result.Status == model.ExecutionStatusFailed {
						return fmt.Errorf("dependency %s failed", depID)
					}
				}
			}
		}

		// 执行节点
		result, err := s.executeNode(ctx, agentNode, input, sharedData)
		if err != nil {
			execution.NodeResults[agentNode.ID] = model.NodeResult{
				NodeID: agentNode.ID,
				Status: model.ExecutionStatusFailed,
				Error:  err.Error(),
			}
			return err
		}

		execution.NodeResults[agentNode.ID] = *result
		execution.CompletedSteps = i + 1
		execution.Progress = (i + 1) * 100 / len(workflow.Agents)
		execution.CurrentStep = agentNode.Name
		execution.UpdatedAt = time.Now()
		s.redis.SaveExecution(ctx, execution)

		// 保存共享数据
		if agentNode.OutputAs != "" && result.Output != nil {
			sharedData[agentNode.OutputAs] = result.Output
		}
	}

	return nil
}

// executeParallel 并行执行
func (s *WorkflowService) executeParallel(ctx context.Context, workflow *model.Workflow, execution *model.Execution, input map[string]interface{}) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Agents))
	mu := sync.Mutex{}

	for _, node := range workflow.Agents {
		wg.Add(1)
		go func(n model.AgentNode) {
			defer wg.Done()
			result, err := s.executeNode(ctx, n, input, nil)
			mu.Lock()
			if err != nil {
				execution.NodeResults[n.ID] = model.NodeResult{
					NodeID: n.ID,
					Status: model.ExecutionStatusFailed,
					Error:  err.Error(),
				}
				errChan <- err
			} else {
				execution.NodeResults[n.ID] = *result
				execution.CompletedSteps++
				execution.Progress = execution.CompletedSteps * 100 / len(workflow.Agents)
			}
			mu.Unlock()
		}(node)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeHybrid 混合执行
func (s *WorkflowService) executeHybrid(ctx context.Context, workflow *model.Workflow, execution *model.Execution, input map[string]interface{}) error {
	// 分析依赖图，分层执行
	// TODO: 实现基于依赖图的执行
	return s.executeSerial(ctx, workflow, execution, input)
}

// executeNode 执行单个节点
func (s *WorkflowService) executeNode(ctx context.Context, node model.AgentNode, input map[string]interface{}, sharedData map[string]json.RawMessage) (*model.NodeResult, error) {
	now := time.Now()
	result := &model.NodeResult{
		NodeID:    node.ID,
		AgentID:   node.AgentID,
		Status:    model.ExecutionStatusRunning,
		StartedAt: &now,
	}

	// 合并输入
	nodeInput := make(map[string]interface{})
	for k, v := range input {
		nodeInput[k] = v
	}
	for k, v := range node.Input {
		nodeInput[k] = v
	}

	// 从共享数据获取输入
	for key, source := range node.InputFrom {
		if data, ok := sharedData[source]; ok {
			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				nodeInput[key] = value
			}
		}
	}

	// 执行
	var output json.RawMessage
	var err error

	if node.AgentID != "" {
		// 使用 Agent 执行
		agent, agentErr := s.agentSvc.GetAgent(ctx, node.AgentID)
		if agentErr != nil {
			result.Status = model.ExecutionStatusFailed
			result.Error = agentErr.Error()
			return result, agentErr
		}

		// 调用 CLI 执行
		output, err = s.executor.ExecuteCLI(ctx, "claude", node.Action, node.Command, nodeInput)
	} else if node.CLI != "" {
		// 直接 CLI 执行
		output, err = s.executor.ExecuteCLI(ctx, node.CLI, node.Action, node.Command, nodeInput)
	}

	completedAt := time.Now()
	result.CompletedAt = &completedAt
	result.Duration = completedAt.Sub(now).Milliseconds()

	if err != nil {
		result.Status = model.ExecutionStatusFailed
		result.Error = err.Error()
		return result, err
	}

	result.Status = model.ExecutionStatusCompleted
	result.Output = output
	return result, nil
}

// CancelExecution 取消执行
func (s *WorkflowService) CancelExecution(ctx context.Context, executionID string) error {
	execution, err := s.redis.GetExecution(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	if execution.Status != model.ExecutionStatusRunning && execution.Status != model.ExecutionStatusPending {
		return fmt.Errorf("cannot cancel execution in status: %s", execution.Status)
	}

	now := time.Now()
	execution.Status = model.ExecutionStatusCancelled
	execution.CompletedAt = &now
	execution.UpdatedAt = now

	s.redis.SaveExecution(ctx, execution)
	s.redis.PublishExecutionUpdate(ctx, executionID, map[string]interface{}{
		"execution_id": executionID,
		"status":       "cancelled",
	})

	return nil
}

// PauseExecution 暂停执行
func (s *WorkflowService) PauseExecution(ctx context.Context, executionID string) error {
	execution, err := s.redis.GetExecution(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	if execution.Status != model.ExecutionStatusRunning {
		return fmt.Errorf("can only pause running execution")
	}

	execution.Status = model.ExecutionStatusPaused
	execution.UpdatedAt = time.Now()
	s.redis.SaveExecution(ctx, execution)

	return nil
}

// ResumeExecution 恢复执行
func (s *WorkflowService) ResumeExecution(ctx context.Context, executionID string) error {
	execution, err := s.redis.GetExecution(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	if execution.Status != model.ExecutionStatusPaused {
		return fmt.Errorf("can only resume paused execution")
	}

	execution.Status = model.ExecutionStatusRunning
	execution.UpdatedAt = time.Now()
	s.redis.SaveExecution(ctx, execution)

	return nil
}

// GetExecution 获取执行
func (s *WorkflowService) GetExecution(ctx context.Context, id string) (*model.Execution, error) {
	return s.redis.GetExecution(ctx, id)
}

// ListExecutions 列出执行
func (s *WorkflowService) ListExecutions(ctx context.Context, filter *model.ExecutionFilter) (*model.ExecutionListResponse, error) {
	return s.redis.ListExecutions(ctx, filter)
}

// CancelAllExecutions 取消所有执行
func (s *WorkflowService) CancelAllExecutions(ctx context.Context, status model.ExecutionStatus) (int, error) {
	filter := &model.ExecutionFilter{Status: status}
	list, err := s.redis.ListExecutions(ctx, filter)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, exec := range list.Executions {
		if err := s.CancelExecution(ctx, exec.ID); err == nil {
			count++
		}
	}

	return count, nil
}