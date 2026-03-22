# Claude Code Agent 编排平台使用示例

## 快速开始

### 1. 启动服务

```bash
# 启动 Redis
redis-server

# 启动服务
go run cmd/server/main.go
```

### 2. 创建 Agent

```bash
# 创建代码审查 Agent
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-reviewer",
    "description": "专业的代码审查助手",
    "model": "claude-sonnet-4-6",
    "system_prompt": "你是一个专业的代码审查助手，帮助开发者发现代码中的问题和改进空间。",
    "max_tokens": 4096,
    "tools": [
      {"name": "read_file", "description": "读取文件内容"},
      {"name": "write_file", "description": "写入文件内容"}
    ],
    "permissions": [
      {"resource": "file", "action": "read"},
      {"resource": "file", "action": "write"}
    ],
    "timeout": 300,
    "isolation": {
      "data_isolation": true,
      "session_isolation": true
    },
    "tags": ["code-review", "quality"],
    "category": "analysis"
  }'
```

### 3. 执行 Agent

```bash
# 直接执行 Agent
curl -X POST http://localhost:8080/api/agents/{agent_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "target": "src/main.go",
      "language": "go"
    },
    "async": true
  }'
```

---

## 完整示例

### 示例 1：创建代码审查 + 测试生成工作流

```bash
# 1. 创建代码审查 Agent
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-reviewer",
    "description": "代码审查",
    "model": "claude-sonnet-4-6",
    "system_prompt": "审查代码并提供改进建议",
    "tools": [{"name": "read_file", "description": "读取文件"}],
    "permissions": [{"resource": "file", "action": "read"}]
  }'

# 2. 创建测试生成 Agent
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-generator",
    "description": "生成测试代码",
    "model": "claude-sonnet-4-6",
    "system_prompt": "根据代码生成单元测试",
    "tools": [
      {"name": "read_file", "description": "读取文件"},
      {"name": "write_file", "description": "写入文件"}
    ],
    "permissions": [
      {"resource": "file", "action": "read"},
      {"resource": "file", "action": "write"}
    ]
  }'

# 3. 创建工作流
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "review-and-test",
    "description": "代码审查并生成测试",
    "agents": [
      {
        "id": "reviewer",
        "agent_id": "{code-reviewer-id}",
        "input": {"target": "{{input.target}}"},
        "output_as": "review_result"
      },
      {
        "id": "test-gen",
        "agent_id": "{test-generator-id}",
        "input_from": {"code": "review_result.code"},
        "output_as": "tests",
        "depends_on": ["reviewer"]
      }
    ],
    "mode": "serial"
  }'

# 4. 执行工作流
curl -X POST http://localhost:8080/api/executions \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "{workflow-id}",
    "input": {"target": "src/"},
    "async": true
  }'
```

### 示例 2：创建定时安全扫描

```bash
# 1. 创建安全扫描 Agent
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "security-scanner",
    "description": "安全漏洞扫描",
    "model": "claude-sonnet-4-6",
    "system_prompt": "扫描代码中的安全漏洞",
    "tools": [{"name": "read_file", "description": "读取文件"}],
    "tags": ["security", "scan"]
  }'

# 2. 创建定时任务
curl -X POST http://localhost:8080/api/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "daily-security-scan",
    "description": "每日安全扫描",
    "target_type": "agent",
    "target_id": "{security-scanner-id}",
    "cron": "0 2 * * *",
    "timezone": "Asia/Shanghai",
    "input": {"target": "src/"},
    "on_failure": "notify",
    "notify_email": "security@example.com"
  }'
```

### 示例 3：多 Agent 并行分析

```bash
# 创建并行分析工作流
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "comprehensive-analysis",
    "description": "全面的代码分析",
    "agents": [
      {
        "id": "security",
        "agent_id": "{security-agent-id}",
        "input": {"target": "{{input.target}}"},
        "output_as": "security_report"
      },
      {
        "id": "performance",
        "agent_id": "{performance-agent-id}",
        "input": {"target": "{{input.target}}"},
        "output_as": "performance_report"
      },
      {
        "id": "quality",
        "agent_id": "{quality-agent-id}",
        "input": {"target": "{{input.target}}"},
        "output_as": "quality_report"
      },
      {
        "id": "summarizer",
        "agent_id": "{summarizer-agent-id}",
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
  }'
```

---

## 前端集成示例

### React 组件使用

```tsx
import { ExecutionList } from './components/ExecutionList';
import api from './api/client';

function App() {
  const handleCreateAgent = async () => {
    const agent = await api.createAgent({
      name: 'my-agent',
      description: 'My custom agent',
      model: 'claude-sonnet-4-6',
    });
    console.log('Created agent:', agent.id);
  };

  const handleExecute = async (workflowId: string) => {
    const execution = await api.executeWorkflow({
      workflow_id: workflowId,
      input: { target: 'src/' },
      async: true,
    });
    console.log('Started execution:', execution.id);
  };

  return (
    <div>
      <button onClick={handleCreateAgent}>Create Agent</button>
      <ExecutionList autoRefresh={true} />
    </div>
  );
}
```

### 实时更新订阅

```tsx
// 订阅单个执行
useEffect(() => {
  const unsubscribe = api.subscribeExecution(executionId, (data) => {
    console.log('Update:', data.status, data.progress);
    setExecution(data);
  });

  return unsubscribe;
}, [executionId]);

// 订阅所有执行
useEffect(() => {
  const unsubscribe = api.subscribeAllExecutions((data) => {
    // 更新执行列表
    updateExecutionInList(data);
  });

  return unsubscribe;
}, []);
```

---

## API 完整列表

### Agent API
| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/agents | 创建 Agent |
| GET | /api/agents | 列出 Agents |
| GET | /api/agents/:id | 获取 Agent |
| PUT | /api/agents/:id | 更新 Agent |
| DELETE | /api/agents/:id | 删除 Agent |
| POST | /api/agents/:id/test | 测试 Agent |
| POST | /api/agents/:id/execute | 执行 Agent |

### Workflow API
| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/workflows | 创建 Workflow |
| GET | /api/workflows | 列出 Workflows |
| GET | /api/workflows/:id | 获取 Workflow |
| PUT | /api/workflows/:id | 更新 Workflow |
| DELETE | /api/workflows/:id | 删除 Workflow |

### Execution API
| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/executions | 执行 Workflow |
| GET | /api/executions | 列出 Executions |
| GET | /api/executions/:id | 获取 Execution |
| POST | /api/executions/:id/cancel | 取消 Execution |
| POST | /api/executions/:id/pause | 暂停 Execution |
| POST | /api/executions/:id/resume | 恢复 Execution |

### Scheduled Job API
| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/schedules | 创建定时任务 |
| GET | /api/schedules | 列出定时任务 |
| GET | /api/schedules/:id | 获取定时任务 |
| PUT | /api/schedules/:id | 更新定时任务 |
| DELETE | /api/schedules/:id | 删除定时任务 |
| POST | /api/schedules/:id/enable | 启用任务 |
| POST | /api/schedules/:id/disable | 禁用任务 |
| POST | /api/schedules/:id/trigger | 手动触发 |
| GET | /api/schedules/:id/history | 执行历史 |

---

## 常见用例

### 1. 代码审查流水线
```
代码提交 -> 触发 Webhook -> 执行代码审查 Agent -> 发送审查结果
```

### 2. 自动化测试生成
```
代码变更 -> 触发工作流 -> 分析代码 -> 生成测试 -> 运行测试 -> 报告结果
```

### 3. 定期安全扫描
```
定时任务(每日凌晨2点) -> 执行安全扫描 Agent -> 发现漏洞 -> 发送通知
```

### 4. 文档自动生成
```
代码更新 -> 触发工作流 -> 分析代码 -> 生成文档 -> 提交文档
```

---

## 最佳实践

1. **Agent 设计**
   - 单一职责：每个 Agent 只做一件事
   - 明确输入输出：定义清晰的 Schema
   - 合理超时：根据任务复杂度设置 timeout

2. **Workflow 设计**
   - 合理划分步骤：避免过深依赖链
   - 选择合适模式：简单流程用 serial，独立任务用 parallel
   - 错误处理：配置合适的错误策略

3. **资源隔离**
   - 多租户环境启用数据隔离
   - 并发执行启用会话隔离
   - 敏感操作启用网络隔离

4. **定时任务**
   - 合理设置 Cron 表达式
   - 配置失败通知
   - 监控执行历史