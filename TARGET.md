# Claude Code Agent 编排平台

## 目标架构

### 愿景
将当前的 CI/CD 流水线系统转变为灵活的 **Claude Code Agent 编排平台**，让用户能够：
1. 配置和组合多个 Claude Code Agent
2. 通过 API 触发 Agent 任务
3. 定制任何业务需求（不只是 CI/CD）
4. 从 Agent 执行中获取结构化结果

---

## 核心概念

### 1. Agent（取代 Skill）
**Agent** 是一个配置好的 Claude Code CLI 实例，具有特定能力：

```go
type Agent struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`

    // Claude Code CLI 配置
    Model       string            `json:"model"`           // claude-sonnet-4-6, claude-opus-4-6 等
    SystemPrompt string           `json:"system_prompt"`   // Agent 的行为/指令
    MaxTokens   int               `json:"max_tokens"`

    // 技能选择 - Agent 可选择不同的 Skill
    Skills      []SkillRef        `json:"skills"`          // 该 Agent 可调用的技能列表
    DefaultSkill string           `json:"default_skill,omitempty"` // 默认技能

    // 能力配置
    Tools       []Tool            `json:"tools"`           // 该 Agent 可用的工具
    Permissions []Permission      `json:"permissions"`     // 该 Agent 的权限

    // 输入/输出 Schema
    InputSchema  json.RawMessage  `json:"input_schema"`    // 期望的输入格式
    OutputSchema json.RawMessage  `json:"output_schema"`   // 期望的输出格式

    // 行为配置
    Timeout     int               `json:"timeout"`
    RetryPolicy RetryPolicy       `json:"retry_policy"`

    // 隔离配置
    Isolation   IsolationConfig   `json:"isolation"`       // 数据/会话隔离配置

    // 元数据
    Tags        []string          `json:"tags"`
    Category    string            `json:"category"`
    Version     string            `json:"version"`
}

// SkillRef 定义 Agent 可引用的技能
type SkillRef struct {
    SkillID     string            `json:"skill_id"`
    Alias       string            `json:"alias,omitempty"` // 在 Agent 内的别名
    InputMapping map[string]string `json:"input_mapping,omitempty"` // 输入映射
    OutputMapping map[string]string `json:"output_mapping,omitempty"` // 输出映射
}

// IsolationConfig 隔离配置
type IsolationConfig struct {
    DataIsolation    bool   `json:"data_isolation"`     // 数据隔离
    SessionIsolation bool   `json:"session_isolation"`  // 会话隔离
    NetworkIsolation bool   `json:"network_isolation"`  // 网络隔离
    FileIsolation    bool   `json:"file_isolation"`     // 文件系统隔离
    Namespace        string `json:"namespace,omitempty"` // 隔离命名空间
}
```

### 2. Workflow（取代 Pipeline）
**Workflow** 定义多个 Agent 如何协同工作：

```go
type Workflow struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`

    // Agent 组合
    Agents      []AgentNode       `json:"agents"`

    // 执行流程
    Connections []Connection      `json:"connections"`     // Agent 之间的数据流
    Mode        ExecutionMode     `json:"mode"`            // serial, parallel, conditional

    // 隔离配置
    SessionID   string            `json:"session_id"`
    TenantID    string            `json:"tenant_id"`

    // 配置
    Context     map[string]interface{} `json:"context"`
    ErrorHandling ErrorConfig     `json:"error_handling"`
}
```

### 3. AgentNode（取代 Step）
**AgentNode** 代表 Agent 在工作流中的角色：

```go
type AgentNode struct {
    ID          string            `json:"id"`
    AgentID     string            `json:"agent_id"`        // 引用 Agent

    // 输入配置
    Input       map[string]interface{} `json:"input"`       // 静态输入
    InputFrom   map[string]string      `json:"input_from"`  // 来自其他节点的动态输入

    // 输出配置
    OutputAs    string            `json:"output_as"`       // 输出的变量名

    // 执行控制
    DependsOn   []string          `json:"depends_on"`
    Condition   string            `json:"condition"`       // 条件执行
    Timeout     int               `json:"timeout"`
    OnError     ErrorStrategy     `json:"on_error"`
}
```

### 4. Execution（取代 Run）
**Execution** 是工作流的执行实例：

```go
type Execution struct {
    ID          string            `json:"id"`
    WorkflowID  string            `json:"workflow_id"`
    SessionID   string            `json:"session_id"`

    Status      ExecutionStatus   `json:"status"`

    // 结果
    NodeResults map[string]NodeResult `json:"node_results"`
    FinalOutput json.RawMessage       `json:"final_output"`

    // 元数据
    Duration    int64             `json:"duration"`
    Error       string            `json:"error,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
    CompletedAt *time.Time        `json:"completed_at,omitempty"`
}
```

---

## API 设计

### 1. Agent 管理 API

```
POST   /api/agents              # 创建新 Agent
GET    /api/agents              # 列出所有 Agent
GET    /api/agents/:id          # 获取 Agent 详情
PUT    /api/agents/:id          # 更新 Agent
DELETE /api/agents/:id          # 删除 Agent
POST   /api/agents/:id/test     # 用示例输入测试 Agent
POST   /api/agents/:id/execute  # 直接执行单个 Agent
```

### 2. Workflow 管理 API

```
POST   /api/workflows           # 创建工作流
GET    /api/workflows           # 列出工作流
GET    /api/workflows/:id       # 获取工作流详情
PUT    /api/workflows/:id       # 更新工作流
DELETE /api/workflows/:id       # 删除工作流
```

### 3. Execution 执行 API

```
POST   /api/executions          # 触发工作流执行
GET    /api/executions/:id      # 获取执行状态/结果
POST   /api/executions/:id/cancel # 取消执行
GET    /api/executions/:id/stream # SSE 实时更新流
```

---

## 技能选择与隔离机制

### Agent 技能选择
每个 Agent 可以选择多个 Skill，实现能力组合：

```json
{
  "id": "code-analyzer",
  "name": "代码分析器",
  "skills": [
    {
      "skill_id": "static-analysis",
      "alias": "linter",
      "input_mapping": {
        "target": "code_path"
      }
    },
    {
      "skill_id": "security-scan",
      "alias": "scanner"
    }
  ],
  "default_skill": "linter"
}
```

### 隔离级别

| 隔离类型 | 说明 | 使用场景 |
|---------|------|---------|
| 数据隔离 | Agent 执行数据独立存储 | 多租户环境 |
| 会话隔离 | 每个 Agent 有独立会话 | 并发执行 |
| 网络隔离 | Agent 网络访问受限 | 安全敏感任务 |
| 文件隔离 | 独立文件系统命名空间 | 代码修改任务 |

### 隔离配置示例

```json
{
  "isolation": {
    "data_isolation": true,
    "session_isolation": true,
    "network_isolation": false,
    "file_isolation": true,
    "namespace": "tenant-123"
  }
}
```

---

## 示例工作流

### 示例 1：代码审查 + 测试生成

```json
{
  "name": "code-review-and-test",
  "description": "审查代码并生成测试",
  "agents": [
    {
      "id": "reviewer",
      "agent_id": "code-reviewer",
      "input": { "target": "{{input.target}}" },
      "output_as": "review_result"
    },
    {
      "id": "test-gen",
      "agent_id": "test-generator",
      "input_from": { "code": "review_result.code" },
      "output_as": "tests",
      "depends_on": ["reviewer"]
    }
  ],
  "mode": "serial"
}
```

### 示例 2：多 Agent 并行分析

```json
{
  "name": "comprehensive-analysis",
  "description": "并行运行多个分析 Agent",
  "agents": [
    {
      "id": "security",
      "agent_id": "security-scanner",
      "input": { "target": "{{input.target}}" },
      "output_as": "security_report"
    },
    {
      "id": "performance",
      "agent_id": "performance-analyzer",
      "input": { "target": "{{input.target}}" },
      "output_as": "performance_report"
    },
    {
      "id": "quality",
      "agent_id": "code-quality-checker",
      "input": { "target": "{{input.target}}" },
      "output_as": "quality_report"
    },
    {
      "id": "summarizer",
      "agent_id": "report-summarizer",
      "input_from": {
        "security": "security_report",
        "performance": "performance_report",
        "quality": "quality_report"
      },
      "output_as": "final_report",
      "depends_on": ["security", "performance", "quality"]
    }
  ],
  "mode": "hybrid"
}
```

### 示例 3：交互式 Agent 链

```json
{
  "name": "interactive-refactor",
  "description": "迭代式代码改进",
  "agents": [
    {
      "id": "analyzer",
      "agent_id": "code-analyzer",
      "input": { "target": "{{input.target}}" },
      "output_as": "issues"
    },
    {
      "id": "planner",
      "agent_id": "refactor-planner",
      "input_from": { "issues": "issues" },
      "output_as": "plan",
      "depends_on": ["analyzer"]
    },
    {
      "id": "executor",
      "agent_id": "code-modifier",
      "input_from": { "plan": "plan" },
      "output_as": "changes",
      "depends_on": ["planner"]
    },
    {
      "id": "validator",
      "agent_id": "test-runner",
      "input_from": { "changes": "changes" },
      "output_as": "validation",
      "depends_on": ["executor"]
    }
  ],
  "mode": "serial"
}
```

---

## 迁移路径

### 第一阶段：模型重构
- [ ] 重命名 Skill -> Agent
- [ ] 重命名 Pipeline -> Workflow
- [ ] 重命名 Step -> AgentNode
- [ ] 重命名 Run -> Execution
- [ ] 更新所有模型结构

### 第二阶段：服务层
- [ ] 创建 AgentService
- [ ] 创建 WorkflowService
- [ ] 更新 Orchestrator 支持 Agent 组合
- [ ] 实现 Agent 之间的数据流

### 第三阶段：API 层
- [ ] 更新路由使用新命名
- [ ] 添加直接 Agent 执行端点
- [ ] 增强 SSE 流式传输

### 第四阶段：CLI 集成
- [ ] 增强 Claude Code CLI 参数支持
- [ ] 支持工具配置
- [ ] 支持权限系统

### 第五阶段：定时调度
- [ ] 创建 ScheduledJob 模型
- [ ] 创建 SchedulerService
- [ ] 实现 Cron 表达式解析
- [ ] 添加定时任务 API
- [ ] 支持时区配置

### 第六阶段：测试与文档
- [ ] 更新所有测试
- [ ] 添加 API 文档
- [ ] 创建使用示例

---

## 定时调度（Cron）

### 5. ScheduledJob 定时任务
支持定时执行 Agent 或 Workflow：

```go
type ScheduledJob struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`

    // 执行目标
    TargetType  string            `json:"target_type"`     // "agent" 或 "workflow"
    TargetID    string            `json:"target_id"`

    // 调度配置
    Cron        string            `json:"cron"`            // Cron 表达式 "*/10 * * * *"
    Timezone    string            `json:"timezone"`        // 时区，默认本地
    Enabled     bool              `json:"enabled"`

    // 执行参数
    Input       map[string]interface{} `json:"input"`

    // 执行历史
    LastRun     *time.Time        `json:"last_run,omitempty"`
    NextRun     *time.Time        `json:"next_run,omitempty"`
    RunCount    int               `json:"run_count"`

    // 失败处理
    OnFailure   string            `json:"on_failure"`      // "notify", "retry", "disable"
    NotifyEmail string            `json:"notify_email,omitempty"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    TenantID    string            `json:"tenant_id,omitempty"`
}
```

### 定时任务 API

```
POST   /api/schedules              # 创建定时任务
GET    /api/schedules              # 列出定时任务
GET    /api/schedules/:id          # 获取定时任务详情
PUT    /api/schedules/:id          # 更新定时任务
DELETE /api/schedules/:id          # 删除定时任务
POST   /api/schedules/:id/enable   # 启用定时任务
POST   /api/schedules/:id/disable  # 禁用定时任务
GET    /api/schedules/:id/history  # 执行历史
```

### 示例：定时代码扫描

```json
{
  "name": "daily-security-scan",
  "description": "每日安全扫描",
  "target_type": "workflow",
  "target_id": "security-scan-workflow",
  "cron": "0 2 * * *",
  "timezone": "Asia/Shanghai",
  "input": {
    "target": "src/"
  },
  "on_failure": "notify",
  "notify_email": "security@example.com"
}
```

---

## 核心优势

1. **灵活性**：支持任何业务工作流，不只是 CI/CD
2. **可组合性**：自由组合不同的 Agent
3. **简洁性**：通过 API 直接触发 Agent 任务
4. **隔离性**：基于会话的数据隔离
5. **可扩展性**：支持并行 Agent 执行
6. **可观测性**：实时执行监控
7. **定时调度**：支持 Cron 表达式定时执行任务