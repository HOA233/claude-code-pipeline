# Claude Code CLI 流水线服务

## 概述

本服务提供一个统一的 API 平台，用户可以选择不同的技能（Skill）来调用 Claude Code CLI。所有 CLI 实例运行在同一个服务中，但根据选择的技能执行不同的任务。

---

## 一、服务架构

### 1.1 整体架构

```
                    ┌─────────────────────────────────────────┐
                    │            用户 / 客户端                 │
                    └─────────────────┬───────────────────────┘
                                      │ HTTP API
                                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                      流水线服务 (单一服务)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  API 网关   │──│ 技能管理器  │──│  任务调度器  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│         │                │                 │                     │
│         │                │                 │                     │
│         ▼                ▼                 ▼                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    CLI 执行池                              │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │   │
│  │  │ CLI #1  │  │ CLI #2  │  │ CLI #3  │  │ CLI #n  │     │   │
│  │  │ 代码审查 │  │  部署   │  │ 测试生成 │  │ 重构   │     │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┴─────────────────┐
                    ▼                                   ▼
            ┌──────────────┐                  ┌──────────────┐
            │   GitLab     │                  │   结果存储    │
            │   技能仓库    │                  │   (数据库)    │
            └──────────────┘                  └──────────────┘
```

### 1.2 核心概念

| 概念 | 说明 |
|------|------|
| **技能 (Skill)** | 定义 Claude Code CLI 执行逻辑的模板，存储在 GitLab |
| **任务 (Task)** | 用户选择技能后创建的执行实例 |
| **CLI 实例** | 根据技能配置启动的 Claude Code CLI 进程 |

---

## 二、API 接口

### 2.1 获取可用技能列表

```http
GET /api/skills
```

**响应示例：**
```json
{
  "skills": [
    {
      "id": "code-review",
      "name": "代码审查",
      "description": "分析代码质量，检测潜在问题",
      "version": "1.2.0",
      "category": "quality",
      "parameters": [
        {
          "name": "target",
          "type": "string",
          "required": true,
          "description": "审查目标路径"
        },
        {
          "name": "depth",
          "type": "enum",
          "values": ["quick", "standard", "deep"],
          "default": "standard"
        }
      ]
    },
    {
      "id": "deploy",
      "name": "部署服务",
      "description": "自动化部署到指定环境",
      "version": "2.0.0",
      "category": "devops",
      "parameters": [
        {
          "name": "environment",
          "type": "enum",
          "values": ["dev", "staging", "production"],
          "required": true
        },
        {
          "name": "dry_run",
          "type": "boolean",
          "default": false
        }
      ]
    },
    {
      "id": "test-gen",
      "name": "测试生成",
      "description": "自动生成单元测试",
      "version": "1.0.0",
      "category": "testing",
      "parameters": [
        {
          "name": "source",
          "type": "string",
          "required": true
        },
        {
          "name": "framework",
          "type": "enum",
          "values": ["jest", "pytest", "go-test"]
        }
      ]
    },
    {
      "id": "refactor",
      "name": "代码重构",
      "description": "智能重构代码结构",
      "version": "1.1.0",
      "category": "development"
    },
    {
      "id": "docs-gen",
      "name": "文档生成",
      "description": "自动生成 API 文档",
      "version": "1.0.0",
      "category": "documentation"
    }
  ]
}
```

### 2.2 创建任务（选择技能执行）

```http
POST /api/tasks
```

**请求体：**
```json
{
  "skill_id": "code-review",
  "parameters": {
    "target": "src/api/",
    "depth": "deep"
  },
  "context": {
    "repository": "https://gitlab.company.com/team/project.git",
    "branch": "main",
    "commit": "abc123def"
  },
  "options": {
    "timeout": 600,
    "callback_url": "https://webhook.company.com/claude-callback",
    "tags": ["urgent", "production"]
  }
}
```

**响应：**
```json
{
  "task_id": "task-20240115-001",
  "status": "pending",
  "skill_id": "code-review",
  "created_at": "2024-01-15T09:00:00Z",
  "estimated_duration": "5-10 minutes"
}
```

### 2.3 查询任务状态

```http
GET /api/tasks/{task_id}
```

**响应：**
```json
{
  "task_id": "task-20240115-001",
  "status": "running",
  "skill_id": "code-review",
  "progress": {
    "current_step": 3,
    "total_steps": 5,
    "message": "分析代码结构..."
  },
  "started_at": "2024-01-15T09:00:01Z",
  "cli_instance": "cli-worker-03"
}
```

### 2.4 获取任务结果

```http
GET /api/tasks/{task_id}/result
```

**响应：**
```json
{
  "task_id": "task-20240115-001",
  "status": "completed",
  "skill_id": "code-review",
  "result": {
    "summary": {
      "files_analyzed": 42,
      "issues_found": 15,
      "risk_level": "medium"
    },
    "issues": [
      {
        "severity": "high",
        "type": "security",
        "file": "src/api/auth.js",
        "line": 45,
        "message": "SQL 注入风险",
        "suggestion": "使用参数化查询"
      },
      {
        "severity": "medium",
        "type": "performance",
        "file": "src/api/users.js",
        "line": 120,
        "message": "N+1 查询问题",
        "suggestion": "使用批量查询优化"
      }
    ],
    "report_url": "/reports/task-20240115-001.md"
  },
  "duration": "8m 32s",
  "completed_at": "2024-01-15T09:08:32Z"
}
```

### 2.5 取消任务

```http
DELETE /api/tasks/{task_id}
```

---

## 三、技能配置（GitLab 仓库）

### 3.1 仓库结构

```
claude-skills/
├── skills/
│   ├── code-review/
│   │   ├── skill.yaml          # 技能元数据
│   │   ├── prompt.md            # Claude 提示词模板
│   │   └── schema.json          # 参数校验
│   ├── deploy/
│   │   ├── skill.yaml
│   │   ├── prompt.md
│   │   └── scripts/
│   │       └── deploy.sh
│   ├── test-gen/
│   │   ├── skill.yaml
│   │   └── prompt.md
│   └── refactor/
│       ├── skill.yaml
│       └── prompt.md
├── shared/
│   ├── templates/
│   │   └── base-prompt.md
│   └── utils/
│       └── helpers.js
└── registry.json              # 技能注册表
```

### 3.2 技能定义 `skills/code-review/skill.yaml`

```yaml
id: code-review
name: 代码审查
description: 分析代码质量，检测安全漏洞和性能问题
version: 1.2.0
category: quality
author: team-ai

# CLI 配置
cli:
  model: claude-sonnet-4-6
  max_tokens: 8192
  timeout: 600

# 参数定义
parameters:
  - name: target
    type: string
    required: true
    description: 审查目标路径
    validation:
      pattern: "^[a-zA-Z0-9/_-]+$"

  - name: depth
    type: enum
    values: [quick, standard, deep]
    default: standard
    description: 审查深度

  - name: include_tests
    type: boolean
    default: false
    description: 是否包含测试文件

# 权限要求
permissions:
  - read:repository
  - write:report

# 输出配置
output:
  format: markdown
  save_to: /reports/{{task_id}}.md
```

### 3.3 提示词模板 `skills/code-review/prompt.md`

```markdown
# 代码审查任务

你是一个专业的代码审查专家。请对以下代码进行全面审查。

## 审查目标

- 路径: {{target}}
- 深度: {{depth}}
{{#if include_tests}}
- 包含测试文件: 是
{{/if}}

## 审查维度

请按以下维度进行审查：

### 1. 代码质量
- 代码风格一致性
- 命名规范
- 注释完整性

### 2. 安全性
- SQL 注入
- XSS 漏洞
- 敏感信息泄露
- 权限校验

### 3. 性能
- 算法复杂度
- 数据库查询优化
- 内存使用

### 4. 可维护性
- 代码复用
- 模块化程度
- 测试覆盖率

## 输出格式

请按以下 JSON 格式输出结果：

```json
{
  "summary": { ... },
  "issues": [ ... ]
}
```
```

### 3.4 技能注册表 `registry.json`

```json
{
  "version": "2.0",
  "last_updated": "2024-01-15T00:00:00Z",
  "skills": [
    {
      "id": "code-review",
      "path": "skills/code-review",
      "enabled": true,
      "tags": ["quality", "security", "performance"]
    },
    {
      "id": "deploy",
      "path": "skills/deploy",
      "enabled": true,
      "tags": ["devops", "automation"]
    },
    {
      "id": "test-gen",
      "path": "skills/test-gen",
      "enabled": true,
      "tags": ["testing", "automation"]
    },
    {
      "id": "refactor",
      "path": "skills/refactor",
      "enabled": true,
      "tags": ["development", "quality"]
    }
  ]
}
```

---

## 四、服务实现

### 4.1 项目结构

```
claude-pipeline-service/
├── cmd/
│   └── server/
│       └── main.go           # 服务入口
├── internal/
│   ├── api/
│   │   ├── handler.go        # HTTP 处理器
│   │   ├── middleware.go     # 中间件
│   │   └── routes.go         # 路由定义
│   ├── service/
│   │   ├── skill.go          # 技能服务
│   │   ├── task.go           # 任务服务
│   │   └── executor.go       # CLI 执行器
│   ├── model/
│   │   ├── skill.go          # 技能模型
│   │   └── task.go           # 任务模型
│   ├── repository/
│   │   └── redis.go          # Redis 存储层
│   └── config/
│       └── config.go         # 配置管理
├── pkg/
│   ├── gitlab/
│   │   └── client.go         # GitLab 客户端
│   └── logger/
│       └── logger.go         # 日志工具
├── config/
│   └── config.yaml
├── go.mod
├── go.sum
└── Dockerfile
```

### 4.2 Go 模块依赖 `go.mod`

```go
module github.com/company/claude-pipeline

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/redis/go-redis/v9 v9.5.1
    github.com/spf13/viper v1.18.2
    gopkg.in/yaml.v3 v3.0.1
    go.uber.org/zap v1.27.0
    github.com/google/uuid v1.6.0
)
```

### 4.3 主入口 `cmd/server/main.go`

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/company/claude-pipeline/internal/api"
    "github.com/company/claude-pipeline/internal/config"
    "github.com/company/claude-pipeline/internal/repository"
    "github.com/company/claude-pipeline/internal/service"
    "github.com/company/claude-pipeline/pkg/logger"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 初始化日志
    logger.Init(cfg.Log.Level)

    // 连接 Redis
    redisClient := repository.NewRedisClient(cfg.Redis)
    defer redisClient.Close()

    // 验证 Redis 连接
    if err := redisClient.Ping(context.Background()); err != nil {
        log.Fatalf("Redis 连接失败: %v", err)
    }

    // 初始化服务
    skillService := service.NewSkillService(redisClient, cfg.GitLab)
    taskService := service.NewTaskService(redisClient)
    executor := service.NewCLIExecutor(redisClient, cfg.CLI)

    // 启动任务消费者
    go executor.StartConsumer(context.Background())

    // 启动 HTTP 服务
    server := api.NewServer(cfg, skillService, taskService, executor)

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-quit
        logger.Info("正在关闭服务...")
        executor.Stop()
        server.Shutdown(context.Background())
    }()

    if err := server.Run(); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

### 4.4 配置管理 `internal/config/config.go`

```go
package config

import (
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    Server ServerConfig `mapstructure:"server"`
    Redis  RedisConfig  `mapstructure:"redis"`
    GitLab GitLabConfig `mapstructure:"gitlab"`
    CLI    CLIConfig    `mapstructure:"cli"`
    Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type RedisConfig struct {
    Addr     string `mapstructure:"addr"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type GitLabConfig struct {
    URL        string `mapstructure:"url"`
    Token      string `mapstructure:"token"`
    SkillsRepo string `mapstructure:"skills_repo"`
}

type CLIConfig struct {
    MaxConcurrency int           `mapstructure:"max_concurrency"`
    DefaultTimeout time.Duration `mapstructure:"default_timeout"`
    ClaudePath     string        `mapstructure:"claude_path"`
}

type LogConfig struct {
    Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")
    viper.AddConfigPath(".")

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### 4.5 Redis 存储层 `internal/repository/redis.go`

```go
package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/company/claude-pipeline/internal/model"
    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(cfg RedisConfig) *RedisClient {
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

// ==================== 技能存储 ====================

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

// ==================== 任务存储 ====================

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

func (r *RedisClient) UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus, result json.RawMessage) error {
    task, err := r.GetTask(ctx, taskID)
    if err != nil {
        return err
    }
    task.Status = status
    task.Result = result
    task.UpdatedAt = time.Now()
    return r.SaveTask(ctx, task)
}

func (r *RedisClient) AppendTaskOutput(ctx context.Context, taskID string, output string) error {
    key := taskKeyPrefix + taskID + ":output"
    return r.client.RPush(ctx, key, output).Err()
}

func (r *RedisClient) GetTaskOutput(ctx context.Context, taskID string) ([]string, error) {
    key := taskKeyPrefix + taskID + ":output"
    return r.client.LRange(ctx, key, 0, -1).Result()
}

// ==================== 任务队列 ====================

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

// ==================== 进程状态 ====================

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

// ==================== 发布订阅 ====================

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
```

### 4.6 数据模型 `internal/model/task.go`

```go
package model

import (
    "encoding/json"
    "time"
)

type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
    ID          string          `json:"id"`
    SkillID     string          `json:"skill_id"`
    Status      TaskStatus      `json:"status"`
    Parameters  json.RawMessage `json:"parameters"`
    Context     json.RawMessage `json:"context,omitempty"`
    Result      json.RawMessage `json:"result,omitempty"`
    Error       string          `json:"error,omitempty"`
    Duration    int64           `json:"duration,omitempty"` // 毫秒
    CreatedAt   time.Time       `json:"created_at"`
    StartedAt   *time.Time      `json:"started_at,omitempty"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

type TaskCreateRequest struct {
    SkillID    string                 `json:"skill_id" binding:"required"`
    Parameters map[string]interface{} `json:"parameters"`
    Context    map[string]interface{} `json:"context,omitempty"`
    Options    *TaskOptions           `json:"options,omitempty"`
}

type TaskOptions struct {
    Timeout     int    `json:"timeout,omitempty"`
    CallbackURL string `json:"callback_url,omitempty"`
}
```

### 4.7 数据模型 `internal/model/skill.go`

```go
package model

type Skill struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Version     string                 `json:"version"`
    Category    string                 `json:"category"`
    Parameters  []SkillParameter       `json:"parameters"`
    CLI         *CLIConfig             `json:"cli,omitempty"`
    Prompt      string                 `json:"prompt_template"`
    Tags        []string               `json:"tags"`
    Enabled     bool                   `json:"enabled"`
}

type SkillParameter struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Required    bool     `json:"required"`
    Description string   `json:"description"`
    Default     interface{} `json:"default,omitempty"`
    Values      []string `json:"values,omitempty"` // for enum type
}

type CLIConfig struct {
    Model     string `json:"model,omitempty"`
    MaxTokens int    `json:"max_tokens,omitempty"`
    Timeout   int    `json:"timeout,omitempty"`
}
```

### 4.8 CLI 执行器 `internal/service/executor.go`

```go
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

    "github.com/company/claude-pipeline/internal/model"
    "github.com/company/claude-pipeline/internal/repository"
    "github.com/company/claude-pipeline/pkg/logger"
)

type CLIExecutor struct {
    redis         *repository.RedisClient
    config        CLIConfig
    activeProcess sync.Map // taskID -> *exec.Cmd
    stopChan      chan struct{}
}

func NewCLIExecutor(redis *repository.RedisClient, cfg CLIConfig) *CLIExecutor {
    return &CLIExecutor{
        redis:    redis,
        config:   cfg,
        stopChan: make(chan struct{}),
    }
}

// StartConsumer 启动任务消费者
func (e *CLIExecutor) StartConsumer(ctx context.Context) {
    logger.Info("CLI 执行器已启动")

    for {
        select {
        case <-e.stopChan:
            return
        default:
            taskID, err := e.redis.PopTaskQueue(ctx)
            if err != nil {
                logger.Error("获取任务失败: ", err)
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

// Stop 停止执行器
func (e *CLIExecutor) Stop() {
    close(e.stopChan)

    // 终止所有运行中的进程
    e.activeProcess.Range(func(key, value interface{}) bool {
        if cmd, ok := value.(*exec.Cmd); ok {
            cmd.Process.Signal(syscall.SIGTERM)
        }
        return true
    })
}

// executeTask 执行单个任务
func (e *CLIExecutor) executeTask(ctx context.Context, taskID string) {
    // 获取任务信息
    task, err := e.redis.GetTask(ctx, taskID)
    if err != nil {
        logger.Error("获取任务失败: ", err)
        return
    }

    // 更新状态为运行中
    now := time.Now()
    task.Status = model.TaskStatusRunning
    task.StartedAt = &now
    task.UpdatedAt = now
    e.redis.SaveTask(ctx, task)

    // 发布状态更新
    e.redis.PublishTaskUpdate(ctx, taskID, map[string]interface{}{
        "task_id": taskID,
        "status":  "running",
    })

    // 获取技能配置
    skill, err := e.redis.GetSkill(ctx, task.SkillID)
    if err != nil {
        e.failTask(ctx, task, "获取技能配置失败: "+err.Error())
        return
    }

    // 构建命令
    cmd, err := e.buildCommand(ctx, skill, task)
    if err != nil {
        e.failTask(ctx, task, "构建命令失败: "+err.Error())
        return
    }

    // 存储进程引用
    e.activeProcess.Store(taskID, cmd)
    defer e.activeProcess.Delete(taskID)

    // 存储进程 PID
    if cmd.Process != nil {
        e.redis.SaveProcess(ctx, taskID, cmd.Process.Pid)
        defer e.redis.DeleteProcess(ctx, taskID)
    }

    // 执行命令
    startTime := time.Now()
    output, err := e.runCommand(cmd, taskID)
    duration := time.Since(startTime).Milliseconds()

    if err != nil {
        e.failTask(ctx, task, "执行失败: "+err.Error())
        return
    }

    // 解析结果
    result := e.parseOutput(output)

    // 更新任务完成
    completedAt := time.Now()
    task.Status = model.TaskStatusCompleted
    task.Result = result
    task.Duration = duration
    task.CompletedAt = &completedAt
    task.UpdatedAt = completedAt
    e.redis.SaveTask(ctx, task)

    // 发布完成通知
    e.redis.PublishTaskUpdate(ctx, taskID, map[string]interface{}{
        "task_id": taskID,
        "status":  "completed",
        "result":  result,
    })

    logger.Info(fmt.Sprintf("任务完成: %s, 耗时: %dms", taskID, duration))
}

// buildCommand 构建 CLI 命令
func (e *CLIExecutor) buildCommand(ctx context.Context, skill *model.Skill, task *model.Task) (*exec.Cmd, error) {
    args := []string{}

    // 模型选择
    if skill.CLI != nil && skill.CLI.Model != "" {
        args = append(args, "--model", skill.CLI.Model)
    }

    // 最大 token
    if skill.CLI != nil && skill.CLI.MaxTokens > 0 {
        args = append(args, "--max-tokens", strconv.Itoa(skill.CLI.MaxTokens))
    }

    // 输出格式
    args = append(args, "--output-format", "json")

    // 渲染提示词
    prompt := e.renderPrompt(skill.Prompt, task.Parameters)
    args = append(args, "--prompt", prompt)

    // 创建命令
    claudePath := e.config.ClaudePath
    if claudePath == "" {
        claudePath = "claude"
    }

    cmd := exec.CommandContext(ctx, claudePath, args...)

    // 设置环境变量
    cmd.Env = append(os.Environ(),
        "ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
    )

    // 设置工作目录
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

// runCommand 运行命令并捕获输出
func (e *CLIExecutor) runCommand(cmd *exec.Cmd, taskID string) (string, error) {
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // 实时输出处理
    stdoutReader, _ := cmd.StdoutPipe()
    go e.streamOutput(taskID, stdoutReader)

    // 启动命令
    if err := cmd.Start(); err != nil {
        return "", err
    }

    // 等待完成
    err := cmd.Wait()
    if err != nil {
        return "", fmt.Errorf("%s: %w", stderr.String(), err)
    }

    return stdout.String(), nil
}

// streamOutput 流式输出
func (e *CLIExecutor) streamOutput(taskID string, reader *os.File) {
    buf := make([]byte, 1024)
    for {
        n, err := reader.Read(buf)
        if n > 0 {
            output := string(buf[:n])
            e.redis.AppendTaskOutput(context.Background(), taskID, output)
            e.redis.PublishTaskUpdate(context.Background(), taskID, map[string]interface{}{
                "task_id": taskID,
                "type":    "output",
                "data":    output,
            })
        }
        if err != nil {
            break
        }
    }
}

// renderPrompt 渲染提示词模板
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

// parseOutput 解析输出
func (e *CLIExecutor) parseOutput(output string) json.RawMessage {
    output = strings.TrimSpace(output)

    // 尝试解析 JSON
    var result interface{}
    if err := json.Unmarshal([]byte(output), &result); err == nil {
        return json.RawMessage(output)
    }

    // 返回原始文本
    result = map[string]interface{}{
        "raw": output,
    }
    data, _ := json.Marshal(result)
    return data
}

// failTask 标记任务失败
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

    logger.Error(fmt.Sprintf("任务失败: %s, 原因: %s", task.ID, errMsg))
}

// CancelTask 取消任务
func (e *CLIExecutor) CancelTask(ctx context.Context, taskID string) error {
    value, ok := e.activeProcess.Load(taskID)
    if !ok {
        return fmt.Errorf("任务不存在或已完成")
    }

    cmd, ok := value.(*exec.Cmd)
    if !ok {
        return fmt.Errorf("无效的进程")
    }

    // 发送终止信号
    if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
        return err
    }

    // 更新任务状态
    task, err := e.redis.GetTask(ctx, taskID)
    if err != nil {
        return err
    }

    task.Status = model.TaskStatusCancelled
    task.UpdatedAt = time.Now()
    e.redis.SaveTask(ctx, task)

    return nil
}

// GetStatus 获取执行器状态
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
```

### 4.9 任务服务 `internal/service/task.go`

```go
package service

import (
    "context"
    "encoding/json"
    "time"

    "github.com/company/claude-pipeline/internal/model"
    "github.com/company/claude-pipeline/internal/repository"
    "github.com/google/uuid"
)

type TaskService struct {
    redis *repository.RedisClient
}

func NewTaskService(redis *repository.RedisClient) *TaskService {
    return &TaskService{redis: redis}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, req *model.TaskCreateRequest) (*model.Task, error) {
    // 验证技能是否存在
    skill, err := s.redis.GetSkill(ctx, req.SkillID)
    if err != nil {
        return nil, err
    }

    // 验证参数
    if err := s.validateParameters(skill, req.Parameters); err != nil {
        return nil, err
    }

    // 创建任务
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

    // 保存任务
    if err := s.redis.SaveTask(ctx, task); err != nil {
        return nil, err
    }

    // 加入任务队列
    if err := s.redis.PushTaskQueue(ctx, task.ID); err != nil {
        return nil, err
    }

    return task, nil
}

// GetTask 获取任务
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
    return s.redis.GetTask(ctx, taskID)
}

// GetTaskResult 获取任务结果
func (s *TaskService) GetTaskResult(ctx context.Context, taskID string) (map[string]interface{}, error) {
    task, err := s.redis.GetTask(ctx, taskID)
    if err != nil {
        return nil, err
    }

    result := map[string]interface{}{
        "task_id":     task.ID,
        "status":      task.Status,
        "skill_id":    task.SkillID,
        "duration":    task.Duration,
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

// validateParameters 验证参数
func (s *TaskService) validateParameters(skill *model.Skill, params map[string]interface{}) error {
    for _, param := range skill.Parameters {
        value, exists := params[param.Name]

        // 必填检查
        if param.Required && !exists {
            return fmt.Errorf("缺少必填参数: %s", param.Name)
        }

        // 枚举值检查
        if param.Type == "enum" && len(param.Values) > 0 {
            strValue := fmt.Sprintf("%v", value)
            valid := false
            for _, v := range param.Values {
                if v == strValue {
                    valid = true
                    break
                }
            }
            if !valid {
                return fmt.Errorf("参数 %s 值无效, 允许值: %v", param.Name, param.Values)
            }
        }
    }
    return nil
}

// SubscribeTaskUpdates 订阅任务更新
func (s *TaskService) SubscribeTaskUpdates(ctx context.Context, taskID string) *redis.PubSub {
    return s.redis.SubscribeTaskUpdates(ctx, taskID)
}
```

### 4.10 HTTP 路由 `internal/api/routes.go`

```go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/company/claude-pipeline/internal/service"
)

func SetupRoutes(r *gin.Engine, skillSvc *service.SkillService, taskSvc *service.TaskService, executor *service.CLIExecutor) {
    api := r.Group("/api")
    {
        // 技能相关
        api.GET("/skills", ListSkills(skillSvc))
        api.GET("/skills/:id", GetSkill(skillSvc))
        api.POST("/skills/sync", SyncSkills(skillSvc))

        // 任务相关
        api.POST("/tasks", CreateTask(taskSvc))
        api.GET("/tasks", ListTasks(taskSvc))
        api.GET("/tasks/:id", GetTask(taskSvc))
        api.GET("/tasks/:id/result", GetTaskResult(taskSvc))
        api.DELETE("/tasks/:id", CancelTask(executor))

        // WebSocket 实时输出
        api.GET("/ws/tasks/:id/output", TaskOutputWS(taskSvc))

        // 服务状态
        api.GET("/status", GetStatus(executor, skillSvc))
    }
}
```

### 4.11 HTTP 处理器 `internal/api/handler.go`

```go
package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/company/claude-pipeline/internal/model"
    "github.com/company/claude-pipeline/internal/service"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// ==================== 技能处理器 ====================

func ListSkills(svc *service.SkillService) gin.HandlerFunc {
    return func(c *gin.Context) {
        skills, err := svc.GetAllSkills(c.Request.Context())
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"skills": skills})
    }
}

func GetSkill(svc *service.SkillService) gin.HandlerFunc {
    return func(c *gin.Context) {
        skillID := c.Param("id")
        skill, err := svc.GetSkill(c.Request.Context(), skillID)
        if err != nil {
            c.JSON(404, gin.H{"error": "技能不存在"})
            return
        }
        c.JSON(200, skill)
    }
}

func SyncSkills(svc *service.SkillService) gin.HandlerFunc {
    return func(c *gin.Context) {
        skills, err := svc.SyncFromGitLab(c.Request.Context())
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{
            "message": "同步成功",
            "count":   len(skills),
        })
    }
}

// ==================== 任务处理器 ====================

func CreateTask(svc *service.TaskService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req model.TaskCreateRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        task, err := svc.CreateTask(c.Request.Context(), &req)
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        c.JSON(202, task)
    }
}

func GetTask(svc *service.TaskService) gin.HandlerFunc {
    return func(c *gin.Context) {
        taskID := c.Param("id")
        task, err := svc.GetTask(c.Request.Context(), taskID)
        if err != nil {
            c.JSON(404, gin.H{"error": "任务不存在"})
            return
        }
        c.JSON(200, task)
    }
}

func GetTaskResult(svc *service.TaskService) gin.HandlerFunc {
    return func(c *gin.Context) {
        taskID := c.Param("id")
        result, err := svc.GetTaskResult(c.Request.Context(), taskID)
        if err != nil {
            c.JSON(404, gin.H{"error": "结果不存在"})
            return
        }
        c.JSON(200, result)
    }
}

func CancelTask(executor *service.CLIExecutor) gin.HandlerFunc {
    return func(c *gin.Context) {
        taskID := c.Param("id")
        if err := executor.CancelTask(c.Request.Context(), taskID); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"cancelled": true, "task_id": taskID})
    }
}

// ==================== WebSocket 处理器 ====================

func TaskOutputWS(svc *service.TaskService) gin.HandlerFunc {
    return func(c *gin.Context) {
        taskID := c.Param("id")

        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            return
        }
        defer conn.Close()

        // 订阅 Redis 频道
        pubsub := svc.SubscribeTaskUpdates(c.Request.Context(), taskID)
        defer pubsub.Close()

        ch := pubsub.Channel()

        for {
            select {
            case msg, ok := <-ch:
                if !ok {
                    return
                }
                if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
                    return
                }
            case <-c.Request.Context().Done():
                return
            }
        }
    }
}

// ==================== 状态处理器 ====================

func GetStatus(executor *service.CLIExecutor, skillSvc *service.SkillService) gin.HandlerFunc {
    return func(c *gin.Context) {
        skills, _ := skillSvc.GetAllSkills(c.Request.Context())
        c.JSON(200, gin.H{
            "status": "healthy",
            "cli":    executor.GetStatus(),
            "skills": gin.H{
                "loaded": len(skills),
            },
        })
    }
}
```

---

## 五、API 操作 CLI 的机制

### 5.1 核心原理

API 通过 **子进程 (Child Process)** 方式操作 Claude Code CLI：

```
┌──────────────────────────────────────────────────────────────┐
│                      API 服务进程                             │
│                                                              │
│  ┌────────────────┐         spawn()        ┌──────────────┐ │
│  │                │ ──────────────────────▶│              │ │
│  │  CLIExecutor   │                        │ Claude CLI   │ │
│  │                │◀────────────────────── │ 子进程       │ │
│  └────────────────┘        stdout/stderr    └──────────────┘ │
│         │                                        │          │
│         │                                        │          │
│         ▼                                        ▼          │
│  ┌────────────────┐                      ┌──────────────┐   │
│  │  任务队列      │                      │ 工作目录     │   │
│  │  (Redis)       │                      │ (项目代码)   │   │
│  └────────────────┘                      └──────────────┘   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 5.2 CLI 命令构造

#### 5.2.1 基础命令结构

```bash
claude [选项] --prompt "<提示词>" [文件或目录]
```

#### 5.2.2 API 构建命令示例

```javascript
// 用户请求
{
  skill_id: "code-review",
  parameters: { target: "src/", depth: "deep" }
}

// API 构建的最终命令
claude \
  --model claude-sonnet-4-6 \
  --max-tokens 8192 \
  --output-format json \
  --prompt "你是一个专业的代码审查专家。请对以下代码进行全面审查..." \
  src/
```

#### 5.2.3 CLI 参数映射

| 技能配置 (skill.yaml) | CLI 参数 | 说明 |
|----------------------|----------|------|
| `cli.model` | `--model` | 指定 Claude 模型 |
| `cli.max_tokens` | `--max-tokens` | 最大输出 token |
| `cli.timeout` | 进程超时控制 | 执行时间限制 |
| `output.format` | `--output-format` | 输出格式 (json/text) |
| `parameters.target` | 位置参数 | 工作目录或文件 |

### 5.3 进程生命周期管理

```
           API 请求
              │
              ▼
    ┌─────────────────┐
    │ 1. 验证参数     │
    │    检查技能配置 │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │ 2. 构建命令     │
    │    生成 CLI 参数│
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │ 3. 创建工作目录 │
    │    git clone    │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │ 4. 启动子进程   │ ◀─── spawn('claude', args)
    │    记录 PID     │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │ 5. 流式读取输出 │ ◀─── stdout.on('data')
    │    实时更新状态 │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │ 6. 进程结束     │ ◀─── process.on('close')
    │    返回结果     │
    └─────────────────┘
```

### 5.4 代码实现详解

#### 5.4.1 进程启动

```javascript
const { spawn } = require('child_process');

class CLIExecutor {
  async execute(skillId, skillConfig, parameters, context) {
    // 1. 构建提示词
    const prompt = this.renderPrompt(skillConfig.promptTemplate, parameters);

    // 2. 构建 CLI 参数
    const args = this.buildArgs(skillConfig, prompt);

    // 3. 准备工作目录
    const workDir = await this.prepareWorkDir(context);

    // 4. 启动子进程
    const cliProcess = spawn('claude', args, {
      cwd: workDir,                              // 工作目录
      env: {                                     // 环境变量
        ...process.env,
        ANTHROPIC_API_KEY: process.env.ANTHROPIC_API_KEY,
        CLAUDE_MODEL: skillConfig.cli?.model,
      },
      stdio: ['pipe', 'pipe', 'pipe'],          // 标准IO配置
      detached: false,                           // 进程归属
    });

    return this.handleProcess(cliProcess, skillId);
  }
}
```

#### 5.4.2 参数构建

```javascript
buildArgs(skillConfig, prompt) {
  const args = [];

  // 模型选择
  if (skillConfig.cli?.model) {
    args.push('--model', skillConfig.cli.model);
  }

  // 最大 token 数
  if (skillConfig.cli?.max_tokens) {
    args.push('--max-tokens', String(skillConfig.cli.max_tokens));
  }

  // 输出格式
  args.push('--output-format', skillConfig.output?.format || 'json');

  // 允许的工具/权限
  if (skillConfig.permissions?.length) {
    args.push('--allowedTools', skillConfig.permissions.join(','));
  }

  // 提示词 (通过 stdin 或参数)
  args.push('--prompt', prompt);

  return args;
}
```

#### 5.4.3 输出流处理

```javascript
handleProcess(cliProcess, skillId) {
  const taskId = generateTaskId();

  return new Promise((resolve, reject) => {
    const output = {
      taskId,
      skillId,
      status: 'running',
      stdout: [],
      stderr: [],
      startTime: Date.now(),
    };

    // 实时读取标准输出
    cliProcess.stdout.on('data', (data) => {
      const chunk = data.toString();
      output.stdout.push(chunk);

      // 解析进度信息
      const progress = this.parseProgress(chunk);
      if (progress) {
        this.updateTaskProgress(taskId, progress);
      }

      // WebSocket 推送实时输出
      this.broadcastOutput(taskId, chunk);
    });

    // 读取错误输出
    cliProcess.stderr.on('data', (data) => {
      output.stderr.push(data.toString());
    });

    // 进程结束
    cliProcess.on('close', (exitCode) => {
      output.status = exitCode === 0 ? 'completed' : 'failed';
      output.exitCode = exitCode;
      output.duration = Date.now() - output.startTime;
      output.result = this.parseResult(output.stdout.join(''));

      // 清理进程引用
      this.activeProcesses.delete(taskId);

      if (exitCode === 0) {
        resolve(output);
      } else {
        reject(new Error(output.stderr.join('\n')));
      }
    });

    // 进程错误处理
    cliProcess.on('error', (err) => {
      output.status = 'error';
      output.error = err.message;
      reject(err);
    });

    // 存储进程引用 (用于取消/状态查询)
    this.activeProcesses.set(taskId, {
      process: cliProcess,
      output,
      skillId,
    });
  });
}
```

#### 5.4.4 任务取消

```javascript
cancel(taskId) {
  const task = this.activeProcesses.get(taskId);
  if (!task) return false;

  const { process } = task;

  // 1. 发送 SIGTERM 信号 (优雅退出)
  process.kill('SIGTERM');

  // 2. 超时后强制终止
  setTimeout(() => {
    if (this.activeProcesses.has(taskId)) {
      process.kill('SIGKILL');
    }
  }, 5000);

  this.activeProcesses.delete(taskId);
  return true;
}
```

### 5.5 输入输出通信

#### 5.5.1 输入方式

```javascript
// 方式 1: 命令行参数
spawn('claude', ['--prompt', promptText]);

// 方式 2: 标准输入 (适合长提示词)
const cliProcess = spawn('claude', ['--stdin']);
cliProcess.stdin.write(promptText);
cliProcess.stdin.end();

// 方式 3: 文件输入
fs.writeFileSync('/tmp/prompt.txt', promptText);
spawn('claude', ['--prompt-file', '/tmp/prompt.txt']);
```

#### 5.5.2 输出解析

```javascript
parseResult(rawOutput) {
  try {
    // 尝试解析 JSON 输出
    return JSON.parse(rawOutput);
  } catch {
    // 解析文本输出
    return {
      raw: rawOutput,
      summary: this.extractSummary(rawOutput),
      actions: this.extractActions(rawOutput),
    };
  }
}

extractSummary(text) {
  // 从 CLI 输出中提取摘要信息
  const match = text.match(/## Summary\n([\s\S]*?)(?=\n##|$)/);
  return match ? match[1].trim() : null;
}

extractActions(text) {
  // 提取 CLI 执行的操作列表
  const actionRegex = /- \[x\] (.+)/g;
  const actions = [];
  let match;
  while ((match = actionRegex.exec(text)) !== null) {
    actions.push(match[1]);
  }
  return actions;
}
```

### 5.6 完整执行流程

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           完整执行流程                                    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  用户请求 ──▶ POST /api/tasks                                            │
│       │                                                                  │
│       ▼                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 1. API 层                                                        │    │
│  │    • 验证 skill_id                                               │    │
│  │    • 校验 parameters                                             │    │
│  │    • 创建 task_id                                                │    │
│  │    • 返回 202 Accepted                                           │    │
│  └───────────────────────────────┬─────────────────────────────────┘    │
│                                  │                                       │
│                                  ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 2. SkillManager                                                  │    │
│  │    • 加载 skill.yaml                                             │    │
│  │    • 读取 prompt.md 模板                                         │    │
│  │    • 渲染提示词: {{target}} → "src/"                             │    │
│  └───────────────────────────────┬─────────────────────────────────┘    │
│                                  │                                       │
│                                  ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 3. CLIExecutor                                                   │    │
│  │    • 构建命令行参数                                              │    │
│  │    • 准备工作目录                                                │    │
│  │    • spawn('claude', args)                                       │    │
│  └───────────────────────────────┬─────────────────────────────────┘    │
│                                  │                                       │
│                                  ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 4. Claude CLI 子进程                                             │    │
│  │                                                                  │    │
│  │    $ claude --model claude-sonnet-4-6 \                          │    │
│  │        --max-tokens 8192 \                                       │    │
│  │        --output-format json \                                    │    │
│  │        --prompt "代码审查..." \                                   │    │
│  │        src/                                                      │    │
│  │                                                                  │    │
│  │    执行中:                                                        │    │
│  │    ├── 读取文件 src/api/auth.js                                  │    │
│  │    ├── 分析代码结构                                              │    │
│  │    ├── 检测安全漏洞                                              │    │
│  │    └── 生成报告                                                  │    │
│  └───────────────────────────────┬─────────────────────────────────┘    │
│                                  │                                       │
│                                  ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 5. 输出流处理                                                    │    │
│  │    • stdout ──▶ 实时解析 + WebSocket 推送                        │    │
│  │    • stderr ──▶ 错误日志记录                                    │    │
│  │    • close  ──▶ 解析最终结果                                    │    │
│  └───────────────────────────────┬─────────────────────────────────┘    │
│                                  │                                       │
│                                  ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ 6. 结果存储                                                      │    │
│  │    • 保存到 Redis/数据库                                         │    │
│  │    • 生成报告文件                                                │    │
│  │    • 回调 webhook (如有配置)                                     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### 5.7 并发控制

```javascript
class CLIExecutor {
  constructor(config) {
    this.maxConcurrency = config.maxConcurrency || 5;
    this.queue = [];
    this.active = new Map();
  }

  async execute(skillId, skillConfig, parameters, context) {
    // 检查并发限制
    if (this.active.size >= this.maxConcurrency) {
      // 加入等待队列
      return new Promise((resolve, reject) => {
        this.queue.push({ skillId, skillConfig, parameters, context, resolve, reject });
      });
    }

    // 执行任务
    const task = this.runProcess(skillId, skillConfig, parameters, context);
    this.active.set(task.taskId, task);

    try {
      const result = await task.promise;
      return result;
    } finally {
      this.active.delete(task.taskId);
      this.processQueue();
    }
  }

  processQueue() {
    if (this.queue.length > 0 && this.active.size < this.maxConcurrency) {
      const next = this.queue.shift();
      this.execute(next.skillId, next.skillConfig, next.parameters, next.context)
        .then(next.resolve)
        .catch(next.reject);
    }
  }
}
```

### 5.8 超时处理

```javascript
executeWithTimeout(cliProcess, timeoutMs) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      cliProcess.kill('SIGTERM');
      reject(new Error(`任务超时 (${timeoutMs}ms)`));
    }, timeoutMs);

    cliProcess.on('close', (code) => {
      clearTimeout(timer);
      resolve(code);
    });
  });
}
```

---

## 六、配置文件

### 6.1 服务配置 `config/config.yaml`

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

redis:
  addr: localhost:6379
  password: ""
  db: 0

gitlab:
  url: https://gitlab.company.com
  token: ${GITLAB_TOKEN}
  skills_repo: ai/claude-skills

cli:
  max_concurrency: 5
  default_timeout: 600s
  claude_path: claude

log:
  level: info
```

### 6.2 Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git

# 安装 Claude CLI
RUN apk add --no-cache nodejs npm && \
    npm install -g @anthropic-ai/claude-code

WORKDIR /app

# 复制二进制文件
COPY --from=builder /server .
COPY --from=builder /app/config ./config

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./server"]
```

### 6.3 Docker Compose `docker-compose.yml`

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - GITLAB_TOKEN=${GITLAB_TOKEN}
    depends_on:
      redis:
        condition: service_healthy
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  redis-data:
```

### 6.4 Makefile

```makefile
.PHONY: build run test clean docker

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

docker:
	docker-compose up -d --build

docker-down:
	docker-compose down

logs:
	docker-compose logs -f api

redis-cli:
	docker-compose exec redis redis-cli
```

---

## 七、使用示例

### 7.1 选择代码审查技能

```bash
# 创建代码审查任务
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "code-review",
    "parameters": {
      "target": "src/",
      "depth": "deep"
    }
  }'
```

### 7.2 选择部署技能

```bash
# 创建部署任务
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "deploy",
    "parameters": {
      "environment": "staging",
      "dry_run": true
    }
  }'
```

### 7.3 选择测试生成技能

```bash
# 创建测试生成任务
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "test-gen",
    "parameters": {
      "source": "src/utils/",
      "framework": "jest"
    }
  }'
```

### 7.4 查询任务

```bash
# 查询任务状态
curl http://localhost:8080/api/tasks/task-20240115-001

# 获取任务结果
curl http://localhost:8080/api/tasks/task-20240115-001/result
```

---

## 八、Redis 数据结构

### 8.1 数据存储架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Redis 数据存储                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │  skill:{id}     │     │  task:{id}      │                   │
│  ├─────────────────┤     ├─────────────────┤                   │
│  │ {               │     │ {               │                   │
│  │   "id": "...",  │     │   "id": "...",  │                   │
│  │   "name": "...",│     │   "status":..., │                   │
│  │   "prompt":..., │     │   "result":..., │                   │
│  │   ...           │     │   ...           │                   │
│  │ }               │     │ }               │                   │
│  └─────────────────┘     └─────────────────┘                   │
│                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │  task:queue     │     │  task:{id}:output│                  │
│  ├─────────────────┤     ├─────────────────┤                   │
│  │  List 类型      │     │  List 类型       │                   │
│  │  ─────────────  │     │  ──────────────  │                   │
│  │  task-001       │     │  "输出行1..."    │                   │
│  │  task-002       │     │  "输出行2..."    │                   │
│  │  ...            │     │  ...             │                   │
│  └─────────────────┘     └─────────────────┘                   │
│                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │  process:{id}   │     │  task:updates:{id}│                  │
│  ├─────────────────┤     ├─────────────────┤                   │
│  │  PID (int)      │     │  Pub/Sub 频道    │                   │
│  │  用于取消任务   │     │  实时推送更新    │                   │
│  └─────────────────┘     └─────────────────┘                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 8.2 Key 设计

| Key | 类型 | 说明 | TTL |
|-----|------|------|-----|
| `skill:{id}` | String (JSON) | 技能配置 | 永久 |
| `task:{id}` | String (JSON) | 任务信息 | 24小时 |
| `task:queue` | List | 任务队列 | 永久 |
| `task:{id}:output` | List | 任务输出日志 | 24小时 |
| `process:{id}` | String | 进程 PID | 永久 |
| `task:updates:{id}` | Pub/Sub | 实时更新频道 | - |

### 8.3 常用 Redis 命令

```bash
# 查看所有技能
KEYS skill:*

# 查看任务队列
LLEN task:queue
LRANGE task:queue 0 -1

# 查看任务详情
GET task:task-abc123

# 查看任务输出
LRANGE task:task-abc123:output 0 -1

# 发布任务更新
PUBLISH task:updates:task-abc123 '{"status":"running"}'

# 订阅任务更新
SUBSCRIBE task:updates:task-abc123
```

---

## 九、流程图

```
用户请求
    │
    ▼
┌─────────────────┐
│  选择技能       │  GET /api/skills
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  提交参数       │  POST /api/tasks
│  skill_id       │
│  parameters     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  任务调度器     │  验证参数
│                 │  创建任务
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  加载技能配置   │  从 GitLab 拉取
│                 │  解析 skill.yaml
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  启动 CLI 实例  │  根据技能配置
│                 │  执行 claude 命令
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  返回结果       │  GET /api/tasks/{id}/result
│                 │
└─────────────────┘
```

---

## 十、前端界面

### 10.1 界面概览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Claude Code CLI 流水线服务                                    [用户: Admin] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐  ┌─────────────────────────────────────────────────────┐  │
│  │             │  │  任务列表                                          [+] │  │
│  │  技能列表   │  ├─────────────────────────────────────────────────────┤  │
│  │             │  │  ● task-001  code-review   运行中   ▶ 查看详情      │  │
│  │  ┌───────┐  │  │  ✓ task-002  deploy       已完成   ▶ 查看结果      │  │
│  │  │代码审查│  │  │  ✗ task-003  test-gen     失败    ▶ 查看错误      │  │
│  │  └───────┘  │  │  ✓ task-004  refactor     已完成   ▶ 查看结果      │  │
│  │  ┌───────┐  │  │  ● task-005  docs-gen     运行中   ▶ 查看详情      │  │
│  │  │ 部署  │  │  └─────────────────────────────────────────────────────┘  │
│  │  └───────┘  │                                                            │
│  │  ┌───────┐  │  ┌─────────────────────────────────────────────────────┐  │
│  │  │测试生成│  │  │  实时输出                                            [×] │  │
│  │  └───────┘  │  ├─────────────────────────────────────────────────────┤  │
│  │  ┌───────┐  │  │  [task-001] code-review                              │  │
│  │  │ 重构  │  │  │  ───────────────────────────────────────────────────│  │
│  │  └───────┘  │  │  ▶ 正在分析 src/api/auth.js...                       │  │
│  │  ┌───────┐  │  │  ▶ 检测到 2 个安全问题                               │  │
│  │  │文档生成│  │  │  ▶ 正在分析 src/utils/helper.js...                   │  │
│  │  └───────┘  │  │  █                                                    │  │
│  │             │  └─────────────────────────────────────────────────────┘  │
│  └─────────────┘                                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.2 技术栈

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 框架 | React 18 / Vue 3 | 现代前端框架 |
| UI 组件 | Ant Design / Element Plus | 企业级组件库 |
| 状态管理 | Zustand / Pinia | 轻量状态管理 |
| HTTP 客户端 | Axios | API 请求 |
| 实时通信 | WebSocket / SSE | 任务状态推送 |
| 构建工具 | Vite | 快速构建 |

### 10.3 项目结构

```
claude-pipeline-frontend/
├── src/
│   ├── main.jsx                 # 入口文件
│   ├── App.jsx                  # 根组件
│   ├── api/
│   │   ├── client.js           # Axios 实例
│   │   ├── skills.js           # 技能 API
│   │   ├── tasks.js            # 任务 API
│   │   └── websocket.js        # WebSocket 连接
│   ├── components/
│   │   ├── Layout/
│   │   │   ├── Header.jsx
│   │   │   ├── Sidebar.jsx
│   │   │   └── MainContent.jsx
│   │   ├── SkillCard/
│   │   │   ├── SkillCard.jsx
│   │   │   └── SkillCard.css
│   │   ├── TaskList/
│   │   │   ├── TaskList.jsx
│   │   │   └── TaskItem.jsx
│   │   ├── TaskDetail/
│   │   │   ├── TaskDetail.jsx
│   │   │   ├── OutputConsole.jsx
│   │   │   └── ResultViewer.jsx
│   │   └── CreateTask/
│   │       ├── CreateTaskModal.jsx
│   │       └── ParameterForm.jsx
│   ├── pages/
│   │   ├── Dashboard.jsx       # 仪表盘
│   │   ├── Skills.jsx          # 技能页
│   │   ├── Tasks.jsx           # 任务页
│   │   └── Settings.jsx        # 设置页
│   ├── hooks/
│   │   ├── useSkills.js       # 技能数据 Hook
│   │   ├── useTasks.js        # 任务数据 Hook
│   │   └── useWebSocket.js    # WebSocket Hook
│   ├── store/
│   │   ├── skillStore.js      # 技能状态
│   │   └── taskStore.js       # 任务状态
│   └── styles/
│       ├── global.css
│       └── variables.css
├── public/
├── index.html
├── vite.config.js
└── package.json
```

### 10.4 核心组件实现

#### 9.4.1 API 客户端 `src/api/client.js`

```javascript
import axios from 'axios';

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器
client.interceptors.response.use(
  (response) => response.data,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export default client;
```

#### 9.4.2 技能 API `src/api/skills.js`

```javascript
import client from './client';

export const skillsApi = {
  // 获取技能列表
  getList: () => client.get('/skills'),

  // 获取技能详情
  getDetail: (skillId) => client.get(`/skills/${skillId}`),

  // 同步技能
  sync: () => client.post('/skills/sync'),
};
```

#### 9.4.3 任务 API `src/api/tasks.js`

```javascript
import client from './client';

export const tasksApi = {
  // 创建任务
  create: (data) => client.post('/tasks', data),

  // 获取任务列表
  getList: (params) => client.get('/tasks', { params }),

  // 获取任务详情
  getDetail: (taskId) => client.get(`/tasks/${taskId}`),

  // 获取任务结果
  getResult: (taskId) => client.get(`/tasks/${taskId}/result`),

  // 取消任务
  cancel: (taskId) => client.delete(`/tasks/${taskId}`),

  // 获取服务状态
  getStatus: () => client.get('/status'),
};
```

#### 9.4.4 WebSocket 连接 `src/api/websocket.js`

```javascript
class WebSocketClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.listeners = new Map();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('WebSocket 连接成功');
      this.reconnectAttempts = 0;
      this.emit('connected');
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.emit(data.type, data.payload);
    };

    this.ws.onclose = () => {
      console.log('WebSocket 连接关闭');
      this.emit('disconnected');
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket 错误:', error);
      this.emit('error', error);
    };
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  off(event, callback) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      const index = callbacks.indexOf(callback);
      if (index > -1) callbacks.splice(index, 1);
    }
  }

  emit(event, data) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach((cb) => cb(data));
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`重连中... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, 2000 * this.reconnectAttempts);
    }
  }

  send(type, payload) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }
}

export const wsClient = new WebSocketClient(
  import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
);
```

#### 9.4.5 技能卡片组件 `src/components/SkillCard/SkillCard.jsx`

```jsx
import React from 'react';
import './SkillCard.css';

const SkillCard = ({ skill, onSelect }) => {
  const getCategoryIcon = (category) => {
    const icons = {
      quality: '🔍',
      devops: '🚀',
      testing: '🧪',
      development: '🛠️',
      documentation: '📄',
    };
    return icons[category] || '⚡';
  };

  return (
    <div className="skill-card" onClick={() => onSelect(skill)}>
      <div className="skill-card-header">
        <span className="skill-icon">{getCategoryIcon(skill.category)}</span>
        <h3 className="skill-name">{skill.name}</h3>
      </div>

      <p className="skill-description">{skill.description}</p>

      <div className="skill-meta">
        <span className="skill-version">v{skill.version}</span>
        <span className={`skill-status ${skill.enabled ? 'enabled' : 'disabled'}`}>
          {skill.enabled ? '可用' : '禁用'}
        </span>
      </div>

      <div className="skill-params">
        <span className="params-count">
          {skill.parameters?.length || 0} 个参数
        </span>
      </div>

      <button className="skill-select-btn">
        选择此技能
      </button>
    </div>
  );
};

export default SkillCard;
```

#### 9.4.6 创建任务弹窗 `src/components/CreateTask/CreateTaskModal.jsx`

```jsx
import React, { useState, useEffect } from 'react';
import { tasksApi } from '../../api/tasks';
import ParameterForm from './ParameterForm';

const CreateTaskModal = ({ skill, onClose, onSubmit }) => {
  const [parameters, setParameters] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // 初始化默认参数值
  useEffect(() => {
    const defaults = {};
    skill.parameters?.forEach((param) => {
      if (param.default !== undefined) {
        defaults[param.name] = param.default;
      }
    });
    setParameters(defaults);
  }, [skill]);

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);

    try {
      const task = await tasksApi.create({
        skill_id: skill.id,
        parameters,
        options: {
          timeout: skill.cli?.timeout || 600,
        },
      });

      onSubmit(task);
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || '创建任务失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2>创建任务: {skill.name}</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="skill-info">
            <p>{skill.description}</p>
          </div>

          <ParameterForm
            parameters={skill.parameters}
            values={parameters}
            onChange={setParameters}
          />

          {error && (
            <div className="error-message">{error}</div>
          )}
        </div>

        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>
            取消
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? '创建中...' : '创建任务'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateTaskModal;
```

#### 9.4.7 参数表单 `src/components/CreateTask/ParameterForm.jsx`

```jsx
import React from 'react';

const ParameterForm = ({ parameters, values, onChange }) => {
  const handleChange = (name, value) => {
    onChange({ ...values, [name]: value });
  };

  const renderField = (param) => {
    const { name, type, required, description, values: enumValues } = param;

    switch (type) {
      case 'string':
        return (
          <input
            type="text"
            value={values[name] || ''}
            onChange={(e) => handleChange(name, e.target.value)}
            placeholder={description}
            required={required}
          />
        );

      case 'enum':
        return (
          <select
            value={values[name] || ''}
            onChange={(e) => handleChange(name, e.target.value)}
            required={required}
          >
            <option value="">请选择</option>
            {enumValues?.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        );

      case 'boolean':
        return (
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={values[name] || false}
              onChange={(e) => handleChange(name, e.target.checked)}
            />
            {description}
          </label>
        );

      case 'number':
        return (
          <input
            type="number"
            value={values[name] || ''}
            onChange={(e) => handleChange(name, Number(e.target.value))}
            placeholder={description}
            required={required}
          />
        );

      default:
        return null;
    }
  };

  return (
    <div className="parameter-form">
      {parameters?.map((param) => (
        <div key={param.name} className="form-group">
          <label className="form-label">
            {param.name}
            {param.required && <span className="required">*</span>}
          </label>
          {renderField(param)}
          {param.description && (
            <small className="form-hint">{param.description}</small>
          )}
        </div>
      ))}
    </div>
  );
};

export default ParameterForm;
```

#### 9.4.8 任务列表组件 `src/components/TaskList/TaskList.jsx`

```jsx
import React, { useEffect, useState } from 'react';
import { tasksApi } from '../../api/tasks';
import TaskItem from './TaskItem';
import { wsClient } from '../../api/websocket';

const TaskList = ({ onSelectTask }) => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);

  // 加载任务列表
  useEffect(() => {
    loadTasks();
  }, []);

  // WebSocket 实时更新
  useEffect(() => {
    wsClient.on('task_update', (data) => {
      setTasks((prev) =>
        prev.map((t) => (t.task_id === data.task_id ? { ...t, ...data } : t))
      );
    });

    wsClient.on('task_created', (task) => {
      setTasks((prev) => [task, ...prev]);
    });

    return () => {
      wsClient.off('task_update');
      wsClient.off('task_created');
    };
  }, []);

  const loadTasks = async () => {
    try {
      const data = await tasksApi.getList();
      setTasks(data.tasks);
    } catch (err) {
      console.error('加载任务失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (taskId) => {
    try {
      await tasksApi.cancel(taskId);
      setTasks((prev) =>
        prev.map((t) => (t.task_id === taskId ? { ...t, status: 'cancelled' } : t))
      );
    } catch (err) {
      console.error('取消任务失败:', err);
    }
  };

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div className="task-list">
      <div className="task-list-header">
        <h2>任务列表</h2>
        <span className="task-count">{tasks.length} 个任务</span>
      </div>

      <div className="task-list-content">
        {tasks.length === 0 ? (
          <div className="empty-state">暂无任务</div>
        ) : (
          tasks.map((task) => (
            <TaskItem
              key={task.task_id}
              task={task}
              onView={() => onSelectTask(task)}
              onCancel={() => handleCancel(task.task_id)}
            />
          ))
        )}
      </div>
    </div>
  );
};

export default TaskList;
```

#### 9.4.9 实时输出控制台 `src/components/TaskDetail/OutputConsole.jsx`

```jsx
import React, { useEffect, useRef, useState } from 'react';
import { wsClient } from '../../api/websocket';

const OutputConsole = ({ taskId }) => {
  const [lines, setLines] = useState([]);
  const consoleRef = useRef(null);

  useEffect(() => {
    // 监听任务输出
    const handleOutput = (data) => {
      if (data.task_id === taskId) {
        setLines((prev) => [...prev, ...data.lines]);
      }
    };

    wsClient.on('task_output', handleOutput);

    // 订阅任务输出
    wsClient.send('subscribe', { task_id: taskId });

    return () => {
      wsClient.off('task_output', handleOutput);
      wsClient.send('unsubscribe', { task_id: taskId });
    };
  }, [taskId]);

  // 自动滚动到底部
  useEffect(() => {
    if (consoleRef.current) {
      consoleRef.current.scrollTop = consoleRef.current.scrollHeight;
    }
  }, [lines]);

  return (
    <div className="output-console" ref={consoleRef}>
      <div className="console-header">
        <span>实时输出</span>
        <button onClick={() => setLines([])}>清空</button>
      </div>
      <div className="console-body">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`console-line ${line.type || 'info'}`}
          >
            <span className="timestamp">[{line.timestamp}]</span>
            <span className="message">{line.text}</span>
          </div>
        ))}
        {lines.length === 0 && (
          <div className="console-placeholder">等待输出...</div>
        )}
      </div>
    </div>
  );
};

export default OutputConsole;
```

### 10.5 状态管理

#### 9.5.1 任务状态 Store `src/store/taskStore.js` (Zustand)

```javascript
import { create } from 'zustand';
import { tasksApi } from '../api/tasks';

export const useTaskStore = create((set, get) => ({
  tasks: [],
  currentTask: null,
  loading: false,
  error: null,

  // 加载任务列表
  fetchTasks: async () => {
    set({ loading: true, error: null });
    try {
      const data = await tasksApi.getList();
      set({ tasks: data.tasks, loading: false });
    } catch (err) {
      set({ error: err.message, loading: false });
    }
  },

  // 创建任务
  createTask: async (skillId, parameters) => {
    set({ loading: true, error: null });
    try {
      const task = await tasksApi.create({
        skill_id: skillId,
        parameters,
      });
      set((state) => ({
        tasks: [task, ...state.tasks],
        loading: false,
      }));
      return task;
    } catch (err) {
      set({ error: err.message, loading: false });
      throw err;
    }
  },

  // 更新任务状态 (WebSocket 回调)
  updateTask: (taskId, updates) => {
    set((state) => ({
      tasks: state.tasks.map((t) =>
        t.task_id === taskId ? { ...t, ...updates } : t
      ),
      currentTask:
        state.currentTask?.task_id === taskId
          ? { ...state.currentTask, ...updates }
          : state.currentTask,
    }));
  },

  // 设置当前任务
  setCurrentTask: (task) => {
    set({ currentTask: task });
  },

  // 取消任务
  cancelTask: async (taskId) => {
    try {
      await tasksApi.cancel(taskId);
      get().updateTask(taskId, { status: 'cancelled' });
    } catch (err) {
      set({ error: err.message });
    }
  },
}));
```

### 10.6 页面布局

#### 9.6.1 仪表盘页面 `src/pages/Dashboard.jsx`

```jsx
import React from 'react';
import SkillCard from '../components/SkillCard/SkillCard';
import TaskList from '../components/TaskList/TaskList';
import { useSkillStore } from '../store/skillStore';
import CreateTaskModal from '../components/CreateTask/CreateTaskModal';

const Dashboard = () => {
  const [selectedSkill, setSelectedSkill] = React.useState(null);
  const [showModal, setShowModal] = React.useState(false);
  const { skills } = useSkillStore();

  const handleSkillSelect = (skill) => {
    setSelectedSkill(skill);
    setShowModal(true);
  };

  const handleTaskCreated = (task) => {
    console.log('任务已创建:', task);
    // 可以跳转到任务详情页
  };

  return (
    <div className="dashboard">
      <aside className="sidebar">
        <h2>技能列表</h2>
        <div className="skills-grid">
          {skills.map((skill) => (
            <SkillCard
              key={skill.id}
              skill={skill}
              onSelect={handleSkillSelect}
            />
          ))}
        </div>
      </aside>

      <main className="main-content">
        <TaskList onSelectTask={(task) => console.log('查看任务:', task)} />
      </main>

      {showModal && selectedSkill && (
        <CreateTaskModal
          skill={selectedSkill}
          onClose={() => setShowModal(false)}
          onSubmit={handleTaskCreated}
        />
      )}
    </div>
  );
};

export default Dashboard;
```

### 10.7 Anthropic 风格样式

#### 9.7.1 设计规范

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Anthropic 设计语言                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  配色方案                                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Primary  │  │ Secondary│  │ Accent   │  │ Dark BG  │  │ Light BG │     │
│  │ #D97706  │  │ #78716C  │  │ #FB923C  │  │ #0C0A09  │  │ #FAFAF9  │     │
│  │ 暖橙色   │  │ 石灰色   │  │ 亮橙色   │  │ 深黑     │  │ 米白色   │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
│                                                                             │
│  字体                                                                        │
│  • 标题: Söhne / Inter (600 weight)                                        │
│  • 正文: Söhne / Inter (400 weight)                                        │
│  • 代码: JetBrains Mono / Fira Code                                        │
│                                                                             │
│  圆角: 12px - 20px (大卡片使用更大圆角)                                      │
│  阴影: 柔和、多层次、带有暖色调                                               │
│  动画: 流畅、subtle、200-300ms                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 9.7.2 全局样式 `src/styles/global.css`

```css
/* ============================================
   Anthropic Style Design System
   ============================================ */

:root {
  /* 品牌色 - Anthropic 暖橙色系 */
  --color-primary: #D97706;
  --color-primary-hover: #B45309;
  --color-primary-light: #FBBF24;
  --color-accent: #FB923C;

  /* 中性色 */
  --color-stone-50: #FAFAF9;
  --color-stone-100: #F5F5F4;
  --color-stone-200: #E7E5E4;
  --color-stone-300: #D6D3D1;
  --color-stone-400: #A8A29E;
  --color-stone-500: #78716C;
  --color-stone-600: #57534E;
  --color-stone-700: #44403C;
  --color-stone-800: #292524;
  --color-stone-900: #1C1917;
  --color-stone-950: #0C0A09;

  /* 功能色 */
  --color-success: #16A34A;
  --color-success-light: #22C55E;
  --color-warning: #CA8A04;
  --color-error: #DC2626;
  --color-info: #2563EB;

  /* 背景 */
  --bg-primary: #0C0A09;
  --bg-secondary: #1C1917;
  --bg-tertiary: #292524;
  --bg-card: #1C1917;
  --bg-elevated: #292524;

  /* 文字 */
  --text-primary: #FAFAF9;
  --text-secondary: #A8A29E;
  --text-tertiary: #78716C;
  --text-inverse: #0C0A09;

  /* 边框 */
  --border-subtle: rgba(255, 255, 255, 0.06);
  --border-default: rgba(255, 255, 255, 0.1);
  --border-strong: rgba(255, 255, 255, 0.15);

  /* 圆角 */
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 20px;
  --radius-full: 9999px;

  /* 阴影 */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.4);
  --shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.5);
  --shadow-glow: 0 0 40px rgba(217, 119, 6, 0.15);

  /* 动画 */
  --transition-fast: 150ms ease;
  --transition-normal: 200ms ease;
  --transition-slow: 300ms ease;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;
}

/* 浅色主题 */
[data-theme="light"] {
  --bg-primary: #FAFAF9;
  --bg-secondary: #F5F5F4;
  --bg-tertiary: #E7E5E4;
  --bg-card: #FFFFFF;
  --bg-elevated: #FFFFFF;
  --text-primary: #1C1917;
  --text-secondary: #57534E;
  --text-tertiary: #78716C;
  --border-subtle: rgba(0, 0, 0, 0.04);
  --border-default: rgba(0, 0, 0, 0.08);
  --border-strong: rgba(0, 0, 0, 0.12);
}

/* ============================================
   基础样式
   ============================================ */

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  line-height: 1.6;
  min-height: 100vh;
}

/* 标题字体 */
h1, h2, h3, h4, h5, h6 {
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.3;
}

h1 { font-size: 2.5rem; }
h2 { font-size: 1.875rem; }
h3 { font-size: 1.5rem; }
h4 { font-size: 1.25rem; }

/* ============================================
   布局组件
   ============================================ */

/* 主容器 */
.app-container {
  display: flex;
  min-height: 100vh;
}

/* 侧边栏 */
.sidebar {
  width: 320px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-subtle);
}

.sidebar-logo {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.sidebar-title {
  font-size: 1.125rem;
  font-weight: 600;
}

/* 主内容区 */
.main-content {
  flex: 1;
  padding: var(--spacing-xl);
  overflow-y: auto;
}

/* ============================================
   技能卡片 - Anthropic 风格
   ============================================ */

.skill-card {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  cursor: pointer;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.skill-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary), var(--color-accent));
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.skill-card:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
}

.skill-card:hover::before {
  opacity: 1;
}

.skill-card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.skill-icon {
  width: 44px;
  height: 44px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.skill-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.skill-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
  margin-bottom: var(--spacing-md);
}

.skill-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.skill-tag {
  display: inline-flex;
  align-items: center;
  padding: var(--spacing-xs) var(--spacing-sm);
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.skill-tag.active {
  background: rgba(217, 119, 6, 0.15);
  color: var(--color-primary);
}

/* ============================================
   任务卡片
   ============================================ */

.task-item {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  transition: all var(--transition-fast);
}

.task-item:hover {
  background: var(--bg-elevated);
  border-color: var(--border-default);
}

.task-status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.task-status-indicator.running {
  background: var(--color-success-light);
  animation: pulse 2s infinite;
}

.task-status-indicator.completed {
  background: var(--color-stone-500);
}

.task-status-indicator.failed {
  background: var(--color-error);
}

.task-status-indicator.pending {
  background: var(--color-stone-600);
  border: 2px solid var(--color-stone-500);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.task-info {
  flex: 1;
  min-width: 0;
}

.task-name {
  font-weight: 500;
  font-size: 0.9375rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.task-skill {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.task-time {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

/* ============================================
   状态徽章
   ============================================ */

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: var(--radius-full);
  font-size: 0.8125rem;
  font-weight: 500;
}

.status-badge.running {
  background: rgba(22, 163, 74, 0.12);
  color: var(--color-success-light);
}

.status-badge.completed {
  background: rgba(120, 113, 108, 0.12);
  color: var(--text-secondary);
}

.status-badge.failed {
  background: rgba(220, 38, 38, 0.12);
  color: var(--color-error);
}

.status-badge.pending {
  background: rgba(120, 113, 108, 0.08);
  color: var(--text-tertiary);
}

/* ============================================
   按钮 - Anthropic 风格
   ============================================ */

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: 10px 18px;
  border-radius: var(--radius-md);
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: none;
  outline: none;
}

.btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.btn-primary {
  background: linear-gradient(135deg, var(--color-primary), #EA580C);
  color: white;
  box-shadow: 0 2px 8px rgba(217, 119, 6, 0.3);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(217, 119, 6, 0.4);
}

.btn-primary:active {
  transform: translateY(0);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover {
  background: var(--bg-elevated);
  border-color: var(--border-strong);
}

.btn-ghost {
  background: transparent;
  color: var(--text-secondary);
}

.btn-ghost:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

/* ============================================
   输出控制台 - Claude 聊天风格
   ============================================ */

.output-console {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.console-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.console-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.console-body {
  padding: var(--spacing-md);
  max-height: 400px;
  overflow-y: auto;
}

.console-line {
  padding: var(--spacing-xs) 0;
  font-size: 0.8125rem;
  line-height: 1.6;
  color: var(--text-secondary);
}

.console-line.error {
  color: var(--color-error);
}

.console-line.success {
  color: var(--color-success-light);
}

.console-line.warning {
  color: var(--color-warning);
}

.console-timestamp {
  color: var(--text-tertiary);
  margin-right: var(--spacing-sm);
}

/* ============================================
   模态框 - Anthropic 风格
   ============================================ */

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn var(--transition-normal);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: slideUp var(--transition-normal);
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-subtle);
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
}

.modal-close {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.modal-close:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.modal-body {
  padding: var(--spacing-lg);
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  border-top: 1px solid var(--border-subtle);
}

/* ============================================
   表单元素
   ============================================ */

.form-group {
  margin-bottom: var(--spacing-lg);
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.form-label .required {
  color: var(--color-primary);
  margin-left: 2px;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 12px 14px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 0.9375rem;
  transition: all var(--transition-fast);
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(217, 119, 6, 0.15);
}

.form-input::placeholder {
  color: var(--text-tertiary);
}

.form-hint {
  display: block;
  margin-top: var(--spacing-xs);
  font-size: 0.8125rem;
  color: var(--text-tertiary);
}

/* Select 下拉框 */
.form-select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2378716C' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 40px;
}

/* ============================================
   进度条
   ============================================ */

.progress-bar {
  height: 4px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-primary), var(--color-accent));
  border-radius: var(--radius-full);
  transition: width var(--transition-normal);
}

/* ============================================
   空状态
   ============================================ */

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-2xl);
  text-align: center;
  color: var(--text-secondary);
}

.empty-state-icon {
  width: 64px;
  height: 64px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-bottom: var(--spacing-md);
}

.empty-state-text {
  font-size: 0.9375rem;
}

/* ============================================
   加载状态
   ============================================ */

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-default);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-dots {
  display: flex;
  gap: 6px;
}

.loading-dots span {
  width: 8px;
  height: 8px;
  background: var(--color-primary);
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out;
}

.loading-dots span:nth-child(1) { animation-delay: -0.32s; }
.loading-dots span:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

/* ============================================
   滚动条
   ============================================ */

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--border-strong);
  border-radius: var(--radius-full);
}

::-webkit-scrollbar-thumb:hover {
  background: var(--text-tertiary);
}

/* ============================================
   响应式
   ============================================ */

@media (max-width: 768px) {
  .app-container {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border-subtle);
    padding: var(--spacing-md);
  }

  .main-content {
    padding: var(--spacing-md);
  }

  .modal-content {
    max-width: none;
    margin: var(--spacing-md);
    border-radius: var(--radius-lg);
  }
}
```

#### 9.7.3 Logo SVG 组件 `src/components/Logo.jsx`

```jsx
import React from 'react';

const Logo = ({ size = 40 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 40 40"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#D97706" />
        <stop offset="100%" stopColor="#FB923C" />
      </linearGradient>
    </defs>
    <rect width="40" height="40" rx="10" fill="url(#logoGradient)" />
    <path
      d="M12 20C12 15.5817 15.5817 12 20 12V12C24.4183 12 28 15.5817 28 20V28H20C15.5817 28 12 24.4183 12 20V20Z"
      fill="white"
      fillOpacity="0.9"
    />
  </svg>
);

export default Logo;
```

#### 9.7.4 主题切换 `src/components/ThemeToggle.jsx`

```jsx
import React, { useState, useEffect } from 'react';

const ThemeToggle = () => {
  const [theme, setTheme] = useState('dark');

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark');
  };

  return (
    <button
      className="btn btn-ghost"
      onClick={toggleTheme}
      aria-label="切换主题"
    >
      {theme === 'dark' ? (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="5" />
          <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
        </svg>
      ) : (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
        </svg>
      )}
    </button>
  );
};

export default ThemeToggle;
```

#### 9.7.5 消息气泡组件 `src/components/MessageBubble.jsx` (Claude 风格)

```jsx
import React from 'react';
import './MessageBubble.css';

const MessageBubble = ({ type = 'assistant', children }) => {
  return (
    <div className={`message-bubble message-${type}`}>
      <div className="message-avatar">
        {type === 'assistant' ? (
          <Logo size={28} />
        ) : (
          <div className="user-avatar">U</div>
        )}
      </div>
      <div className="message-content">
        {children}
      </div>
    </div>
  );
};

export default MessageBubble;
```

```css
/* MessageBubble.css */
.message-bubble {
  display: flex;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  border-radius: var(--radius-lg);
}

.message-assistant {
  background: var(--bg-secondary);
}

.message-user {
  background: var(--bg-primary);
}

.message-avatar {
  flex-shrink: 0;
}

.user-avatar {
  width: 28px;
  height: 28px;
  background: var(--color-primary);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
  color: white;
}

.message-content {
  flex: 1;
  min-width: 0;
  line-height: 1.7;
}

.message-content p {
  margin-bottom: var(--spacing-md);
}

.message-content p:last-child {
  margin-bottom: 0;
}

.message-content code {
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.875em;
}

.message-content pre {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  overflow-x: auto;
  margin: var(--spacing-md) 0;
}

.message-content pre code {
  background: transparent;
  padding: 0;
}
```

### 10.8 前端构建配置 `vite.config.js`

```javascript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
});
```

---

## 十一、总结

本服务提供统一的 API 接口，用户可以：

1. **查看可用技能** - 通过 `/api/skills` 获取所有可用的技能列表
2. **选择技能执行** - 通过 `/api/tasks` 提交任务，指定 `skill_id` 和参数
3. **追踪任务状态** - 通过 `/api/tasks/{id}` 查询执行进度
4. **获取执行结果** - 通过 `/api/tasks/{id}/result` 获取输出

所有 CLI 实例运行在同一个服务中，通过技能配置区分不同的执行逻辑。技能存储在 GitLab 仓库中，支持动态更新和版本管理。