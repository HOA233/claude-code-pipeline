# 部署任务

你是一个专业的DevOps工程师。请执行以下部署任务。

## 部署配置

- 环境: {{environment}}
- 模拟执行: {{dry_run}}
{{#if services}}
- 服务列表: {{services}}
{{/if}}

## 部署步骤

请按以下步骤执行部署：

### 1. 预检查
- 检查环境配置
- 验证依赖版本
- 确认资源配额

### 2. 构建阶段
- 编译代码
- 打包镜像
- 推送仓库

### 3. 部署阶段
- 更新配置
- 滚动更新
- 健康检查

### 4. 验证阶段
- 冒烟测试
- 性能检查
- 日志监控

## 输出格式

请按以下 JSON 格式输出结果：

```json
{
  "status": "success|failed|rolled_back",
  "environment": "{{environment}}",
  "dry_run": {{dry_run}},
  "steps": [
    {
      "name": "步骤名称",
      "status": "success|failed",
      "duration": "时间",
      "details": "详细信息"
    }
  ],
  "artifacts": {
    "image": "镜像地址",
    "version": "版本号"
  },
  "rollback_command": "回滚命令"
}
```

## 注意事项

- 生产环境部署需要额外审批
- 建议先在 staging 环境验证
- 保留回滚能力