package gitlab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

// Client for GitLab API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitLab client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Repository represents a GitLab repository
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	WebURL        string `json:"web_url"`
	DefaultBranch string `json:"default_branch"`
}

// File represents a file in GitLab
type File struct {
	FileName     string `json:"file_name"`
	FilePath     string `json:"file_path"`
	Size         int    `json:"size"`
	Encoding     string `json:"encoding"`
	Content      string `json:"content"`
	Ref          string `json:"ref"`
	BlobID       string `json:"blob_id"`
	CommitID     string `json:"commit_id"`
	LastCommitID string `json:"last_commit_id"`
}

// SkillFile represents a skill configuration file
type SkillFile struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Category    string                 `yaml:"category"`
	Author      string                 `yaml:"author"`
	CLI         *CLIConfig             `yaml:"cli"`
	Parameters  []SkillParameter       `yaml:"parameters"`
	Permissions []string               `yaml:"permissions"`
	Output      map[string]interface{} `yaml:"output"`
}

type CLIConfig struct {
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
	Timeout   int    `yaml:"timeout"`
}

type SkillParameter struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Required    bool        `yaml:"required"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Values      []string    `yaml:"values"`
	Validation  struct {
		Pattern string `yaml:"pattern"`
	} `yaml:"validation"`
}

// GetProject gets a project by ID or path
func (c *Client) GetProject(ctx context.Context, projectPath string) (*Repository, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s", c.baseURL, encodeProjectPath(projectPath))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get project: status %d", resp.StatusCode)
	}

	var repo Repository
	if err := decodeJSON(resp.Body, &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetFile gets a file from a repository
func (c *Client) GetFile(ctx context.Context, projectPath, filePath, ref string) (*File, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s?ref=%s",
		c.baseURL, encodeProjectPath(projectPath), encodeFilePath(filePath), ref)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file: status %d", resp.StatusCode)
	}

	var file File
	if err := decodeJSON(resp.Body, &file); err != nil {
		return nil, err
	}

	return &file, nil
}

// GetSkill loads a skill from a GitLab repository
func (c *Client) GetSkill(ctx context.Context, projectPath, skillPath, ref string) (*SkillFile, string, error) {
	// Get skill.yaml
	skillFile, err := c.GetFile(ctx, projectPath, skillPath+"/skill.yaml", ref)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get skill.yaml: %w", err)
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(skillFile.Content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode skill.yaml: %w", err)
	}

	var skill SkillFile
	if err := yaml.Unmarshal(content, &skill); err != nil {
		return nil, "", fmt.Errorf("failed to parse skill.yaml: %w", err)
	}

	// Get prompt.md
	promptContent := ""
	promptFile, err := c.GetFile(ctx, projectPath, skillPath+"/prompt.md", ref)
	if err == nil {
		promptBytes, err := base64.StdEncoding.DecodeString(promptFile.Content)
		if err == nil {
			promptContent = string(promptBytes)
		}
	}

	return &skill, promptContent, nil
}

// ListFiles lists files in a repository directory
func (c *Client) ListFiles(ctx context.Context, projectPath, path, ref string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?path=%s&ref=%s",
		c.baseURL, encodeProjectPath(projectPath), path, ref)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list files: status %d", resp.StatusCode)
	}

	var files []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := decodeJSON(resp.Body, &files); err != nil {
		return nil, err
	}

	var paths []string
	for _, f := range files {
		if f.Type == "tree" {
			paths = append(paths, f.Path)
		}
	}

	return paths, nil
}

// Helper functions

func encodeProjectPath(path string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(path))
}

func encodeFilePath(path string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(path))
}

func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}