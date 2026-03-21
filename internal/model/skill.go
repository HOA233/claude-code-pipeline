package model

type Skill struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Category    string           `json:"category"`
	Parameters  []SkillParameter `json:"parameters"`
	CLI         *CLIConfig       `json:"cli,omitempty"`
	Prompt      string           `json:"prompt_template"`
	Tags        []string         `json:"tags"`
	Enabled     bool             `json:"enabled"`
}

type SkillParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Values      []string    `json:"values,omitempty"`
}

type CLIConfig struct {
	Model     string `json:"model,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}