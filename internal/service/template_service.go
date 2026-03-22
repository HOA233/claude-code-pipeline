package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
)

// TemplateService 工作流模板服务
type TemplateService struct {
	redis       *repository.RedisClient
	agentSvc    *AgentService
	workflowSvc *WorkflowService
}

// NewTemplateService 创建模板服务
func NewTemplateService(redis *repository.RedisClient, agentSvc *AgentService, workflowSvc *WorkflowService) *TemplateService {
	return &TemplateService{
		redis:       redis,
		agentSvc:    agentSvc,
		workflowSvc: workflowSvc,
	}
}

// WorkflowTemplate 工作流模板
type WorkflowTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Icon        string                 `json:"icon"`
	Agents      []TemplateAgent        `json:"agents"`
	Connections []TemplateConnection   `json:"connections"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
}

// TemplateAgent 模板中的 Agent 定义
type TemplateAgent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Model       string                 `json:"model"`
	SystemPrompt string                `json:"system_prompt"`
	Tools       []Tool                 `json:"tools"`
	Permissions []Permission           `json:"permissions"`
	Timeout     int                    `json:"timeout"`
	Input       map[string]interface{} `json:"input"`
	OutputAs    string                 `json:"output_as"`
	DependsOn   []string               `json:"depends_on"`
}

// TemplateConnection 模板中的连接定义
type TemplateConnection struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GetBuiltInTemplates 获取内置模板
func (s *TemplateService) GetBuiltInTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			ID:          "code-review-workflow",
			Name:        "代码审查工作流",
			Description: "完整的代码审查流程，包括代码分析、测试生成和文档更新",
			Category:    "code-review",
			Icon:        "🔍",
			Agents: []TemplateAgent{
				{
					ID:          "reviewer",
					Name:        "Code Reviewer",
					Description: "代码审查 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个专业的代码审查专家。分析代码质量、潜在问题、最佳实践 violations，并提供改进建议。",
					Timeout:     300,
					OutputAs:    "review_result",
				},
				{
					ID:          "test-gen",
					Name:        "Test Generator",
					Description: "测试生成 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个测试工程师。根据代码分析结果生成全面的单元测试和集成测试。",
					Timeout:     300,
					OutputAs:    "tests",
					DependsOn:   []string{"reviewer"},
				},
				{
					ID:          "doc-updater",
					Name:        "Doc Updater",
					Description: "文档更新 Agent",
					Model:       "claude-haiku-4-5",
					SystemPrompt: "你是一个技术文档专家。更新代码相关的文档，包括 README、API 文档和注释。",
					Timeout:     180,
					OutputAs:    "documentation",
					DependsOn:   []string{"reviewer"},
				},
			},
			Connections: []TemplateConnection{
				{From: "reviewer", To: "test-gen"},
				{From: "reviewer", To: "doc-updater"},
			},
		},
		{
			ID:          "security-scan-workflow",
			Name:        "安全扫描工作流",
			Description: "全面的安全漏洞扫描和修复建议",
			Category:    "security",
			Icon:        "🔒",
			Agents: []TemplateAgent{
				{
					ID:          "static-scanner",
					Name:        "Static Scanner",
					Description: "静态代码安全扫描",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个安全分析专家。扫描代码中的安全漏洞，包括 SQL 注入、XSS、CSRF、硬编码密钥等。",
					Timeout:     300,
					OutputAs:    "static_findings",
				},
				{
					ID:          "dep-scanner",
					Name:        "Dependency Scanner",
					Description: "依赖漏洞扫描",
					Model:       "claude-haiku-4-5",
					SystemPrompt: "你是一个依赖安全专家。分析项目依赖中的已知漏洞和版本风险。",
					Timeout:     180,
					OutputAs:    "dep_findings",
				},
				{
					ID:          "remediation",
					Name:        "Remediation Advisor",
					Description: "修复建议生成",
					Model:       "claude-opus-4-6",
					SystemPrompt: "你是一个安全修复专家。根据扫描结果生成详细的修复建议和代码示例。",
					Timeout:     300,
					OutputAs:    "remediation_plan",
					DependsOn:   []string{"static-scanner", "dep-scanner"},
				},
			},
			Connections: []TemplateConnection{
				{From: "static-scanner", To: "remediation"},
				{From: "dep-scanner", To: "remediation"},
			},
		},
		{
			ID:          "feature-development-workflow",
			Name:        "功能开发工作流",
			Description: "从需求到代码的完整功能开发流程",
			Category:    "development",
			Icon:        "🚀",
			Agents: []TemplateAgent{
				{
					ID:          "planner",
					Name:        "Feature Planner",
					Description: "功能规划 Agent",
					Model:       "claude-opus-4-6",
					SystemPrompt: "你是一个技术架构师。分析需求，设计技术方案，拆分开发任务。",
					Timeout:     300,
					OutputAs:    "plan",
				},
				{
					ID:          "coder",
					Name:        "Code Generator",
					Description: "代码生成 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个全栈开发者。根据规划生成高质量的代码实现。",
					Timeout:     600,
					OutputAs:    "code",
					DependsOn:   []string{"planner"},
				},
				{
					ID:          "tester",
					Name:        "Test Engineer",
					Description: "测试工程师 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个测试工程师。为新功能编写全面的测试用例。",
					Timeout:     300,
					OutputAs:    "tests",
					DependsOn:   []string{"coder"},
				},
				{
					ID:          "reviewer",
					Name:        "Code Reviewer",
					Description: "代码审查 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个代码审查专家。审查代码质量、最佳实践和潜在问题。",
					Timeout:     180,
					OutputAs:    "review",
					DependsOn:   []string{"coder"},
				},
			},
			Connections: []TemplateConnection{
				{From: "planner", To: "coder"},
				{From: "coder", To: "tester"},
				{From: "coder", To: "reviewer"},
			},
		},
		{
			ID:          "bug-fix-workflow",
			Name:        "Bug 修复工作流",
			Description: "自动化的 Bug 分析和修复流程",
			Category:    "debugging",
			Icon:        "🐛",
			Agents: []TemplateAgent{
				{
					ID:          "analyzer",
					Name:        "Bug Analyzer",
					Description: "Bug 分析 Agent",
					Model:       "claude-opus-4-6",
					SystemPrompt: "你是一个 Bug 分析专家。分析错误日志、堆栈跟踪，定位问题根因。",
					Timeout:     300,
					OutputAs:    "analysis",
				},
				{
					ID:          "fixer",
					Name:        "Bug Fixer",
					Description: "Bug 修复 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个修复专家。根据分析结果生成修复代码，确保不引入新问题。",
					Timeout:     300,
					OutputAs:    "fix",
					DependsOn:   []string{"analyzer"},
				},
				{
					ID:          "validator",
					Name:        "Fix Validator",
					Description: "修复验证 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个 QA 工程师。验证修复是否有效，检查边界情况和回归风险。",
					Timeout:     180,
					OutputAs:    "validation",
					DependsOn:   []string{"fixer"},
				},
			},
			Connections: []TemplateConnection{
				{From: "analyzer", To: "fixer"},
				{From: "fixer", To: "validator"},
			},
		},
		{
			ID:          "documentation-workflow",
			Name:        "文档生成工作流",
			Description: "自动生成和更新项目文档",
			Category:    "documentation",
			Icon:        "📝",
			Agents: []TemplateAgent{
				{
					ID:          "analyzer",
					Name:        "Code Analyzer",
					Description: "代码分析 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个代码分析专家。分析代码结构、API 接口、数据模型。",
					Timeout:     300,
					OutputAs:    "analysis",
				},
				{
					ID:          "doc-gen",
					Name:        "Doc Generator",
					Description: "文档生成 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个技术文档专家。根据代码分析生成清晰、完整的文档。",
					Timeout:     300,
					OutputAs:    "docs",
					DependsOn:   []string{"analyzer"},
				},
			},
			Connections: []TemplateConnection{
				{From: "analyzer", To: "doc-gen"},
			},
		},
		{
			ID:          "ci-cd-pipeline-workflow",
			Name:        "CI/CD 流水线工作流",
			Description: "完整的持续集成和部署流水线",
			Category:    "devops",
			Icon:        "🔄",
			Agents: []TemplateAgent{
				{
					ID:          "build-checker",
					Name:        "Build Checker",
					Description: "构建检查 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个构建系统专家。检查构建配置、依赖管理和编译问题。",
					Timeout:     300,
					OutputAs:    "build_status",
				},
				{
					ID:          "test-runner",
					Name:        "Test Runner",
					Description: "测试运行 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个测试专家。运行和分析单元测试、集成测试、端到端测试结果。",
					Timeout:     600,
					OutputAs:    "test_results",
					DependsOn:   []string{"build-checker"},
				},
				{
					ID:          "deploy-agent",
					Name:        "Deploy Agent",
					Description: "部署 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个 DevOps 工程师。管理部署流程、环境配置和发布说明。",
					Timeout:     300,
					OutputAs:    "deployment",
					DependsOn:   []string{"test-runner"},
				},
			},
			Connections: []TemplateConnection{
				{From: "build-checker", To: "test-runner"},
				{From: "test-runner", To: "deploy-agent"},
			},
		},
		{
			ID:          "performance-analysis-workflow",
			Name:        "性能分析工作流",
			Description: "分析系统性能瓶颈并提供优化建议",
			Category:    "performance",
			Icon:        "⚡",
			Agents: []TemplateAgent{
				{
					ID:          "profiler",
					Name:        "Profiler",
					Description: "性能分析 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个性能分析专家。分析 CPU、内存、I/O 使用情况，识别性能瓶颈。",
					Timeout:     300,
					OutputAs:    "profile_data",
				},
				{
					ID:          "db-analyzer",
					Name:        "DB Analyzer",
					Description: "数据库分析 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个数据库优化专家。分析 SQL 查询性能、索引使用和连接池配置。",
					Timeout:     300,
					OutputAs:    "db_analysis",
				},
				{
					ID:          "optimizer",
					Name:        "Optimizer",
					Description: "优化建议 Agent",
					Model:       "claude-opus-4-6",
					SystemPrompt: "你是一个系统优化专家。根据性能分析结果提供具体的优化建议和实现方案。",
					Timeout:     300,
					OutputAs:    "optimization_plan",
					DependsOn:   []string{"profiler", "db-analyzer"},
				},
			},
			Connections: []TemplateConnection{
				{From: "profiler", To: "optimizer"},
				{From: "db-analyzer", To: "optimizer"},
			},
		},
		{
			ID:          "api-development-workflow",
			Name:        "API 开发工作流",
			Description: "从设计到文档的完整 API 开发流程",
			Category:    "development",
			Icon:        "🔌",
			Agents: []TemplateAgent{
				{
					ID:          "designer",
					Name:        "API Designer",
					Description: "API 设计 Agent",
					Model:       "claude-opus-4-6",
					SystemPrompt: "你是一个 API 架构师。设计 RESTful 或 GraphQL API，定义端点、数据模型和认证方式。",
					Timeout:     300,
					OutputAs:    "api_design",
				},
				{
					ID:          "implementer",
					Name:        "API Implementer",
					Description: "API 实现 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个后端开发者。根据 API 设计实现控制器、服务和数据访问层。",
					Timeout:     600,
					OutputAs:    "implementation",
					DependsOn:   []string{"designer"},
				},
				{
					ID:          "doc-writer",
					Name:        "API Doc Writer",
					Description: "API 文档 Agent",
					Model:       "claude-sonnet-4-6",
					SystemPrompt: "你是一个 API 文档专家。生成 OpenAPI/Swagger 规范和使用示例。",
					Timeout:     180,
					OutputAs:    "api_docs",
					DependsOn:   []string{"implementer"},
				},
			},
			Connections: []TemplateConnection{
				{From: "designer", To: "implementer"},
				{From: "implementer", To: "doc-writer"},
			},
		},
	}
}

// InstantiateTemplate 实例化模板
func (s *TemplateService) InstantiateTemplate(ctx context.Context, templateID string, name string) (*model.Workflow, error) {
	template := s.findTemplate(templateID)
	if template == nil {
		return nil, nil
	}

	// 创建 Agents
	agentMap := make(map[string]string) // template agent id -> real agent id
	for _, ta := range template.Agents {
		agent := &model.Agent{
			ID:          uuid.New().String(),
			Name:        ta.Name,
			Description: ta.Description,
			Model:       ta.Model,
			SystemPrompt: ta.SystemPrompt,
			Tools:       ta.Tools,
			Permissions: ta.Permissions,
			Timeout:     ta.Timeout,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.agentSvc.CreateAgentDirect(ctx, agent); err != nil {
			return nil, err
		}
		agentMap[ta.ID] = agent.ID
	}

	// 创建 Workflow
	workflowAgents := make([]model.AgentNode, 0, len(template.Agents))
	for _, ta := range template.Agents {
		node := model.AgentNode{
			ID:        uuid.New().String(),
			AgentID:   agentMap[ta.ID],
			Name:      ta.Name,
			Input:     ta.Input,
			OutputAs:  ta.OutputAs,
			DependsOn: ta.DependsOn,
		}
		workflowAgents = append(workflowAgents, node)
	}

	workflow := &model.WorkflowCreateRequest{
		Name:        name,
		Description: template.Description,
		Agents:      workflowAgents,
		Mode:        "hybrid",
	}

	return s.workflowSvc.CreateWorkflow(ctx, workflow)
}

// SaveCustomTemplate 保存自定义模板
func (s *TemplateService) SaveCustomTemplate(ctx context.Context, template *WorkflowTemplate) error {
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	template.CreatedAt = time.Now()

	data, err := json.Marshal(template)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, "template:"+template.ID, data, 0)
}

// ListCustomTemplates 列出自定义模板
func (s *TemplateService) ListCustomTemplates(ctx context.Context) ([]*WorkflowTemplate, error) {
	keys, err := s.redis.Keys(ctx, "template:*")
	if err != nil {
		return nil, err
	}

	templates := make([]*WorkflowTemplate, 0, len(keys))
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var template WorkflowTemplate
		if err := json.Unmarshal(data, &template); err != nil {
			continue
		}
		templates = append(templates, &template)
	}

	return templates, nil
}

// DeleteCustomTemplate 删除自定义模板
func (s *TemplateService) DeleteCustomTemplate(ctx context.Context, templateID string) error {
	return s.redis.Delete(ctx, "template:"+templateID)
}

func (s *TemplateService) findTemplate(id string) *WorkflowTemplate {
	templates := s.GetBuiltInTemplates()
	for i := range templates {
		if templates[i].ID == id {
			return &templates[i]
		}
	}
	return nil
}