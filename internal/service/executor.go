package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
)

type CLIExecutor struct {
	redis         *repository.RedisClient
	config        config.CLIConfig
	activeProcess sync.Map
	stopChan      chan struct{}
}

func NewCLIExecutor(redis *repository.RedisClient, cfg config.CLIConfig) *CLIExecutor {
	return &CLIExecutor{
		redis:    redis,
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

func (e *CLIExecutor) StartConsumer(ctx context.Context) {
	logger.Info("CLI executor started")

	for {
		select {
		case <-e.stopChan:
			return
		default:
			taskID, err := e.redis.PopTaskQueue(ctx)
			if err != nil {
				logger.Error("Failed to get task: ", err)
				time.Sleep(time.Second)
				continue
			}

			if taskID == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			go e.executeTask(ctx, taskID)
		}
	}
}

func (e *CLIExecutor) Stop() {
	close(e.stopChan)

	e.activeProcess.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(*exec.Cmd); ok {
			cmd.Process.Signal(syscall.SIGTERM)
		}
		return true
	})
}

func (e *CLIExecutor) executeTask(ctx context.Context, taskID string) {
	task, err := e.redis.GetTask(ctx, taskID)
	if err != nil {
		logger.Error("Failed to get task: ", err)
		return
	}

	// Update status to running
	now := time.Now()
	task.Status = model.TaskStatusRunning
	task.StartedAt = &now
	task.UpdatedAt = now
	e.redis.SaveTask(ctx, task)

	e.redis.PublishTaskUpdate(ctx, taskID, map[string]interface{}{
		"task_id": taskID,
		"status":  "running",
	})

	// Get skill config
	skill, err := e.redis.GetSkill(ctx, task.SkillID)
	if err != nil {
		e.failTask(ctx, task, "Failed to get skill config: "+err.Error())
		return
	}

	// Build command
	cmd, err := e.buildCommand(ctx, skill, task)
	if err != nil {
		e.failTask(ctx, task, "Failed to build command: "+err.Error())
		return
	}

	// Store process reference
	e.activeProcess.Store(taskID, cmd)
	defer e.activeProcess.Delete(taskID)

	// Execute command
	startTime := time.Now()
	output, err := e.runCommand(cmd, taskID)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		e.failTask(ctx, task, "Execution failed: "+err.Error())
		return
	}

	// Parse result
	result := e.parseOutput(output)

	// Update task as completed
	completedAt := time.Now()
	task.Status = model.TaskStatusCompleted
	task.Result = result
	task.Duration = duration
	task.CompletedAt = &completedAt
	task.UpdatedAt = completedAt
	e.redis.SaveTask(ctx, task)

	e.redis.PublishTaskUpdate(ctx, taskID, map[string]interface{}{
		"task_id": taskID,
		"status":  "completed",
		"result":  result,
	})

	logger.Infof("Task completed: %s, duration: %dms", taskID, duration)
}

func (e *CLIExecutor) buildCommand(ctx context.Context, skill *model.Skill, task *model.Task) (*exec.Cmd, error) {
	args := []string{}

	if skill.CLI != nil && skill.CLI.Model != "" {
		args = append(args, "--model", skill.CLI.Model)
	}

	if skill.CLI != nil && skill.CLI.MaxTokens > 0 {
		args = append(args, "--max-tokens", strconv.Itoa(skill.CLI.MaxTokens))
	}

	args = append(args, "--output-format", "json")

	prompt := e.renderPrompt(skill.Prompt, task.Parameters)
	args = append(args, "--prompt", prompt)

	claudePath := e.config.ClaudePath
	if claudePath == "" {
		claudePath = "claude"
	}

	cmd := exec.CommandContext(ctx, claudePath, args...)

	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
	)

	var contextData struct {
		WorkDir string `json:"work_dir"`
	}
	if len(task.Context) > 0 {
		json.Unmarshal(task.Context, &contextData)
	}
	if contextData.WorkDir != "" {
		cmd.Dir = contextData.WorkDir
	}

	return cmd, nil
}

func (e *CLIExecutor) runCommand(cmd *exec.Cmd, taskID string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if cmd.Process != nil {
		e.redis.SaveProcess(context.Background(), taskID, cmd.Process.Pid)
		defer e.redis.DeleteProcess(context.Background(), taskID)
	}

	err := cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

func (e *CLIExecutor) renderPrompt(template string, params json.RawMessage) string {
	var parameters map[string]interface{}
	json.Unmarshal(params, &parameters)

	result := template
	for key, value := range parameters {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (e *CLIExecutor) parseOutput(output string) json.RawMessage {
	output = strings.TrimSpace(output)

	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		return json.RawMessage(output)
	}

	result = map[string]interface{}{
		"raw": output,
	}
	data, _ := json.Marshal(result)
	return data
}

func (e *CLIExecutor) failTask(ctx context.Context, task *model.Task, errMsg string) {
	task.Status = model.TaskStatusFailed
	task.Error = errMsg
	task.UpdatedAt = time.Now()
	e.redis.SaveTask(ctx, task)

	e.redis.PublishTaskUpdate(ctx, task.ID, map[string]interface{}{
		"task_id": task.ID,
		"status":  "failed",
		"error":   errMsg,
	})

	logger.Errorf("Task failed: %s, reason: %s", task.ID, errMsg)
}

func (e *CLIExecutor) CancelTask(ctx context.Context, taskID string) error {
	value, ok := e.activeProcess.Load(taskID)
	if !ok {
		return fmt.Errorf("task not found or already completed")
	}

	cmd, ok := value.(*exec.Cmd)
	if !ok {
		return fmt.Errorf("invalid process")
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	task, err := e.redis.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.Status = model.TaskStatusCancelled
	task.UpdatedAt = time.Now()
	e.redis.SaveTask(ctx, task)

	return nil
}

func (e *CLIExecutor) GetStatus() map[string]interface{} {
	var activeTasks []string
	e.activeProcess.Range(func(key, value interface{}) bool {
		activeTasks = append(activeTasks, key.(string))
		return true
	})

	return map[string]interface{}{
		"active_count":    len(activeTasks),
		"max_concurrency": e.config.MaxConcurrency,
		"active_tasks":    activeTasks,
	}
}