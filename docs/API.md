# Claude Code Agent 编排平台 API 文档

## 概述

本文档描述 Agent 编排平台的所有 API 端点。

## 认证

所有 API 请求需要在 Header 中包含认证信息：
```
Authorization: Bearer <token>
```

---

## Agent API

### 创建 Agent

```
POST /api/agents
```

**请求体：**
```json
{
  "name": "code-reviewer",
  "description": "代码审查 Agent",
  "model": "claude-sonnet-4-6",
  "system_prompt": "你是一个专业的代码审查助手...",
  "max_tokens": 4096,
  "skills": [
    {
      "skill_id": "static-analysis",
      "alias": "linter"
    }
  ],
  "tools": [
    {"name": "read_file", "description": "读取文件"}
  ],
  "permissions": [
    {"resource": "file", "action": "read"}
  ],
  "timeout": 300,
  "isolation": {
    "data_isolation": true,
    "session_isolation": true
  },
  "tags": ["code-review", "quality"],
  "category": "analysis"
}
```

**响应：** `201 Created`
```json
{
  "id": "agent-xxx",
  "name": "code-reviewer",
  "enabled": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

### 列出 Agents

```
GET /api/agents?tenant_id=xxx&category=analysis
```

**响应：**
```json
{
  "agents": [...],
  "total": 10
}
```

### 获取 Agent

```
GET /api/agents/:id
```

### 更新 Agent

```
PUT /api/agents/:id
```

### 删除 Agent

```
DELETE /api/agents/:id
```

### 测试 Agent

```
POST /api/agents/:id/test
```

**请求体：**
```json
{
  "target": "src/main.go"
}
```

### 执行 Agent

```
POST /api/agents/:id/execute
```

**请求体：**
```json
{
  "input": {
    "target": "src/"
  },
  "async": true,
  "callback": "https://example.com/webhook"
}
```

---

## Workflow API

### 创建 Workflow

```
POST /api/workflows
```

**请求体：**
```json
{
  "name": "code-review-and-test",
  "description": "代码审查并生成测试",
  "agents": [
    {
      "id": "reviewer",
      "agent_id": "agent-xxx",
      "input": {"target": "{{input.target}}"},
      "output_as": "review_result"
    },
    {
      "id": "test-gen",
      "agent_id": "agent-yyy",
      "input_from": {"code": "review_result.code"},
      "output_as": "tests",
      "depends_on": ["reviewer"]
    }
  ],
  "mode": "serial"
}
```

### 列出 Workflows

```
GET /api/workflows?tenant_id=xxx
```

### 获取 Workflow

```
GET /api/workflows/:id
```

### 更新 Workflow

```
PUT /api/workflows/:id
```

### 删除 Workflow

```
DELETE /api/workflows/:id
```

---

## Execution API

### 执行 Workflow

```
POST /api/executions
```

**请求体：**
```json
{
  "workflow_id": "workflow-xxx",
  "input": {
    "target": "src/"
  },
  "async": true
}
```

**响应：** `202 Accepted`
```json
{
  "id": "exec-xxx",
  "workflow_id": "workflow-xxx",
  "status": "pending",
  "progress": 0,
  "created_at": "2024-01-15T10:00:00Z"
}
```

### 列出 Executions

```
GET /api/executions?status=running&workflow_id=xxx&page=1&page_size=20
```

**响应：**
```json
{
  "executions": [
    {
      "id": "exec-xxx",
      "workflow_name": "code-review",
      "status": "running",
      "progress": 65,
      "current_step": "test-gen",
      "total_steps": 3,
      "completed_steps": 2,
      "duration": 45000,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

### 获取 Execution

```
GET /api/executions/:id
```

### 取消 Execution

```
POST /api/executions/:id/cancel
```

### 暂停 Execution

```
POST /api/executions/:id/pause
```

### 恢复 Execution

```
POST /api/executions/:id/resume
```

### 取消所有 Executions

```
POST /api/executions/cancel-all?status=running
```

---

## Scheduled Job API

### 创建定时任务

```
POST /api/schedules
```

**请求体：**
```json
{
  "name": "daily-security-scan",
  "description": "每日安全扫描",
  "target_type": "workflow",
  "target_id": "workflow-xxx",
  "cron": "0 2 * * *",
  "timezone": "Asia/Shanghai",
  "input": {
    "target": "src/"
  },
  "on_failure": "notify",
  "notify_email": "security@example.com"
}
```

### 列出定时任务

```
GET /api/schedules?tenant_id=xxx
```

### 获取定时任务

```
GET /api/schedules/:id
```

### 更新定时任务

```
PUT /api/schedules/:id
```

### 删除定时任务

```
DELETE /api/schedules/:id
```

### 启用定时任务

```
POST /api/schedules/:id/enable
```

### 禁用定时任务

```
POST /api/schedules/:id/disable
```

### 手动触发任务

```
POST /api/schedules/:id/trigger
```

### 获取执行历史

```
GET /api/schedules/:id/history?page=1&page_size=20
```

---

## SSE 实时更新

### 单个 Execution 更新流

```
GET /sse/executions/:id
```

**事件格式：**
```
event: execution_update
data: {"execution_id":"exec-xxx","status":"running","progress":45}
```

### 所有 Execution 更新流

```
GET /sse/executions
```

---

## WebSocket 实时通信

### 连接

```
ws://host/ws/executions
```

### 消息格式

```json
{
  "type": "subscribe",
  "execution_id": "exec-xxx"
}
```

```json
{
  "type": "execution_update",
  "data": {
    "execution_id": "exec-xxx",
    "status": "completed",
    "progress": 100
  }
}
```