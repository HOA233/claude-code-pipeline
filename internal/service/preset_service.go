package service

import (
	"context"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
)

// PresetAgentService 预置 Agent 服务
type PresetAgentService struct {
	redis *repository.RedisClient
}

// NewPresetAgentService 创建预置 Agent 服务
func NewPresetAgentService(redis *repository.RedisClient) *PresetAgentService {
	return &PresetAgentService{redis: redis}
}

// InitializePresetAgents 初始化预置 Agent
func (s *PresetAgentService) InitializePresetAgents(ctx context.Context) error {
	presets := s.getPresetAgents()

	for _, agent := range presets {
		// 检查是否已存在
		if existing, _ := s.redis.GetAgent(ctx, agent.ID); existing == nil {
			if err := s.redis.SaveAgent(ctx, agent); err != nil {
				return err
			}
		}
	}

	return nil
}

// getPresetAgents 获取预置 Agent 列表
func (s *PresetAgentService) getPresetAgents() []*model.Agent {
	now := time.Now()

	return []*model.Agent{
		// 代码审查 Agent
		{
			ID:           "preset-code-reviewer",
			Name:         "Code Reviewer",
			Description:  "专业的代码审查助手，帮助发现代码问题和改进空间",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个专业的代码审查助手。请审查用户提供的代码，找出潜在问题、安全漏洞、性能问题和改进建议。输出结构化的审查报告。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
				{Name: "glob", Description: "搜索文件模式"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
			},
			Timeout:     300,
			Category:    "analysis",
			Tags:        []string{"code-review", "quality", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 测试生成 Agent
		{
			ID:           "preset-test-generator",
			Name:         "Test Generator",
			Description:  "根据代码自动生成单元测试",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个测试生成专家。根据提供的代码，生成全面的单元测试，包括正常情况、边界情况和错误情况。使用标准测试框架。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
				{Name: "write_file", Description: "写入文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
				{Resource: "file", Action: "write"},
			},
			Timeout:     300,
			Category:    "testing",
			Tags:        []string{"test", "unit-test", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 安全扫描 Agent
		{
			ID:           "preset-security-scanner",
			Name:         "Security Scanner",
			Description:  "扫描代码中的安全漏洞",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个安全专家。扫描代码中的安全漏洞，包括 SQL 注入、XSS、CSRF、敏感信息泄露等问题。输出详细的安全报告和修复建议。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
				{Name: "grep", Description: "搜索代码模式"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
			},
			Timeout:     300,
			Category:    "security",
			Tags:        []string{"security", "vulnerability", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 性能分析 Agent
		{
			ID:           "preset-performance-analyzer",
			Name:         "Performance Analyzer",
			Description:  "分析代码性能问题",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个性能优化专家。分析代码的性能问题，包括时间复杂度、空间复杂度、内存泄漏、不必要的计算等。提供优化建议。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
			},
			Timeout:     300,
			Category:    "performance",
			Tags:        []string{"performance", "optimization", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 文档生成 Agent
		{
			ID:           "preset-doc-generator",
			Name:         "Documentation Generator",
			Description:  "根据代码生成文档",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个技术文档专家。根据代码生成清晰的文档，包括函数说明、参数描述、返回值说明、使用示例等。使用 Markdown 格式。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
				{Name: "write_file", Description: "写入文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
				{Resource: "file", Action: "write"},
			},
			Timeout:     300,
			Category:    "documentation",
			Tags:        []string{"documentation", "markdown", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 重构建议 Agent
		{
			ID:           "preset-refactor-advisor",
			Name:         "Refactor Advisor",
			Description:  "提供代码重构建议",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个重构专家。分析代码的结构和设计，提供重构建议，包括设计模式应用、代码组织、模块化、消除重复代码等。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
			},
			Timeout:     300,
			Category:    "refactoring",
			Tags:        []string{"refactor", "design", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// 代码解释 Agent
		{
			ID:           "preset-code-explainer",
			Name:         "Code Explainer",
			Description:  "解释复杂代码逻辑",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个代码解释专家。用简单易懂的语言解释复杂的代码逻辑，帮助开发者理解代码的工作原理。",
			MaxTokens:    4096,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
			},
			Timeout:     120,
			Category:    "explanation",
			Tags:        []string{"explain", "learning", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// Bug 修复 Agent
		{
			ID:           "preset-bug-fixer",
			Name:         "Bug Fixer",
			Description:  "帮助定位和修复 Bug",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "你是一个调试专家。根据错误信息和代码，帮助定位 Bug 的原因，并提供修复方案。解释问题的根本原因。",
			MaxTokens:    8192,
			Tools: []model.Tool{
				{Name: "read_file", Description: "读取文件内容"},
				{Name: "write_file", Description: "写入文件内容"},
			},
			Permissions: []model.Permission{
				{Resource: "file", Action: "read"},
				{Resource: "file", Action: "write"},
			},
			Timeout:     300,
			Category:    "debugging",
			Tags:        []string{"bug", "fix", "debug", "preset"},
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// ListPresetAgents 列出预置 Agent
func (s *PresetAgentService) ListPresetAgents() []*model.Agent {
	return s.getPresetAgents()
}

// WorkflowTemplateService 工作流模板服务
type WorkflowTemplateService struct {
	redis *repository.RedisClient
}

// NewWorkflowTemplateService 创建工作流模板服务
func NewWorkflowTemplateService(redis *repository.RedisClient) *WorkflowTemplateService {
	return &WorkflowTemplateService{redis: redis}
}

// GetWorkflowTemplates 获取工作流模板
func (s *WorkflowTemplateService) GetWorkflowTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			ID:          "template-code-review",
			Name:        "代码审查流程",
			Description: "完整的代码审查流程，包括静态分析、安全检查和性能分析",
			Category:    "quality",
			Definition: model.WorkflowCreateRequest{
				Name:        "Code Review Pipeline",
				Description: "完整的代码审查流程",
				Agents: []model.AgentNode{
					{
						ID:        "review",
						AgentID:   "preset-code-reviewer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "review_result",
					},
					{
						ID:        "security",
						AgentID:   "preset-security-scanner",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "security_result",
					},
					{
						ID:        "performance",
						AgentID:   "preset-performance-analyzer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "performance_result",
					},
				},
				Mode: model.ModeParallel,
			},
		},
		{
			ID:          "template-test-generation",
			Name:        "测试生成流程",
			Description: "分析代码并生成全面的单元测试",
			Category:    "testing",
			Definition: model.WorkflowCreateRequest{
				Name:        "Test Generation Pipeline",
				Description: "分析代码并生成测试",
				Agents: []model.AgentNode{
					{
						ID:        "analyze",
						AgentID:   "preset-code-explainer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "analysis",
					},
					{
						ID:        "generate",
						AgentID:   "preset-test-generator",
						InputFrom: map[string]string{"analysis": "analysis"},
						OutputAs:  "tests",
						DependsOn: []string{"analyze"},
					},
				},
				Mode: model.ModeSerial,
			},
		},
		{
			ID:          "template-full-analysis",
			Name:        "全面分析流程",
			Description: "代码审查 + 测试生成 + 文档生成",
			Category:    "comprehensive",
			Definition: model.WorkflowCreateRequest{
				Name:        "Full Analysis Pipeline",
				Description: "全面代码分析",
				Agents: []model.AgentNode{
					{
						ID:        "review",
						AgentID:   "preset-code-reviewer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "review",
					},
					{
						ID:        "security",
						AgentID:   "preset-security-scanner",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "security",
					},
					{
						ID:        "performance",
						AgentID:   "preset-performance-analyzer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "performance",
					},
					{
						ID:        "tests",
						AgentID:   "preset-test-generator",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "tests",
					},
					{
						ID:        "docs",
						AgentID:   "preset-doc-generator",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "docs",
					},
				},
				Mode: model.ModeParallel,
			},
		},
		{
			ID:          "template-refactor",
			Name:        "重构建议流程",
			Description: "分析代码并提供重构建议",
			Category:    "refactoring",
			Definition: model.WorkflowCreateRequest{
				Name:        "Refactor Pipeline",
				Description: "代码重构分析和建议",
				Agents: []model.AgentNode{
					{
						ID:        "review",
						AgentID:   "preset-code-reviewer",
						Input:     map[string]interface{}{"target": "{{input.target}}"},
						OutputAs:  "issues",
					},
					{
						ID:        "refactor",
						AgentID:   "preset-refactor-advisor",
						InputFrom: map[string]string{"issues": "issues"},
						OutputAs:  "suggestions",
						DependsOn: []string{"review"},
					},
				},
				Mode: model.ModeSerial,
			},
		},
	}
}

// WorkflowTemplate 工作流模板
type WorkflowTemplate struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Category    string                      `json:"category"`
	Definition  model.WorkflowCreateRequest `json:"definition"`
}

// CreateFromTemplate 从模板创建工作流
func (s *WorkflowTemplateService) CreateFromTemplate(ctx context.Context, templateID string, customName string) (*model.Workflow, error) {
	templates := s.GetWorkflowTemplates()

	var template *WorkflowTemplate
	for _, t := range templates {
		if t.ID == templateID {
			template = &t
			break
		}
	}

	if template == nil {
		return nil, nil
	}

	def := template.Definition
	if customName != "" {
		def.Name = customName
	}

	now := time.Now()
	workflow := &model.Workflow{
		ID:          uuid.New().String(),
		Name:        def.Name,
		Description: def.Description,
		Agents:      def.Agents,
		Connections: def.Connections,
		Mode:        def.Mode,
		Context:     def.Context,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.redis.SaveWorkflow(ctx, workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}