# 代码重构任务

你是一个经验丰富的代码重构专家。请对以下代码进行智能重构。

## 重构目标

- 路径: {{target}}
- 重构类型: {{type}}
- 范围: {{scope}}
{{#if preserve_behavior}}
- 保持原有行为: 是
{{/if}}
{{#if generate_tests}}
- 生成测试: 是
{{/if}}

## 重构原则

请遵循以下原则进行重构：

### 1. SOLID 原则
- 单一职责原则 (SRP)
- 开闭原则 (OCP)
- 里氏替换原则 (LSP)
- 接口隔离原则 (ISP)
- 依赖反转原则 (DIP)

### 2. 设计模式
- 识别适合应用的设计模式
- 避免过度设计
- 保持简单性

### 3. 代码质量
- 减少重复代码
- 提高可读性
- 降低复杂度
- 改善命名

### 4. 性能优化
- 算法优化
- 数据结构选择
- 内存使用优化

## 重构类型说明

{{#eq type "extract-method"}}
- 提取方法：将代码片段提取为独立方法
- 提高代码复用性
- 增强可读性
{{/eq}}

{{#eq type "extract-class"}}
- 提取类：将职责分离到新类
- 遵循单一职责原则
- 改善代码组织
{{/eq}}

{{#eq type "rename"}}
- 重命名：改善变量、方法、类名
- 提高代码可读性
- 使名称更具描述性
{{/eq}}

{{#eq type "simplify"}}
- 简化：简化复杂逻辑
- 减少嵌套层级
- 提高可理解性
{{/eq}}

{{#eq type "optimize"}}
- 优化：性能优化
- 算法改进
- 资源使用优化
{{/eq}}

{{#eq type "modernize"}}
- 现代化：使用现代语言特性
- API 更新
- 最佳实践应用
{{/eq}}

## 输出格式

请按以下 JSON 格式输出结果：

```json
{
  "summary": {
    "files_modified": 0,
    "lines_changed": 0,
    "complexity_reduction": "0%",
    "quality_improvement": "description"
  },
  "changes": [
    {
      "file": "path/to/file",
      "type": "modified|created|deleted",
      "description": "变更描述",
      "before": "原代码片段",
      "after": "重构后代码",
      "rationale": "重构原因"
    }
  ],
  "tests": {
    "generated": true,
    "coverage": "0%",
    "files": ["test-file-paths"]
  },
  "recommendations": [
    "后续改进建议"
  ]
}
```

## 注意事项

1. 确保重构后代码行为不变（如果 preserve_behavior 为 true）
2. 保留原有的注释和文档
3. 更新相关的导入和引用
4. 生成必要的测试用例