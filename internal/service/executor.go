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

// ExecuteCLI executes a generic CLI command
func (e *CLIExecutor) ExecuteCLI(ctx context.Context, cli, action, command string, params map[string]interface{}) (json.RawMessage, error) {
	// Build command based on CLI type
	args := e.buildCLIArgs(cli, action, command, params)

	// Determine executable path
	execPath := e.getCLIPath(cli)

	cmd := exec.CommandContext(ctx, execPath, args...)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	err := cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return json.RawMessage(stdout.String()), nil
}

// buildCLIArgs builds arguments for different CLI types
func (e *CLIExecutor) buildCLIArgs(cli, action, command string, params map[string]interface{}) []string {
	args := []string{}

	switch cli {
	case "claude":
		args = e.buildClaudeArgs(action, params)
	case "npm":
		args = e.buildNpmArgs(action, command, params)
	case "git":
		args = e.buildGitArgs(action, params)
	default:
		if command != "" {
			args = strings.Fields(command)
		}
	}

	return args
}

func (e *CLIExecutor) buildClaudeArgs(action string, params map[string]interface{}) []string {
	args := []string{}

	// Model selection
	if model, ok := params["model"].(string); ok {
		args = append(args, "--model", model)
	} else {
		args = append(args, "--model", "claude-sonnet-4-6")
	}

	// Max tokens
	if maxTokens, ok := params["max_tokens"].(int); ok {
		args = append(args, "--max-tokens", strconv.Itoa(maxTokens))
	}

	args = append(args, "--output-format", "json")

	// Action-specific prompt
	prompt := e.buildActionPrompt(action, params)
	args = append(args, "--prompt", prompt)

	return args
}

func (e *CLIExecutor) buildActionPrompt(action string, params map[string]interface{}) string {
	switch action {
	case "analyze":
		target, _ := params["target"].(string)
		return fmt.Sprintf("Analyze the code at %s for quality, security, and performance issues. Output as JSON.", target)

	case "test-gen":
		source, _ := params["source"].(string)
		framework, _ := params["framework"].(string)
		return fmt.Sprintf("Generate %s tests for code at %s. Output complete test code.", framework, source)

	case "security-scan":
		target, _ := params["target"].(string)
		return fmt.Sprintf("Scan %s for security vulnerabilities. Output findings as JSON.", target)

	case "review":
		target, _ := params["target"].(string)
		depth, _ := params["depth"].(string)
		return fmt.Sprintf("Review code at %s with %s depth. Output review as JSON.", target, depth)

	default:
		if prompt, ok := params["prompt"].(string); ok {
			return prompt
		}
		return action
	}
}

func (e *CLIExecutor) buildNpmArgs(action, command string, params map[string]interface{}) []string {
	if command != "" {
		return []string{"run", command}
	}
	return []string{action}
}

func (e *CLIExecutor) buildGitArgs(action string, params map[string]interface{}) []string {
	args := []string{}
	switch action {
	case "clone":
		if repo, ok := params["repo"].(string); ok {
			args = append(args, "clone", repo)
		}
	case "pull":
		args = append(args, "pull")
	case "push":
		args = append(args, "push")
	}
	return args
}

func (e *CLIExecutor) getCLIPath(cli string) string {
	switch cli {
	case "claude":
		if e.config.ClaudePath != "" {
			return e.config.ClaudePath
		}
		return "claude"
	case "npm":
		return "npm"
	case "git":
		return "git"
	default:
		return cli
	}
}