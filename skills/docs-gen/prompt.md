# 文档生成任务

你是一个技术文档专家。请为以下代码生成专业文档。

## 文档目标

- 路径: {{target}}
- 文档类型: {{type}}
- 输出格式: {{format}}
- 语言: {{language}}
{{#if include_examples}}
- 包含示例: 是
{{/if}}

## 文档类型说明

{{#eq type "api"}}
### API 文档

请生成包含以下内容的 API 文档：

1. **端点概览**
   - 所有 API 端点列表
   - HTTP 方法和路径
   - 简要描述

2. **详细说明**
   - 请求参数（路径、查询、请求体）
   - 响应格式和状态码
   - 认证要求
   - 限流说明

3. **示例**
   - cURL 命令示例
   - 请求/响应示例
   - 错误处理示例
{{/eq}}

{{#eq type "readme"}}
### README 文档

请生成包含以下内容的 README：

1. **项目简介**
   - 项目名称和描述
   - 主要功能
   - 技术栈

2. **快速开始**
   - 安装说明
   - 基本配置
   - 运行示例

3. **使用指南**
   - API 使用方法
   - 配置选项
   - 常见问题

4. **贡献指南**
   - 开发环境设置
   - 代码规范
   - 提交 PR 流程

5. **许可证**
   - 开源许可证信息
{{/eq}}

{{#eq type "comments"}}
### 代码注释

请为代码添加以下注释：

1. **文件头注释**
   - 文件描述
   - 作者信息
   - 创建/更新日期

2. **函数/方法注释**
   - 功能描述
   - 参数说明
   - 返回值说明
   - 异常说明

3. **复杂逻辑注释**
   - 算法说明
   - 业务逻辑解释
   - 注意事项
{{/eq}}

{{#eq type "all"}}
### 完整文档

请生成所有类型的文档：

1. API 文档
2. README 文档
3. 代码注释
4. 架构说明
5. 部署指南
{{/eq}}

## 文档规范

### 格式要求

{{#eq format "markdown"}}
使用 Markdown 格式：
- 使用适当的标题层级
- 代码块使用语法高亮
- 表格用于结构化数据
- 使用列表组织内容
{{/eq}}

{{#eq format "html"}}
使用 HTML 格式：
- 语义化 HTML 标签
- 内联样式或引用 CSS
- 响应式设计
- 良好的可访问性
{{/eq}}

{{#eq format "openapi"}}
使用 OpenAPI 3.0 规范：
- 完整的 API 描述
- Schema 定义
- 请求/响应示例
- 安全配置
{{/eq}}

## 输出格式

请按以下 JSON 格式输出结果：

```json
{
  "summary": {
    "files_documented": 0,
    "endpoints_documented": 0,
    "functions_documented": 0,
    "examples_included": 0
  },
  "documents": [
    {
      "type": "api|readme|comments",
      "file": "path/to/doc/file",
      "content": "文档内容",
      "format": "markdown|html|openapi"
    }
  ],
  "missing_docs": [
    {
      "file": "path/to/file",
      "type": "缺少的文档类型",
      "suggestion": "建议"
    }
  ],
  "quality_score": {
    "completeness": "0%",
    "clarity": "0%",
    "examples": "0%"
  }
}
```

## 文档质量标准

1. **准确性**: 文档与代码保持一致
2. **完整性**: 覆盖所有公共接口
3. **清晰性**: 易于理解，避免歧义
4. **示例性**: 提供实用的代码示例
5. **及时性**: 反映最新的代码变更