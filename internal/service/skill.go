package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
	"gopkg.in/yaml.v3"
)

type SkillService struct {
	redis   *repository.RedisClient
	gitlab  config.GitLabConfig
	http    *http.Client
}

func NewSkillService(redis *repository.RedisClient, gitlab config.GitLabConfig) *SkillService {
	return &SkillService{
		redis:  redis,
		gitlab: gitlab,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SkillService) GetAllSkills(ctx context.Context) ([]*model.Skill, error) {
	return s.redis.GetAllSkills(ctx)
}

func (s *SkillService) GetSkill(ctx context.Context, skillID string) (*model.Skill, error) {
	return s.redis.GetSkill(ctx, skillID)
}

func (s *SkillService) CreateSkill(ctx context.Context, skill *model.Skill) error {
	return s.redis.SaveSkill(ctx, skill)
}

// SyncFromGitLab syncs skills from GitLab repository
func (s *SkillService) SyncFromGitLab(ctx context.Context) ([]*model.Skill, error) {
	// For demo purposes, create sample skills
	skills := s.getDefaultSkills()

	for _, skill := range skills {
		if err := s.redis.SaveSkill(ctx, skill); err != nil {
			logger.Errorf("Failed to save skill %s: %v", skill.ID, err)
			continue
		}
	}

	logger.Infof("Synced %d skills", len(skills))
	return skills, nil
}

func (s *SkillService) getDefaultSkills() []*model.Skill {
	return []*model.Skill{
		{
			ID:          "code-review",
			Name:        "代码审查",
			Description: "分析代码质量，检测安全漏洞和性能问题",
			Version:     "1.0.0",
			Category:    "quality",
			Enabled:     true,
			Tags:        []string{"quality", "security", "performance"},
			Parameters: []model.SkillParameter{
				{Name: "target", Type: "string", Required: true, Description: "审查目标路径"},
				{Name: "depth", Type: "enum", Required: false, Default: "standard", Values: []string{"quick", "standard", "deep"}, Description: "审查深度"},
			},
			CLI: &model.CLIConfig{Model: "claude-sonnet-4-6", MaxTokens: 8192, Timeout: 600},
			Prompt: `你是一个专业的代码审查专家。请对以下代码进行全面审查。

## 审查目标
- 路径: {{target}}
- 深度: {{depth}}

## 审查维度
1. 代码质量
2. 安全性
3. 性能
4. 可维护性

请输出 JSON 格式的审查结果。`,
		},
		{
			ID:          "deploy",
			Name:        "部署服务",
			Description: "自动化部署到指定环境",
			Version:     "1.0.0",
			Category:    "devops",
			Enabled:     true,
			Tags:        []string{"deploy", "automation"},
			Parameters: []model.SkillParameter{
				{Name: "environment", Type: "enum", Required: true, Values: []string{"dev", "staging", "production"}, Description: "部署环境"},
				{Name: "dry_run", Type: "boolean", Required: false, Default: false, Description: "是否模拟执行"},
			},
			CLI: &model.CLIConfig{Model: "claude-sonnet-4-6", MaxTokens: 4096, Timeout: 300},
			Prompt: `执行部署任务到 {{environment}} 环境。
模拟模式: {{dry_run}}

请输出部署步骤和结果。`,
		},
		{
			ID:          "test-gen",
			Name:        "测试生成",
			Description: "自动生成单元测试",
			Version:     "1.0.0",
			Category:    "testing",
			Enabled:     true,
			Tags:        []string{"test", "automation"},
			Parameters: []model.SkillParameter{
				{Name: "source", Type: "string", Required: true, Description: "源代码路径"},
				{Name: "framework", Type: "enum", Required: false, Default: "jest", Values: []string{"jest", "pytest", "go-test"}, Description: "测试框架"},
			},
			CLI: &model.CLIConfig{Model: "claude-sonnet-4-6", MaxTokens: 8192, Timeout: 600},
			Prompt: `为以下代码生成单元测试。

源代码路径: {{source}}
测试框架: {{framework}}

请输出完整的测试代码。`,
		},
		{
			ID:          "refactor",
			Name:        "代码重构",
			Description: "智能重构代码结构",
			Version:     "1.0.0",
			Category:    "development",
			Enabled:     true,
			Tags:        []string{"refactor", "quality"},
			Parameters: []model.SkillParameter{
				{Name: "target", Type: "string", Required: true, Description: "重构目标"},
				{Name: "type", Type: "enum", Required: false, Default: "general", Values: []string{"general", "performance", "readability"}, Description: "重构类型"},
			},
			CLI: &model.CLIConfig{Model: "claude-sonnet-4-6", MaxTokens: 8192, Timeout: 600},
			Prompt: `重构以下代码。

目标: {{target}}
重构类型: {{type}}

请输出重构后的代码和改进说明。`,
		},
	}
}

// LoadFromFile loads a skill from a YAML file
func (s *SkillService) LoadFromFile(path string) (*model.Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	var skill model.Skill
	if err := yaml.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("failed to parse skill file: %w", err)
	}

	// Load prompt template if exists
	promptFile := filepath.Join(filepath.Dir(path), "prompt.md")
	if promptData, err := os.ReadFile(promptFile); err == nil {
		skill.Prompt = string(promptData)
	}

	return &skill, nil
}

// FetchSkillFromGitLab fetches a skill from GitLab API
func (s *SkillService) FetchSkillFromGitLab(ctx context.Context, repoPath string) (*model.Skill, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s/raw?ref=main",
		s.gitlab.URL, s.gitlab.SkillsRepo, repoPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", s.gitlab.Token)

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch skill: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var skill model.Skill
	if err := yaml.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("failed to parse skill: %w", err)
	}

	return &skill, nil
}

// ValidateParameters validates skill parameters
func (s *SkillService) ValidateParameters(skill *model.Skill, params map[string]interface{}) error {
	for _, param := range skill.Parameters {
		value, exists := params[param.Name]

		if param.Required && !exists {
			return fmt.Errorf("missing required parameter: %s", param.Name)
		}

		if param.Type == "enum" && len(param.Values) > 0 && exists {
			strValue := fmt.Sprintf("%v", value)
			valid := false
			for _, v := range param.Values {
				if v == strValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid value for parameter %s: %v (allowed: %v)", param.Name, value, param.Values)
			}
		}
	}
	return nil
}

// ToJSON converts skill to JSON
func (s *SkillService) ToJSON(skill *model.Skill) (json.RawMessage, error) {
	return json.Marshal(skill)
}