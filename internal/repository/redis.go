package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisClient{client: client}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// ==================== Skill Storage ====================

const skillKeyPrefix = "skill:"

func (r *RedisClient) SaveSkill(ctx context.Context, skill *model.Skill) error {
	data, err := json.Marshal(skill)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, skillKeyPrefix+skill.ID, data, 0).Err()
}

func (r *RedisClient) GetSkill(ctx context.Context, skillID string) (*model.Skill, error) {
	data, err := r.client.Get(ctx, skillKeyPrefix+skillID).Bytes()
	if err != nil {
		return nil, err
	}
	var skill model.Skill
	if err := json.Unmarshal(data, &skill); err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *RedisClient) GetAllSkills(ctx context.Context) ([]*model.Skill, error) {
	keys, err := r.client.Keys(ctx, skillKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	skills := make([]*model.Skill, 0, len(keys))
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var skill model.Skill
		if err := json.Unmarshal(data, &skill); err != nil {
			continue
		}
		skills = append(skills, &skill)
	}
	return skills, nil
}

// ==================== Task Storage ====================

const taskKeyPrefix = "task:"

func (r *RedisClient) SaveTask(ctx context.Context, task *model.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, taskKeyPrefix+task.ID, data, 24*time.Hour).Err()
}

func (r *RedisClient) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	data, err := r.client.Get(ctx, taskKeyPrefix+taskID).Bytes()
	if err != nil {
		return nil, err
	}
	var task model.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *RedisClient) GetAllTasks(ctx context.Context) ([]*model.Task, error) {
	keys, err := r.client.Keys(ctx, taskKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.Task, 0, len(keys))
	for _, key := range keys {
		// Skip output keys
		if len(key) > len(taskKeyPrefix)+8 && key[len(taskKeyPrefix)+8] == ':' {
			continue
		}
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var task model.Task
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (r *RedisClient) AppendTaskOutput(ctx context.Context, taskID string, output string) error {
	key := taskKeyPrefix + taskID + ":output"
	return r.client.RPush(ctx, key, output).Err()
}

func (r *RedisClient) GetTaskOutput(ctx context.Context, taskID string) ([]string, error) {
	key := taskKeyPrefix + taskID + ":output"
	return r.client.LRange(ctx, key, 0, -1).Result()
}

// ==================== Task Queue ====================

const taskQueueKey = "task:queue"

func (r *RedisClient) PushTaskQueue(ctx context.Context, taskID string) error {
	return r.client.RPush(ctx, taskQueueKey, taskID).Err()
}

func (r *RedisClient) PopTaskQueue(ctx context.Context) (string, error) {
	result, err := r.client.LPop(ctx, taskQueueKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// ==================== Process Status ====================

const processKeyPrefix = "process:"

func (r *RedisClient) SaveProcess(ctx context.Context, taskID string, pid int) error {
	return r.client.Set(ctx, processKeyPrefix+taskID, pid, 0).Err()
}

func (r *RedisClient) GetProcess(ctx context.Context, taskID string) (int, error) {
	return r.client.Get(ctx, processKeyPrefix+taskID).Int()
}

func (r *RedisClient) DeleteProcess(ctx context.Context, taskID string) error {
	return r.client.Del(ctx, processKeyPrefix+taskID).Err()
}

// ==================== Pub/Sub ====================

func (r *RedisClient) PublishTaskUpdate(ctx context.Context, taskID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, "task:updates:"+taskID, jsonData).Err()
}

func (r *RedisClient) SubscribeTaskUpdates(ctx context.Context, taskID string) *redis.PubSub {
	return r.client.Subscribe(ctx, "task:updates:"+taskID)
}