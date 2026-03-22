package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/google/uuid"
)

// TemplateService manages pipeline templates
type TemplateService struct {
	redis       *repository.RedisClient
	agentSvc    *AgentService
	workflowSvc *WorkflowService
}

// NewTemplateService creates a new template service
func NewTemplateService(redis *repository.RedisClient, agentSvc *AgentService, workflowSvc *WorkflowService) *TemplateService {
	return &TemplateService{
		redis:       redis,
		agentSvc:    agentSvc,
		workflowSvc: workflowSvc,
	}
}

// PipelineTmpl represents a reusable pipeline template
type PipelineTmpl struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Steps       []TemplateStep      `json:"steps"`
	Variables   map[string]VarDef   `json:"variables"`
	Mode        model.ExecutionMode `json:"mode"`
	CreatedAt   string              `json:"created_at"`
}

// TemplateStep represents a step in a template
type TemplateStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	CLI         string                 `json:"cli"`
	Action      string                 `json:"action"`
	Command     string                 `json:"command,omitempty"`
	Params      map[string]interface{} `json:"params"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	OnError     string                 `json:"on_error,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
}

// VarDef defines a template variable
type VarDef struct {
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Values      []string    `json:"values,omitempty"` // for enum types
}

// Built-in templates
var builtinTemplates = []PipelineTmpl{
	{
		ID:          "ci-pipeline",
		Name:        "CI Pipeline",
		Description: "Standard CI pipeline with lint, test, and build",
		Category:    "ci",
		Mode:        model.ModeSerial,
		Variables: map[string]VarDef{
			"source_dir": {
				Type:        "string",
				Required:    true,
				Description: "Source directory to process",
			},
			"test_framework": {
				Type:        "enum",
				Required:    false,
				Default:     "jest",
				Values:      []string{"jest", "pytest", "go-test", "junit"},
				Description: "Test framework to use",
			},
		},
		Steps: []TemplateStep{
			{
				ID:     "lint",
				Name:   "Lint Code",
				CLI:    "npm",
				Action: "run",
				Command: "lint",
				Params: map[string]interface{}{},
			},
			{
				ID:     "test",
				Name:   "Run Tests",
				CLI:    "npm",
				Action: "test",
				Params: map[string]interface{}{},
				DependsOn: []string{"lint"},
			},
			{
				ID:     "build",
				Name:   "Build",
				CLI:    "npm",
				Action: "run",
				Command: "build",
				Params: map[string]interface{}{},
				DependsOn: []string{"test"},
			},
		},
	},
	{
		ID:          "code-review-pipeline",
		Name:        "Code Review Pipeline",
		Description: "Automated code review with security analysis",
		Category:    "quality",
		Mode:        model.ModeParallel,
		Variables: map[string]VarDef{
			"target": {
				Type:        "string",
				Required:    true,
				Description: "Target directory or files to review",
			},
		},
		Steps: []TemplateStep{
			{
				ID:     "style-check",
				Name:   "Style Check",
				CLI:    "claude",
				Action: "review",
				Params: map[string]interface{}{
					"type": "style",
				},
			},
			{
				ID:     "security-scan",
				Name:   "Security Scan",
				CLI:    "claude",
				Action: "review",
				Params: map[string]interface{}{
					"type": "security",
				},
			},
			{
				ID:     "perf-check",
				Name:   "Performance Check",
				CLI:    "claude",
				Action: "review",
				Params: map[string]interface{}{
					"type": "performance",
				},
			},
		},
	},
	{
		ID:          "deploy-pipeline",
		Name:        "Deploy Pipeline",
		Description: "Standard deployment pipeline",
		Category:    "deploy",
		Mode:        model.ModeSerial,
		Variables: map[string]VarDef{
			"environment": {
				Type:        "enum",
				Required:    true,
				Values:      []string{"dev", "staging", "production"},
				Description: "Target environment",
			},
			"dry_run": {
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Perform dry run first",
			},
		},
		Steps: []TemplateStep{
			{
				ID:     "validate",
				Name:   "Validate Config",
				CLI:    "kubectl",
				Action: "apply",
				Command: "--dry-run=client",
				Params: map[string]interface{}{},
			},
			{
				ID:     "deploy",
				Name:   "Deploy",
				CLI:    "kubectl",
				Action: "apply",
				Params: map[string]interface{}{},
				DependsOn: []string{"validate"},
			},
			{
				ID:     "verify",
				Name:   "Verify Deployment",
				CLI:    "kubectl",
				Action: "rollout",
				Command: "status",
				Params: map[string]interface{}{},
				DependsOn: []string{"deploy"},
			},
		},
	},
	{
		ID:          "test-gen-pipeline",
		Name:        "Test Generation Pipeline",
		Description: "Generate tests for source code",
		Category:    "testing",
		Mode:        model.ModeSerial,
		Variables: map[string]VarDef{
			"source": {
				Type:        "string",
				Required:    true,
				Description: "Source directory",
			},
			"framework": {
				Type:        "enum",
				Required:    false,
				Default:     "jest",
				Values:      []string{"jest", "pytest", "go-test"},
				Description: "Test framework",
			},
		},
		Steps: []TemplateStep{
			{
				ID:     "analyze",
				Name:   "Analyze Code",
				CLI:    "claude",
				Action: "analyze",
				Params: map[string]interface{}{},
			},
			{
				ID:     "generate",
				Name:   "Generate Tests",
				CLI:    "claude",
				Action: "generate-tests",
				Params: map[string]interface{}{},
				DependsOn: []string{"analyze"},
			},
			{
				ID:     "run-tests",
				Name:   "Run Generated Tests",
				CLI:    "npm",
				Action: "test",
				Params: map[string]interface{}{},
				DependsOn: []string{"generate"},
			},
		},
	},
}

// ListTemplates returns all available templates
func (s *TemplateService) ListTemplates(ctx context.Context) ([]PipelineTmpl, error) {
	// Get custom templates from Redis
	keys, err := s.redis.ListCacheKeys(ctx, "template:*")
	if err != nil {
		return builtinTemplates, nil
	}

	templates := make([]PipelineTmpl, len(builtinTemplates))
	copy(templates, builtinTemplates)

	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var template PipelineTmpl
		if err := json.Unmarshal(data, &template); err != nil {
			continue
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// GetTemplate returns a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*PipelineTmpl, error) {
	// Check built-in templates
	for _, t := range builtinTemplates {
		if t.ID == id {
			return &t, nil
		}
	}

	// Check custom templates
	data, err := s.redis.Get(ctx, "template:"+id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	var template PipelineTmpl
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// CreatePipelineFromTemplate creates a pipeline from a template
func (s *TemplateService) CreatePipelineFromTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*model.PipelineCreateRequest, error) {
	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Validate and apply variables
	for name, def := range template.Variables {
		value, exists := variables[name]
		if !exists && def.Required {
			return nil, fmt.Errorf("required variable missing: %s", name)
		}
		if !exists && def.Default != nil {
			variables[name] = def.Default
		}

		// Validate enum values
		if def.Type == "enum" && len(def.Values) > 0 {
			strVal := fmt.Sprintf("%v", value)
			valid := false
			for _, v := range def.Values {
				if v == strVal {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("invalid value for %s: %s, allowed: %v", name, strVal, def.Values)
			}
		}
	}

	// Build steps with variable substitution
	steps := make([]model.Step, len(template.Steps))
	for i, ts := range template.Steps {
		params := make(map[string]interface{})
		for k, v := range ts.Params {
			params[k] = s.substituteVariables(v, variables)
		}

		steps[i] = model.Step{
			ID:        ts.ID,
			Name:      ts.Name,
			CLI:       ts.CLI,
			Action:    ts.Action,
			Command:   ts.Command,
			Params:    params,
			DependsOn: ts.DependsOn,
			OnError:   model.ErrorStrategy(ts.OnError),
			Timeout:   ts.Timeout,
		}
	}

	return &model.PipelineCreateRequest{
		Name:        template.Name,
		Description: template.Description,
		Mode:        template.Mode,
		Steps:       steps,
	}, nil
}

// SaveTemplate saves a custom template
func (s *TemplateService) SaveTemplate(ctx context.Context, template *PipelineTmpl) error {
	if template.ID == "" {
		template.ID = "template-" + uuid.New().String()[:8]
	}

	data, err := json.Marshal(template)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, "template:"+template.ID, data, 0)
}

// DeleteTemplate deletes a custom template
func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) error {
	// Don't allow deleting built-in templates
	for _, t := range builtinTemplates {
		if t.ID == id {
			return fmt.Errorf("cannot delete built-in template")
		}
	}

	return s.redis.Delete(ctx, "template:"+id)
}

// GetBuiltInTemplates returns built-in templates
func (s *TemplateService) GetBuiltInTemplates() []PipelineTmpl {
	return builtinTemplates
}

// ListCustomTemplates lists custom templates
func (s *TemplateService) ListCustomTemplates(ctx context.Context) ([]PipelineTmpl, error) {
	keys, err := s.redis.ListCacheKeys(ctx, "template:*")
	if err != nil {
		return []PipelineTmpl{}, nil
	}

	templates := []PipelineTmpl{}
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var template PipelineTmpl
		if err := json.Unmarshal(data, &template); err != nil {
			continue
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// InstantiateTemplate creates a workflow from a template
func (s *TemplateService) InstantiateTemplate(ctx context.Context, templateID string, name string) (*model.Workflow, error) {
	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	workflow := &model.Workflow{
		ID:          "workflow-" + uuid.New().String()[:8],
		Name:        name,
		Description: template.Description,
		Mode:        template.Mode,
		Enabled:     true,
	}

	return workflow, nil
}

// SaveCustomTemplate saves a custom template
func (s *TemplateService) SaveCustomTemplate(ctx context.Context, template *PipelineTmpl) error {
	return s.SaveTemplate(ctx, template)
}

// DeleteCustomTemplate deletes a custom template
func (s *TemplateService) DeleteCustomTemplate(ctx context.Context, id string) error {
	return s.DeleteTemplate(ctx, id)
}

func (s *TemplateService) substituteVariables(value interface{}, variables map[string]interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Simple variable substitution: ${var}
		for name, val := range variables {
			placeholder := "${" + name + "}"
			if contains(v, placeholder) {
				return replaceAll(v, placeholder, fmt.Sprintf("%v", val))
			}
		}
		return v
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = s.substituteVariables(val, variables)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = s.substituteVariables(val, variables)
		}
		return result
	default:
		return value
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}