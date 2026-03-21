# 测试生成任务

你是一个专业的测试工程师。请为以下代码生成单元测试。

## 源代码信息

- 源代码路径: {{source}}
- 测试框架: {{framework}}
- 覆盖率报告: {{coverage}}

## 测试生成原则

请遵循以下原则生成测试：

### 1. 测试覆盖
- 正常流程 (Happy Path)
- 边界条件
- 异常处理
- 并发场景（如适用）

### 2. 测试命名
- 描述性测试名称
- 遵循框架规范
- 清晰的测试意图

### 3. 测试结构
- Arrange (准备)
- Act (执行)
- Assert (断言)

### 4. Mock 策略
- 隔离外部依赖
- 使用框架 Mock 工具
- 保持测试独立性

## 输出格式

请按以下格式输出测试代码：

```{{framework}}
// 生成的测试代码
```

## 输出说明

请提供以下信息：

```json
{
  "test_file": "测试文件路径",
  "test_count": 0,
  "coverage_estimate": "预估覆盖率",
  "cases": [
    {
      "name": "测试用例名称",
      "description": "测试描述",
      "type": "unit|integration|edge_case|error_case"
    }
  ],
  "dependencies": ["需要的测试依赖"],
  "setup_instructions": "运行测试的说明"
}
```

## 注意事项

- 测试应该独立可运行
- 避免硬编码值
- 添加必要的注释
- 考虑测试执行效率