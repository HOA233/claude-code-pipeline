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