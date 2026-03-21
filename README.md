# Claude CLI Orchestration Service

一个 CLI 编排服务，可以将多个 CLI 实例组合成流水线协同工作。

## 核心概念

### 什么是 CLI Pipeline？

CLI Pipeline 允许将多个 CLI 命令串联或并联执行，形成一个完整的工作流：

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLI Pipeline 编排                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐     │
│  │ CLI #1  │───▶│ CLI #2  │───▶│ CLI #3  │───▶│ Result  │     │
│  │代码分析  │    │生成测试  │    │运行测试  │    │  输出   │     │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘     │
│                                                                 │
│  或并行执行:                                                     │
│                                                                 │
│  ┌─────────┐                                                   │
│  │ CLI #1  │──┐                                                │
│  └─────────┘  │    ┌─────────┐                                │
│  ┌─────────┐  ├───▶│ 合并结果 │                                │
│  │ CLI #2  │──┘    └─────────┘                                │
│  └─────────┘                                                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 主要功能

1. **CLI 编排** - 将多个 CLI 命令组合成流水线
2. **串行/并行** - 支持串行和并行执行模式
3. **数据传递** - CLI 之间可以传递数据
4. **状态管理** - 实时跟踪每个 CLI 的执行状态
5. **错误处理** - 支持重试、回滚等策略

## 快速开始

```bash
# 启动服务
docker-compose up -d

# 创建一个 CLI Pipeline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-quality-check",
    "steps": [
      {"cli": "claude", "action": "analyze", "params": {"target": "src/"}},
      {"cli": "claude", "action": "test-gen", "params": {"source": "src/"}},
      {"cli": "claude", "action": "review", "params": {"depth": "deep"}}
    ]
  }'
```

## API 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/pipelines` | GET/POST | 管道列表/创建 |
| `/api/pipelines/:id` | GET/PUT/DELETE | 管道详情/更新/删除 |
| `/api/pipelines/:id/run` | POST | 执行管道 |
| `/api/pipelines/:id/status` | GET | 获取执行状态 |
| `/api/runs` | GET | 所有执行记录 |
| `/api/runs/:id` | GET | 执行详情 |

## 实时更新

服务支持 WebSocket 和 Server-Sent Events (SSE) 用于实时状态更新。

### WebSocket 端点

| 端点 | 说明 |
|------|------|
| `/ws` | 全局系统更新 |
| `/ws/tasks/:id` | 任务状态更新 |
| `/ws/runs/:id` | 执行状态更新 |

### SSE 端点

| 端点 | 说明 |
|------|------|
| `/sse` | 全局系统统计 (每5秒) |
| `/sse/tasks/:id` | 任务状态更新 |
| `/sse/runs/:id` | 执行状态更新 |

### 使用示例

```javascript
// WebSocket 连接
const ws = new WebSocket('ws://localhost:8080/ws/tasks/task-123');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Task update:', data);
};

// SSE 连接
const eventSource = new EventSource('http://localhost:8080/sse/tasks/task-123');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Task update:', data);
};
```

## Pipeline 配置示例

```yaml
name: full-code-review
description: 完整的代码审查流程

# 执行模式: serial(串行) / parallel(并行)
mode: serial

# 步骤定义
steps:
  - id: analyze
    cli: claude
    action: analyze
    params:
      target: src/
    on_error: continue  # continue / stop / retry

  - id: security
    cli: claude
    action: security-scan
    params:
      target: src/
    depends_on: [analyze]  # 依赖关系

  - id: test-gen
    cli: claude
    action: test-gen
    params:
      source: "{{analyze.output.files}}"  # 引用上一步输出
    depends_on: [analyze]

  - id: run-tests
    cli: npm
    command: test
    depends_on: [test-gen]

# 并行组
parallel_groups:
  - steps: [security, test-gen]  # 这两个步骤并行执行
    wait_all: true  # 等待全部完成

# 错误处理
error_handling:
  retry: 2
  on_failure: notify  # notify / rollback
  webhook: https://hooks.example.com/failure

# 输出配置
output:
  format: json
  merge_strategy: deep  # deep / shallow
```

## 项目结构

```
claude-cli-orchestration/
├── cmd/server/           # 服务入口
├── internal/
│   ├── api/              # HTTP API
│   ├── orchestrator/     # 编排引擎
│   ├── executor/         # CLI 执行器
│   ├── pipeline/         # 管道定义
│   ├── step/             # 步骤执行
│   └── storage/          # 存储层 (Redis)
├── pkg/
│   ├── cli/              # CLI 客户端
│   └── logger/           # 日志
├── config/               # 配置文件
├── examples/             # 示例 Pipeline
└── frontend/             # Web UI
```