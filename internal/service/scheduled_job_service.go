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

// ScheduledJobService 定时任务服务
type ScheduledJobService struct {
	redis        *repository.RedisClient
	agentSvc     *AgentService
	workflowSvc  *WorkflowService
	schedulers   sync.Map
	stopChan     chan struct{}
	mu           sync.RWMutex
}

// NewScheduledJobService 创建定时任务服务
func NewScheduledJobService(redis *repository.RedisClient, agentSvc *AgentService, workflowSvc *WorkflowService) *ScheduledJobService {
	return &ScheduledJobService{
		redis:       redis,
		agentSvc:    agentSvc,
		workflowSvc: workflowSvc,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动调度器
func (s *ScheduledJobService) Start(ctx context.Context) {
	logger.Info("Scheduled job service started")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Scheduled job service stopped")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunJobs(ctx)
		}
	}
}

// Stop 停止调度器
func (s *ScheduledJobService) Stop() {
	close(s.stopChan)
}

// CreateJob 创建定时任务
func (s *ScheduledJobService) CreateJob(ctx context.Context, req *model.ScheduledJobCreateRequest) (*model.ScheduledJob, error) {
	// 验证目标存在
	if req.TargetType == "agent" {
		if _, err := s.agentSvc.GetAgent(ctx, req.TargetID); err != nil {
			return nil, fmt.Errorf("agent not found: %w", err)
		}
	} else if req.TargetType == "workflow" {
		if _, err := s.workflowSvc.GetWorkflow(ctx, req.TargetID); err != nil {
			return nil, fmt.Errorf("workflow not found: %w", err)
		}
	} else {
		return nil, fmt.Errorf("invalid target_type: %s", req.TargetType)
	}

	// 验证 Cron 表达式
	if _, err := parseCron(req.Cron); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	now := time.Now()
	job := &model.ScheduledJob{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		Cron:        req.Cron,
		Timezone:    req.Timezone,
		Input:       req.Input,
		OnFailure:   req.OnFailure,
		NotifyEmail: req.NotifyEmail,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 计算下次执行时间
	nextRun := calculateNextRun(job.Cron, job.Timezone, now)
	job.NextRun = &nextRun

	if err := s.redis.SaveScheduledJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save job: %w", err)
	}

	logger.Infof("Created scheduled job: %s (%s)", job.Name, job.ID)
	return job, nil
}

// GetJob 获取定时任务
func (s *ScheduledJobService) GetJob(ctx context.Context, id string) (*model.ScheduledJob, error) {
	return s.redis.GetScheduledJob(ctx, id)
}

// ListJobs 列出定时任务
func (s *ScheduledJobService) ListJobs(ctx context.Context, tenantID string) ([]*model.ScheduledJob, error) {
	return s.redis.ListScheduledJobs(ctx, tenantID)
}

// UpdateJob 更新定时任务
func (s *ScheduledJobService) UpdateJob(ctx context.Context, id string, req *model.ScheduledJobUpdateRequest) (*model.ScheduledJob, error) {
	job, err := s.redis.GetScheduledJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	if req.Name != "" {
		job.Name = req.Name
	}
	if req.Description != "" {
		job.Description = req.Description
	}
	if req.Cron != "" {
		if _, err := parseCron(req.Cron); err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		job.Cron = req.Cron
		nextRun := calculateNextRun(job.Cron, job.Timezone, time.Now())
		job.NextRun = &nextRun
	}
	if req.Timezone != "" {
		job.Timezone = req.Timezone
	}
	if req.Input != nil {
		job.Input = req.Input
	}
	if req.OnFailure != "" {
		job.OnFailure = req.OnFailure
	}
	if req.NotifyEmail != "" {
		job.NotifyEmail = req.NotifyEmail
	}
	if req.Enabled != nil {
		job.Enabled = *req.Enabled
	}

	job.UpdatedAt = time.Now()

	if err := s.redis.SaveScheduledJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	return job, nil
}

// DeleteJob 删除定时任务
func (s *ScheduledJobService) DeleteJob(ctx context.Context, id string) error {
	return s.redis.DeleteScheduledJob(ctx, id)
}

// EnableJob 启用定时任务
func (s *ScheduledJobService) EnableJob(ctx context.Context, id string) error {
	enabled := true
	_, err := s.UpdateJob(ctx, id, &model.ScheduledJobUpdateRequest{Enabled: &enabled})
	return err
}

// DisableJob 禁用定时任务
func (s *ScheduledJobService) DisableJob(ctx context.Context, id string) error {
	enabled := false
	_, err := s.UpdateJob(ctx, id, &model.ScheduledJobUpdateRequest{Enabled: &enabled})
	return err
}

// TriggerJob 手动触发任务
func (s *ScheduledJobService) TriggerJob(ctx context.Context, id string) (*model.Execution, error) {
	job, err := s.redis.GetScheduledJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	return s.runJob(ctx, job)
}

// GetJobHistory 获取执行历史
func (s *ScheduledJobService) GetJobHistory(ctx context.Context, jobID string, page, pageSize int) (*model.JobExecutionHistoryListResponse, error) {
	return s.redis.GetJobExecutionHistory(ctx, jobID, page, pageSize)
}

// checkAndRunJobs 检查并运行到期的任务
func (s *ScheduledJobService) checkAndRunJobs(ctx context.Context) {
	jobs, err := s.redis.GetEnabledScheduledJobs(ctx)
	if err != nil {
		logger.Error("Failed to get enabled jobs: ", err)
		return
	}

	now := time.Now()
	for _, job := range jobs {
		if job.NextRun != nil && (job.NextRun.Before(now) || job.NextRun.Equal(now)) {
			go s.runJob(ctx, job)
		}
	}
}

// runJob 执行任务
func (s *ScheduledJobService) runJob(ctx context.Context, job *model.ScheduledJob) (*model.Execution, error) {
	logger.Infof("Running scheduled job: %s", job.Name)

	var execution *model.Execution
	var err error

	// 根据目标类型执行
	if job.TargetType == "agent" {
		execution, err = s.agentSvc.ExecuteAgent(ctx, job.TargetID, &model.AgentExecuteRequest{
			Input: job.Input,
			Async: false,
		})
	} else if job.TargetType == "workflow" {
		execution, err = s.workflowSvc.ExecuteWorkflow(ctx, &model.ExecutionCreateRequest{
			WorkflowID: job.TargetID,
			Input:      job.Input,
			Async:      false,
		})
	}

	// 记录执行历史
	now := time.Now()
	history := &model.JobExecutionHistory{
		ID:      uuid.New().String(),
		JobID:   job.ID,
		StartedAt: now,
	}

	if err != nil {
		history.Status = "failed"
		history.Error = err.Error()
		logger.Errorf("Failed to run job %s: %v", job.Name, err)

		// 失败处理
		s.handleFailure(ctx, job, err)
	} else {
		history.Status = string(execution.Status)
		history.ExecutionID = execution.ID
		history.Duration = execution.Duration
		if execution.CompletedAt != nil {
			history.CompletedAt = execution.CompletedAt
		}
	}

	// 保存历史
	s.redis.SaveJobExecutionHistory(ctx, history)

	// 更新任务状态
	job.LastRun = &now
	job.LastStatus = history.Status
	if err != nil {
		job.LastError = err.Error()
	} else {
		job.LastError = ""
	}
	job.RunCount++

	// 计算下次执行时间
	nextRun := calculateNextRun(job.Cron, job.Timezone, now)
	job.NextRun = &nextRun

	s.redis.SaveScheduledJob(ctx, job)

	return execution, err
}

// handleFailure 处理失败
func (s *ScheduledJobService) handleFailure(ctx context.Context, job *model.ScheduledJob, err error) {
	switch job.OnFailure {
	case "disable":
		job.Enabled = false
		s.redis.SaveScheduledJob(ctx, job)
		logger.Infof("Disabled job %s due to failure", job.Name)
	case "notify":
		// TODO: 发送通知
		logger.Infof("Should notify %s about job %s failure", job.NotifyEmail, job.Name)
	case "retry":
		// 重试逻辑在下次调度周期自动处理
	}
}

// parseCron 解析 Cron 表达式
func parseCron(expr string) (string, error) {
	// 简单验证，支持标准 5 字段 cron
	// 生产环境应使用 cron 库
	if expr == "" {
		return "", fmt.Errorf("empty cron expression")
	}
	return expr, nil
}

// calculateNextRun 计算下次执行时间
func calculateNextRun(cronExpr, timezone string, from time.Time) time.Time {
	// 简化实现：支持常用模式
	// 生产环境应使用完整 cron 解析库

	loc := time.Local
	if timezone != "" {
		if l, err := time.LoadLocation(timezone); err == nil {
			loc = l
		}
	}

	// 解析简单 cron 格式 "M H DoM Mon DoW"
	// 这里简化处理，支持基本模式
	switch cronExpr {
	case "@hourly":
		return from.Add(1 * time.Hour).Truncate(time.Hour)
	case "@daily":
		return time.Date(from.Year(), from.Month(), from.Day()+1, 0, 0, 0, 0, loc)
	case "@weekly":
		daysUntilSunday := 7 - int(from.Weekday())
		return time.Date(from.Year(), from.Month(), from.Day()+daysUntilSunday, 0, 0, 0, 0, loc)
	case "@monthly":
		return time.Date(from.Year(), from.Month()+1, 1, 0, 0, 0, 0, loc)
	default:
		// 解析标准 cron "*/10 * * * *"
		// 简化：假设每 N 分钟
		return from.Add(1 * time.Hour)
	}
}