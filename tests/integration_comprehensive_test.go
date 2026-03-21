package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisConnection tests Redis connectivity
func TestRedisConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1, // Use test database
	}

	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx)
	require.NoError(t, err, "Redis connection should succeed")
}

// TestSkillStorage tests skill CRUD operations
func TestSkillStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	skill := &model.Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill for integration testing",
		Version:     "1.0.0",
		Category:    "test",
		Enabled:     true,
	}

	// Test Save
	err := client.SaveSkill(ctx, skill)
	require.NoError(t, err)

	// Test Get
	retrieved, err := client.GetSkill(ctx, skill.ID)
	require.NoError(t, err)
	assert.Equal(t, skill.ID, retrieved.ID)
	assert.Equal(t, skill.Name, retrieved.Name)

	// Test GetAll
	skills, err := client.GetAllSkills(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, skills)
}

// TestTaskStorage tests task CRUD operations
func TestTaskStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	task := &model.Task{
		ID:        "test-task-" + time.Now().Format("20060102150405"),
		SkillID:   "code-review",
		Status:    "pending",
		Params:    map[string]interface{}{"target": "src/"},
		CreatedAt: time.Now(),
	}

	// Test Save
	err := client.SaveTask(ctx, task)
	require.NoError(t, err)

	// Test Get
	retrieved, err := client.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Status, retrieved.Status)

	// Test GetAll
	tasks, err := client.GetAllTasks(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks)

	// Test Output
	err = client.AppendTaskOutput(ctx, task.ID, "line 1")
	require.NoError(t, err)
	err = client.AppendTaskOutput(ctx, task.ID, "line 2")
	require.NoError(t, err)

	output, err := client.GetTaskOutput(ctx, task.ID)
	require.NoError(t, err)
	assert.Len(t, output, 2)
}

// TestTaskQueue tests task queue operations
func TestTaskQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	taskID := "queue-test-" + time.Now().Format("20060102150405")

	// Test Push
	err := client.PushTaskQueue(ctx, taskID)
	require.NoError(t, err)

	// Test Length
	length, err := client.GetQueueLength(ctx)
	require.NoError(t, err)
	assert.Greater(t, length, 0)

	// Test Pop
	popped, err := client.PopTaskQueue(ctx)
	require.NoError(t, err)
	assert.Equal(t, taskID, popped)
}

// TestPipelineStorage tests pipeline CRUD operations
func TestPipelineStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	pipeline := &model.Pipeline{
		ID:          "test-pipeline-" + time.Now().Format("20060102150405"),
		Name:        "Test Pipeline",
		Description: "A test pipeline",
		Mode:        "serial",
		Steps: []model.PipelineStep{
			{ID: "step1", CLI: "echo", Action: "test"},
		},
		CreatedAt: time.Now(),
	}

	// Test Save
	err := client.SavePipeline(ctx, pipeline)
	require.NoError(t, err)

	// Test Get
	retrieved, err := client.GetPipeline(ctx, pipeline.ID)
	require.NoError(t, err)
	assert.Equal(t, pipeline.ID, retrieved.ID)
	assert.Equal(t, pipeline.Name, retrieved.Name)

	// Test GetAll
	pipelines, err := client.GetAllPipelines(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, pipelines)

	// Test Delete
	err = client.DeletePipeline(ctx, pipeline.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = client.GetPipeline(ctx, pipeline.ID)
	assert.Error(t, err)
}

// TestPriorityQueue tests priority queue operations
func TestPriorityQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()
	queueName := "test-priority-queue"

	// Clear queue first
	client.ClearPriorityQueue(ctx, queueName)

	// Test Push with different priorities
	err := client.PushPriorityQueue(ctx, queueName, "low-priority", 10)
	require.NoError(t, err)
	err = client.PushPriorityQueue(ctx, queueName, "high-priority", 1)
	require.NoError(t, err)
	err = client.PushPriorityQueue(ctx, queueName, "medium-priority", 5)
	require.NoError(t, err)

	// Test Length
	length, err := client.GetPriorityQueueLength(ctx, queueName)
	require.NoError(t, err)
	assert.Equal(t, 3, length)

	// Test Pop (should get highest priority = lowest score)
	popped, err := client.PopPriorityQueue(ctx, queueName)
	require.NoError(t, err)
	assert.Equal(t, "high-priority", popped)

	// Clean up
	client.ClearPriorityQueue(ctx, queueName)
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()
	key := "rate-limit-test"

	// Test incrementing counter
	for i := 1; i <= 5; i++ {
		count, err := client.IncrementRateLimit(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, i, count)
	}

	// Verify count
	count, err := client.GetRateLimit(key)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

// TestPubSub tests pub/sub functionality
func TestPubSub(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()
	taskID := "pubsub-test"

	// Subscribe
	pubsub := client.SubscribeTaskUpdates(ctx, taskID)
	defer pubsub.Close()

	// Publish message
	testData := map[string]interface{}{
		"status":  "running",
		"message": "Test update",
	}
	err := client.PublishTaskUpdate(ctx, taskID, testData)
	require.NoError(t, err)

	// Receive message (with timeout)
	select {
	case msg := <-pubsub.Channel():
		var received map[string]interface{}
		err = json.Unmarshal([]byte(msg.Payload), &received)
		require.NoError(t, err)
		assert.Equal(t, "running", received["status"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for pub/sub message")
	}
}

// TestBatchOperations tests batch operations
func TestBatchOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.RedisConfig{Addr: "localhost:6379", DB: 1}
	client := repository.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	// Test batch save
	batchData := map[string]interface{}{
		"id":     "batch-test",
		"status": "pending",
	}
	batchJSON, _ := json.Marshal(batchData)

	err := client.SaveBatch(ctx, "test-batch-id", batchJSON)
	require.NoError(t, err)

	// Test batch get
	retrieved, err := client.GetBatch(ctx, "test-batch-id")
	require.NoError(t, err)
	assert.NotEmpty(t, retrieved)

	// Test batch list
	keys, err := client.ListBatchKeys(ctx)
	require.NoError(t, err)
	assert.Contains(t, keys, "test-batch-id")

	// Test batch delete
	err = client.DeleteBatch(ctx, "test-batch-id")
	require.NoError(t, err)
}