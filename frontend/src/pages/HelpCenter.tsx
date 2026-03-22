import React, { useState } from 'react';
import './HelpCenter.css';

interface HelpArticle {
  id: string;
  title: string;
  category: string;
  content: string;
}

function HelpCenter() {
  const [activeCategory, setActiveCategory] = useState('getting-started');
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedArticle, setExpandedArticle] = useState<string | null>(null);

  const categories = [
    { id: 'getting-started', label: '快速开始', icon: '🚀' },
    { id: 'agents', label: 'Agent 管理', icon: '🤖' },
    { id: 'workflows', label: '工作流编排', icon: '🔄' },
    { id: 'schedules', label: '定时任务', icon: '⏰' },
    { id: 'api', label: 'API 使用', icon: '🔌' },
    { id: 'shortcuts', label: '快捷键', icon: '⌨️' },
    { id: 'troubleshooting', label: '常见问题', icon: '❓' },
  ];

  const articles: HelpArticle[] = [
    {
      id: '1',
      category: 'getting-started',
      title: '平台简介',
      content: `
## Claude Code Agent 编排平台

这是一个灵活的 Agent 编排平台，让你能够：
- 配置和组合多个 Claude Code Agent
- 通过 API 触发 Agent 任务
- 定制任何业务需求
- 从 Agent 执行中获取结构化结果

### 核心概念

1. **Agent**: 配置好的 Claude Code CLI 实例，具有特定能力
2. **Workflow**: 定义多个 Agent 如何协同工作
3. **Execution**: 工作流的执行实例
4. **ScheduledJob**: 定时执行 Agent 或工作流
      `,
    },
    {
      id: '2',
      category: 'getting-started',
      title: '创建第一个 Agent',
      content: `
## 创建 Agent

1. 进入 **Agents** 页面
2. 点击 **+ 创建 Agent** 按钮
3. 填写 Agent 基本信息：
   - 名称：Agent 的标识
   - 描述：说明 Agent 的用途
   - 模型：选择 Claude 模型
   - 系统提示词：定义 Agent 行为
4. 配置 Agent 能力：
   - 工具：Agent 可使用的工具
   - 权限：Agent 的操作权限
5. 点击保存完成创建
      `,
    },
    {
      id: '3',
      category: 'agents',
      title: 'Agent 配置详解',
      content: `
## Agent 配置项

### 基本配置
- **模型**: claude-sonnet-4-6, claude-opus-4-6, claude-haiku-4-5
- **系统提示词**: 定义 Agent 的行为和角色
- **最大 Tokens**: 限制输出长度
- **超时时间**: 执行超时设置

### 技能配置
Agent 可以选择多个 Skill，实现能力组合：
- 输入映射：定义输入参数转换
- 输出映射：定义输出格式转换

### 隔离配置
- 数据隔离：Agent 执行数据独立存储
- 会话隔离：每个 Agent 有独立会话
- 网络隔离：限制网络访问
- 文件隔离：独立文件系统命名空间
      `,
    },
    {
      id: '4',
      category: 'workflows',
      title: '工作流编排指南',
      content: `
## 工作流模式

### 串行执行 (Serial)
Agent 按顺序依次执行，前一个 Agent 的输出可作为后一个 Agent 的输入。

### 并行执行 (Parallel)
所有 Agent 同时执行，适合独立任务。

### 混合模式 (Hybrid)
结合串行和并行，通过依赖关系定义执行流程。

## 创建工作流

1. 从模板库选择模板，或手动创建
2. 添加 Agent 节点
3. 配置节点间的数据流
4. 设置执行模式
5. 保存并测试
      `,
    },
    {
      id: '5',
      category: 'schedules',
      title: '定时任务配置',
      content: `
## Cron 表达式

格式：\`分 时 日 月 周\`

常用示例：
- \`*/10 * * * *\` - 每 10 分钟
- \`0 * * * *\` - 每小时
- \`0 9 * * *\` - 每天上午 9 点
- \`0 9 * * 1-5\` - 工作日上午 9 点
- \`0 0 1 * *\` - 每月 1 号

## 失败处理

- **通知**: 发送邮件通知
- **重试**: 自动重试执行
- **禁用**: 禁用定时任务
      `,
    },
    {
      id: '6',
      category: 'api',
      title: 'API 认证',
      content: `
## API 密钥

在 **Settings > API 密钥** 中管理 API 密钥。

### 使用方式

\`\`\`bash
curl -X GET "https://api.example.com/api/agents" \\
  -H "Authorization: Bearer YOUR_API_KEY"
\`\`\`

### 权限级别

- **read**: 只读访问
- **write**: 写入权限
- **execute**: 执行权限
- **admin**: 管理员权限
      `,
    },
    {
      id: '7',
      category: 'troubleshooting',
      title: '执行失败排查',
      content: `
## 常见执行失败原因

### 1. 超时
- 检查 Agent 超时设置
- 优化任务复杂度

### 2. 权限不足
- 检查 Agent 权限配置
- 确认所需工具已启用

### 3. 输入参数错误
- 验证输入 Schema
- 检查必填字段

### 4. 资源限制
- 检查配额使用情况
- 确认并发限制

## 查看日志

1. 进入 Executions 页面
2. 点击具体执行记录
3. 切换到日志标签页
4. 分析错误信息
      `,
    },
    {
      id: '8',
      category: 'shortcuts',
      title: '键盘快捷键',
      content: `
## 全局快捷键

### 导航
- \`g h\` - 前往首页
- \`g a\` - 前往 Agent 页面
- \`g w\` - 前往工作流页面
- \`g e\` - 前往执行页面
- \`g s\` - 前往定时任务页面
- \`/\` - 打开全局搜索

### 操作
- \`n\` - 创建新项目 (Agent/工作流)
- \`r\` - 刷新当前页面
- \`?\` - 显示快捷键帮助
- \`Esc\` - 关闭弹窗/取消操作

### 执行
- \`Ctrl+Enter\` - 提交表单/执行
- \`Ctrl+K\` - 打开命令面板

按 \`?\` 可随时打开快捷键帮助面板。
      `,
    },
    {
      id: '9',
      category: 'shortcuts',
      title: '快速导航技巧',
      content: `
## 高效使用平台

### 快速跳转
使用 \`g\` + 字母组合可以快速跳转到各个页面：
1. 先按 \`g\` 键
2. 在 1 秒内按目标键（如 \`a\` 跳转到 Agent 页面）

### 全局搜索
按 \`/\` 键打开全局搜索，可以搜索：
- Agent 名称
- 工作流名称
- 执行记录
- 定时任务

### 批量操作
在执行页面：
- 使用复选框选择多个执行
- 批量取消选中的执行
- 按状态筛选执行

### 实时监控
- Dashboard 页面实时显示执行状态
- 使用 SSE 实时更新执行进度
- 通知中心显示重要事件
      `,
    },
  ];

  const filteredArticles = articles.filter((article) => {
    const matchesCategory = article.category === activeCategory;
    const matchesSearch = !searchQuery ||
      article.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      article.content.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  return (
    <div className="help-center-page">
      <div className="page-header">
        <h1>帮助中心</h1>
        <p>了解如何使用平台的各项功能</p>
      </div>

      {/* Search */}
      <div className="search-section">
        <input
          type="text"
          placeholder="搜索帮助文档..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="help-search"
        />
      </div>

      <div className="help-content">
        {/* Categories Sidebar */}
        <div className="categories-sidebar">
          {categories.map((category) => (
            <button
              key={category.id}
              className={`category-btn ${activeCategory === category.id ? 'active' : ''}`}
              onClick={() => {
                setActiveCategory(category.id);
                setExpandedArticle(null);
              }}
            >
              <span className="category-icon">{category.icon}</span>
              <span className="category-label">{category.label}</span>
            </button>
          ))}
        </div>

        {/* Articles List */}
        <div className="articles-section">
          <h2 className="section-title">
            {categories.find((c) => c.id === activeCategory)?.label}
          </h2>

          <div className="articles-list">
            {filteredArticles.length > 0 ? (
              filteredArticles.map((article) => (
                <div key={article.id} className="article-card">
                  <div
                    className="article-header"
                    onClick={() => setExpandedArticle(
                      expandedArticle === article.id ? null : article.id
                    )}
                  >
                    <h3>{article.title}</h3>
                    <span className="expand-icon">
                      {expandedArticle === article.id ? '▼' : '▶'}
                    </span>
                  </div>
                  {expandedArticle === article.id && (
                    <div
                      className="article-content"
                      dangerouslySetInnerHTML={{
                        __html: article.content
                          .replace(/## (.*)/g, '<h4>$1</h4>')
                          .replace(/### (.*)/g, '<h5>$1</h5>')
                          .replace(/- (.*)/g, '<li>$1</li>')
                          .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
                          .replace(/\n\n/g, '</p><p>')
                          .replace(/\n/g, '<br/>'),
                      }}
                    />
                  )}
                </div>
              ))
            ) : (
              <div className="no-results">
                <span className="no-results-icon">🔍</span>
                <p>未找到相关文档</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Quick Links */}
      <div className="quick-links">
        <h3>快速链接</h3>
        <div className="links-grid">
          <a href="https://github.com/HOA233/claude-code-pipeline" target="_blank" rel="noopener noreferrer" className="link-card">
            <span className="link-icon">📚</span>
            <span>GitHub 仓库</span>
          </a>
          <a href="https://docs.anthropic.com" target="_blank" rel="noopener noreferrer" className="link-card">
            <span className="link-icon">📖</span>
            <span>Claude API 文档</span>
          </a>
          <a href="mailto:support@example.com" className="link-card">
            <span className="link-icon">📧</span>
            <span>联系支持</span>
          </a>
        </div>
      </div>
    </div>
  );
}

export default HelpCenter;