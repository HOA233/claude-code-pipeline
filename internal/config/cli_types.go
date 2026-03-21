package config

import (
	"fmt"
	"strings"
)

// CLITypeConfig defines configuration for a specific CLI type
type CLITypeConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Path        string            `yaml:"path" json:"path"`
	DefaultArgs []string          `yaml:"default_args" json:"default_args"`
	EnvVars     map[string]string `yaml:"env_vars" json:"env_vars"`
	Timeout     int               `yaml:"timeout" json:"timeout"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
}

// DefaultCLITypes returns the default CLI type configurations
func DefaultCLITypes() map[string]*CLITypeConfig {
	return map[string]*CLITypeConfig{
		"claude": {
			Name:    "Claude Code CLI",
			Path:    "claude",
			Enabled: true,
			DefaultArgs: []string{
				"--output-format", "json",
			},
			EnvVars: map[string]string{
				"ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
			},
			Timeout: 600,
		},
		"npm": {
			Name:    "NPM CLI",
			Path:    "npm",
			Enabled: true,
			DefaultArgs: []string{
				"--no-audit",
				"--no-fund",
			},
			Timeout: 300,
		},
		"git": {
			Name:    "Git CLI",
			Path:    "git",
			Enabled: true,
			Timeout: 120,
		},
		"docker": {
			Name:    "Docker CLI",
			Path:    "docker",
			Enabled: true,
			Timeout: 600,
		},
		"kubectl": {
			Name:    "Kubectl CLI",
			Path:    "kubectl",
			Enabled: true,
			Timeout: 300,
		},
		"echo": {
			Name:    "Echo (Test)",
			Path:    "echo",
			Enabled: true,
			Timeout: 10,
		},
		"bash": {
			Name:    "Bash",
			Path:    "bash",
			DefaultArgs: []string{
				"-c",
			},
			Enabled: true,
			Timeout: 300,
		},
		"python": {
			Name:    "Python",
			Path:    "python3",
			Enabled: true,
			Timeout: 600,
		},
		"go": {
			Name:    "Go",
			Path:    "go",
			Enabled: true,
			Timeout: 300,
		},
	}
}

// BuildCommand builds the command arguments for a CLI type
func (c *CLITypeConfig) BuildCommand(action, command string, params map[string]interface{}) []string {
	args := make([]string, 0)

	// Add default args
	args = append(args, c.DefaultArgs...)

	// Add action
	if action != "" {
		args = append(args, action)
	}

	// Add command
	if command != "" {
		args = append(args, command)
	}

	// Add params as flags
	for key, value := range params {
		switch v := value.(type) {
		case bool:
			if v {
				args = append(args, fmt.Sprintf("--%s", key))
			}
		case string:
			args = append(args, fmt.Sprintf("--%s", key), v)
		case int, int64, float64:
			args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", v))
		case []interface{}:
			for _, item := range v {
				args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", item))
			}
		default:
			args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", v))
		}
	}

	return args
}

// Validate validates the CLI type configuration
func (c *CLITypeConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("CLI type name is required")
	}
	if c.Path == "" {
		return fmt.Errorf("CLI path is required")
	}
	if c.Timeout <= 0 {
		c.Timeout = 300 // Default timeout
	}
	return nil
}

// CLITypesManager manages CLI type configurations
type CLITypesManager struct {
	types map[string]*CLITypeConfig
}

// NewCLITypesManager creates a new CLI types manager
func NewCLITypesManager() *CLITypesManager {
	return &CLITypesManager{
		types: DefaultCLITypes(),
	}
}

// Get returns a CLI type configuration
func (m *CLITypesManager) Get(name string) (*CLITypeConfig, error) {
	name = strings.ToLower(name)
	cfg, exists := m.types[name]
	if !exists {
		return nil, fmt.Errorf("CLI type not found: %s", name)
	}
	if !cfg.Enabled {
		return nil, fmt.Errorf("CLI type is disabled: %s", name)
	}
	return cfg, nil
}

// List returns all CLI type configurations
func (m *CLITypesManager) List() map[string]*CLITypeConfig {
	return m.types
}

// Register registers a new CLI type
func (m *CLITypesManager) Register(name string, cfg *CLITypeConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	m.types[strings.ToLower(name)] = cfg
	return nil
}

// Enable enables a CLI type
func (m *CLITypesManager) Enable(name string) error {
	cfg, exists := m.types[strings.ToLower(name)]
	if !exists {
		return fmt.Errorf("CLI type not found: %s", name)
	}
	cfg.Enabled = true
	return nil
}

// Disable disables a CLI type
func (m *CLITypesManager) Disable(name string) error {
	cfg, exists := m.types[strings.ToLower(name)]
	if !exists {
		return fmt.Errorf("CLI type not found: %s", name)
	}
	cfg.Enabled = false
	return nil
}