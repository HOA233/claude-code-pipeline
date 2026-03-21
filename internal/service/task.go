package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	redis *repository.RedisClient
}

func NewTaskService(redis *repository.RedisClient) *TaskService {
	return &TaskService{redis: redis}
}

func (s *TaskService) CreateTask(ctx context.Context, req *model.TaskCreateRequest) (*model.Task, error) {
	// Verify skill exists
	skill, err := s.redis.GetSkill(ctx, req.SkillID)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}

	// Validate parameters
	if err := s.validateParameters(skill, req.Parameters); err != nil {
		return nil, err
	}

	// Create task
	now := time.Now()
	paramsJSON, _ := json.Marshal(req.Parameters)
	contextJSON, _ := json.Marshal(req.Context)

	task := &model.Task{
		ID:         "task-" + uuid.New().String()[:8],
		SkillID:    req.SkillID,
		Status:     model.TaskStatusPending,
		Parameters: paramsJSON,
		Context:    contextJSON,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Save task
	if err := s.redis.SaveTask(ctx, task); err != nil {
		return nil, err
	}

	// Add to queue
	if err := s.redis.PushTaskQueue(ctx, task.ID); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	return s.redis.GetTask(ctx, taskID)
}

func (s *TaskService) GetAllTasks(ctx context.Context) ([]*model.Task, error) {
	return s.redis.GetAllTasks(ctx)
}

func (s *TaskService) GetTaskResult(ctx context.Context, taskID string) (map[string]interface{}, error) {
	task, err := s.redis.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"task_id":      task.ID,
		"status":       task.Status,
		"skill_id":     task.SkillID,
		"duration":     task.Duration,
		"completed_at": task.CompletedAt,
	}

	if len(task.Result) > 0 {
		var taskResult interface{}
		json.Unmarshal(task.Result, &taskResult)
		result["result"] = taskResult
	}

	if task.Error != "" {
		result["error"] = task.Error
	}

	return result, nil
}

func (s *TaskService) GetTaskOutput(ctx context.Context, taskID string) ([]string, error) {
	return s.redis.GetTaskOutput(ctx, taskID)
}

func (s *TaskService) validateParameters(skill *model.Skill, params map[string]interface{}) error {
	for _, param := range skill.Parameters {
		value, exists := params[param.Name]

		if param.Required && !exists {
			return fmt.Errorf("missing required parameter: %s", param.Name)
		}

		if param.Type == "enum" && len(param.Values) > 0 && exists {
			strValue := fmt.Sprintf("%v", value)
			valid := false
			for _, v := range param.Values {
				if v == strValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid value for parameter %s: %v (allowed: %v)", param.Name, value, param.Values)
			}
		}
	}
	return nil
}

func (s *TaskService) SubscribeTaskUpdates(ctx context.Context, taskID string) *redis.PubSub {
	return s.redis.SubscribeTaskUpdates(ctx, taskID)
}