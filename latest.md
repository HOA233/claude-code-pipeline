# Claude Code Agent 编排平台优化目标

## 当前问题

现有系统设计为 CI/CD 流水线模式，不够灵活，无法满足以下需求：
- 定制任意业务需求
- 自由组合不同的 Claude Code Agent
- 通过简单 API 调用获取结果

## 目标

将系统优化为 **Claude Code Agent 编排平台**：

1. **Agent 配置化**：每个 Agent 可独立配置（模型、提示词、工具、权限等）
2. **工作流编排**：灵活组合多个 Agent，支持串行、并行、条件执行
3. **API 驱动**：通过 REST API 触发任务，获取结构化结果
4. **业务无关**：不局限于 CI/CD，支持任何业务场景

## 核心概念变更

| 原概念 | 新概念 | 说明 |
|--------|--------|------|
| Skill | Agent | 配置化的 Claude Code 实例 |
| Pipeline | Workflow | Agent 编排工作流 |
| Step | AgentNode | 工作流中的 Agent 节点 |
| Run | Execution | 工作流执行实例 |

## 主要改动

### 1. 模型层
- 新增 `Agent` 模型，支持完整的 Claude Code CLI 配置
- 更新 `Workflow` 支持数据流和 Agent 组合
- 增强 `AgentNode` 支持动态输入映射

### 2. 服务层
- 新增 `AgentService` 管理 Agent 配置
- 新增 `WorkflowService` 处理工作流逻辑
- 更新 `Orchestrator` 支持 Agent 间数据传递

### 3. API 层
- `/api/agents` - Agent CRUD 和直接执行
- `/api/workflows` - 工作流管理
- `/api/executions` - 执行触发和结果获取

## 典型用例

```
用户请求 -> API -> Workflow -> Agent1 -> Agent2 -> Agent3 -> 结构化结果
```

示例：
- 代码分析 → 安全扫描 → 生成报告
- 需求文档 → 代码生成 → 测试生成 → 代码审查
- 日志分析 → 问题诊断 → 修复建议

---

详细设计见 [TARGET.md](./TARGET.md)