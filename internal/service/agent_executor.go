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

// AgentExecutor Agent 执行器
type AgentExecutor struct {
	redis         *repository.RedisClient
	config        config.CLIConfig
	activeProcess sync.Map
	stopChan      chan struct{}
}

// NewAgentExecutor 创建 Agent 执行器
func NewAgentExecutor(redis *repository.RedisClient, cfg config.CLIConfig) *AgentExecutor {
	return &AgentExecutor{
		redis:    redis,
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

// ExecuteAgent 执行 Agent
func (e *AgentExecutor) ExecuteAgent(ctx context.Context, agent *model.Agent, input map[string]interface{}, workDir string) (json.RawMessage, error) {
	// 构建 Claude Code CLI 命令
	args := e.buildAgentArgs(agent, input)

	claudePath := e.config.ClaudePath
	if claudePath == "" {
		claudePath = "claude"
	}

	cmd := exec.CommandContext(ctx, claudePath, args...)
	cmd.Env = e.buildEnv(agent)

	if workDir != "" {
		cmd.Dir = workDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// 记录进程
	processKey := agent.ID + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
	e.activeProcess.Store(processKey, cmd)
	defer e.activeProcess.Delete(processKey)

	err := cmd.Wait()
	duration := time.Since(startTime)

	logger.Infof("Agent %s executed in %v", agent.Name, duration)

	if err != nil {
		return nil, fmt.Errorf("execution failed: %w, stderr: %s", err, stderr.String())
	}

	return e.parseOutput(stdout.String()), nil
}

// buildAgentArgs 构建 Agent 执行参数
func (e *AgentExecutor) buildAgentArgs(agent *model.Agent, input map[string]interface{}) []string {
	args := []string{}

	// 模型选择
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	} else {
		args = append(args, "--model", "claude-sonnet-4-6")
	}

	// 最大 tokens
	if agent.MaxTokens > 0 {
		args = append(args, "--max-tokens", strconv.Itoa(agent.MaxTokens))
	}

	// 输出格式
	args = append(args, "--output-format", "json")

	// 系统提示词
	if agent.SystemPrompt != "" {
		args = append(args, "--system", agent.SystemPrompt)
	}

	// 工具配置
	if len(agent.Tools) > 0 {
		toolsJSON, _ := json.Marshal(agent.Tools)
		args = append(args, "--tools", string(toolsJSON))
	}

	// 权限配置
	if len(agent.Permissions) > 0 {
		perms := make([]string, len(agent.Permissions))
		for i, p := range agent.Permissions {
			perms[i] = fmt.Sprintf("%s:%s", p.Resource, p.Action)
		}
		args = append(args, "--allowedTools", strings.Join(perms, ","))
	}

	// 构建提示词
	prompt := e.buildPrompt(agent, input)
	args = append(args, "--prompt", prompt)

	return args
}

// buildPrompt 构建执行提示词
func (e *AgentExecutor) buildPrompt(agent *model.Agent, input map[string]interface{}) string {
	var sb strings.Builder

	// 添加输入 Schema 提示
	if agent.InputSchema != nil {
		sb.WriteString("Expected input format:\n")
		sb.WriteString(string(agent.InputSchema))
		sb.WriteString("\n\n")
	}

	// 添加输入数据
	if len(input) > 0 {
		sb.WriteString("Input:\n")
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		sb.WriteString(string(inputJSON))
		sb.WriteString("\n\n")
	}

	// 添加输出要求
	if agent.OutputSchema != nil {
		sb.WriteString("Expected output format:\n")
		sb.WriteString(string(agent.OutputSchema))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// buildEnv 构建环境变量
func (e *AgentExecutor) buildEnv(agent *model.Agent) []string {
	env := os.Environ()

	// API Key
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+apiKey)
	}

	// 隔离命名空间
	if agent.Isolation.Namespace != "" {
		env = append(env, "AGENT_NAMESPACE="+agent.Isolation.Namespace)
	}

	// 数据隔离
	if agent.Isolation.DataIsolation {
		env = append(env, "AGENT_DATA_ISOLATION=true")
	}

	return env
}

// parseOutput 解析输出
func (e *AgentExecutor) parseOutput(output string) json.RawMessage {
	output = strings.TrimSpace(output)

	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		return json.RawMessage(output)
	}

	// 非 JSON 输出包装
	wrapped := map[string]interface{}{
		"raw":      output,
		"format":   "text",
		"truncated": len(output) > 10000,
	}
	data, _ := json.Marshal(wrapped)
	return data
}

// Stop 停止执行器
func (e *AgentExecutor) Stop() {
	close(e.stopChan)

	e.activeProcess.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(*exec.Cmd); ok {
			cmd.Process.Signal(syscall.SIGTERM)
		}
		return true
	})
}

// ExecuteWithRetry 带重试的执行
func (e *AgentExecutor) ExecuteWithRetry(ctx context.Context, agent *model.Agent, input map[string]interface{}, workDir string) (json.RawMessage, error) {
	maxRetries := 0
	backoff := "exponential"
	maxDelay := 60

	if agent.RetryPolicy.MaxRetries > 0 {
		maxRetries = agent.RetryPolicy.MaxRetries
		backoff = agent.RetryPolicy.Backoff
		maxDelay = agent.RetryPolicy.MaxDelay
	}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		result, err := e.ExecuteAgent(ctx, agent, input, workDir)
		if err == nil {
			return result, nil
		}

		lastErr = err
		logger.Warnf("Agent execution attempt %d failed: %v", i+1, err)

		if i < maxRetries {
			delay := e.calculateBackoff(i, backoff, maxDelay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// calculateBackoff 计算退避时间
func (e *AgentExecutor) calculateBackoff(attempt int, backoff string, maxDelay int) time.Duration {
	baseDelay := time.Second * time.Duration(1<<attempt)

	if backoff == "linear" {
		baseDelay = time.Second * time.Duration(attempt+1)
	}

	if int(baseDelay.Seconds()) > maxDelay {
		baseDelay = time.Second * time.Duration(maxDelay)
	}

	return baseDelay
}

// IsAllowed 检查权限
func (e *AgentExecutor) IsAllowed(agent *model.Agent, resource, action string) bool {
	for _, perm := range agent.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
		if perm.Resource == "*" || perm.Action == "*" {
			return true
		}
	}
	return false
}

// ValidateInput 验证输入
func (e *AgentExecutor) ValidateInput(agent *model.Agent, input map[string]interface{}) error {
	if agent.InputSchema == nil {
		return nil
	}

	// 简单验证 - 生产环境应使用 JSON Schema 验证库
	return nil
}

// ValidateOutput 验证输出
func (e *AgentExecutor) ValidateOutput(agent *model.Agent, output json.RawMessage) error {
	if agent.OutputSchema == nil {
		return nil
	}

	// 简单验证 - 生产环境应使用 JSON Schema 验证库
	return nil
}