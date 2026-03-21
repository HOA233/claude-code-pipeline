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

// ==================== Pipeline Storage ====================

const pipelineKeyPrefix = "pipeline:"

func (r *RedisClient) SavePipeline(ctx context.Context, pipeline *model.Pipeline) error {
	data, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, pipelineKeyPrefix+pipeline.ID, data, 0).Err()
}

func (r *RedisClient) GetPipeline(ctx context.Context, pipelineID string) (*model.Pipeline, error) {
	data, err := r.client.Get(ctx, pipelineKeyPrefix+pipelineID).Bytes()
	if err != nil {
		return nil, err
	}
	var pipeline model.Pipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (r *RedisClient) GetAllPipelines(ctx context.Context) ([]*model.Pipeline, error) {
	keys, err := r.client.Keys(ctx, pipelineKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	pipelines := make([]*model.Pipeline, 0, len(keys))
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var pipeline model.Pipeline
		if err := json.Unmarshal(data, &pipeline); err != nil {
			continue
		}
		pipelines = append(pipelines, &pipeline)
	}
	return pipelines, nil
}

func (r *RedisClient) DeletePipeline(ctx context.Context, pipelineID string) error {
	return r.client.Del(ctx, pipelineKeyPrefix+pipelineID).Err()
}

// ==================== Run Storage ====================

const runKeyPrefix = "run:"

func (r *RedisClient) SaveRun(ctx context.Context, run *model.Run) error {
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, runKeyPrefix+run.ID, data, 24*time.Hour).Err()
}

func (r *RedisClient) GetRun(ctx context.Context, runID string) (*model.Run, error) {
	data, err := r.client.Get(ctx, runKeyPrefix+runID).Bytes()
	if err != nil {
		return nil, err
	}
	var run model.Run
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *RedisClient) GetAllRuns(ctx context.Context) ([]*model.Run, error) {
	keys, err := r.client.Keys(ctx, runKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	runs := make([]*model.Run, 0, len(keys))
	for _, key := range keys {
		if len(key) > len(runKeyPrefix)+12 && key[len(runKeyPrefix)+12] == ':' {
			continue
		}
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var run model.Run
		if err := json.Unmarshal(data, &run); err != nil {
			continue
		}
		runs = append(runs, &run)
	}
	return runs, nil
}

// ==================== Run Queue ====================

const runQueueKey = "run:queue"

func (r *RedisClient) PushRunQueue(ctx context.Context, runID string) error {
	return r.client.RPush(ctx, runQueueKey, runID).Err()
}

func (r *RedisClient) PopRunQueue(ctx context.Context) (string, error) {
	result, err := r.client.LPop(ctx, runQueueKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// ==================== Run Pub/Sub ====================

func (r *RedisClient) PublishRunUpdate(ctx context.Context, runID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, "run:updates:"+runID, jsonData).Err()
}

func (r *RedisClient) SubscribeRunUpdates(ctx context.Context, runID string) *redis.PubSub {
	return r.client.Subscribe(ctx, "run:updates:"+runID)
}

// ==================== Schedule Storage ====================

const scheduleKeyPrefix = "schedule:"

func (r *RedisClient) SaveSchedule(ctx context.Context, scheduleID string, data []byte) error {
	return r.client.Set(ctx, scheduleKeyPrefix+scheduleID, data, 0).Err()
}

func (r *RedisClient) GetSchedule(ctx context.Context, scheduleID string) ([]byte, error) {
	return r.client.Get(ctx, scheduleKeyPrefix+scheduleID).Bytes()
}

func (r *RedisClient) DeleteSchedule(ctx context.Context, scheduleID string) error {
	return r.client.Del(ctx, scheduleKeyPrefix+scheduleID).Err()
}

func (r *RedisClient) ListScheduleKeys(ctx context.Context) ([]string, error) {
	keys, err := r.client.Keys(ctx, scheduleKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	// Strip prefix
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key[len(scheduleKeyPrefix):])
	}
	return result, nil
}

// ==================== Cache Storage ====================

func (r *RedisClient) SetCache(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, "cache:"+key, value, ttl).Err()
}

func (r *RedisClient) GetCache(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, "cache:"+key).Bytes()
}

func (r *RedisClient) DeleteCache(ctx context.Context, key string) error {
	return r.client.Del(ctx, "cache:"+key).Err()
}

// ==================== Rate Limiting ====================

func (r *RedisClient) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int, error) {
	val, err := r.client.Incr(ctx, "ratelimit:"+key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry on first increment
	if val == 1 {
		r.client.Expire(ctx, "ratelimit:"+key, window)
	}

	return int(val), nil
}

func (r *RedisClient) GetRateLimit(key string) (int, error) {
	val, err := r.client.Get(ctx, "ratelimit:"+key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}