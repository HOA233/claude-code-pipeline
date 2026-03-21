package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Redis     RedisConfig     `mapstructure:"redis"`
	GitLab    GitLabConfig    `mapstructure:"gitlab"`
	CLI       CLIConfig       `mapstructure:"cli"`
	Log       LogConfig       `mapstructure:"log"`
	Auth      AuthConfig      `mapstructure:"auth"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Webhook   WebhookConfig   `mapstructure:"webhook"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type GitLabConfig struct {
	URL        string `mapstructure:"url"`
	Token      string `mapstructure:"token"`
	SkillsRepo string `mapstructure:"skills_repo"`
}

type CLIConfig struct {
	MaxConcurrency int           `mapstructure:"max_concurrency"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	ClaudePath     string        `mapstructure:"claude_path"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type AuthConfig struct {
	APIKeys   []string `mapstructure:"api_keys"`
	JWTSecret string   `mapstructure:"jwt_secret"`
}

type RateLimitConfig struct {
	Enabled  bool `mapstructure:"enabled"`
	Requests int  `mapstructure:"requests"`
	Window   int  `mapstructure:"window"`
}

type WebhookConfig struct {
	DefaultURL   string `mapstructure:"default_url"`
	SlackURL     string `mapstructure:"slack_url"`
	SlackChannel string `mapstructure:"slack_channel"`
	NotifyOnFail bool   `mapstructure:"notify_on_fail"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	// Try to read config file, but don't fail if it doesn't exist
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Set defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// LoadFromEnv loads configuration solely from environment variables
func LoadFromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  time.Duration(getEnvInt("SERVER_READ_TIMEOUT", 30)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT", 30)) * time.Second,
			IdleTimeout:  time.Duration(getEnvInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		GitLab: GitLabConfig{
			URL:        getEnv("GITLAB_URL", "https://gitlab.com"),
			Token:      getEnv("GITLAB_TOKEN", ""),
			SkillsRepo: getEnv("GITLAB_SKILLS_REPO", ""),
		},
		CLI: CLIConfig{
			MaxConcurrency: getEnvInt("CLI_MAX_CONCURRENCY", 5),
			DefaultTimeout: time.Duration(getEnvInt("CLI_DEFAULT_TIMEOUT", 600)) * time.Second,
			ClaudePath:     getEnv("CLI_CLAUDE_PATH", "claude"),
			MaxRetries:     getEnvInt("CLI_MAX_RETRIES", 3),
			RetryDelay:     time.Duration(getEnvInt("CLI_RETRY_DELAY", 5)) * time.Second,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Auth: AuthConfig{
			APIKeys:   getEnvSlice("API_KEYS", []string{}),
			JWTSecret: getEnv("JWT_SECRET", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled:  getEnvBool("RATE_LIMIT_ENABLED", true),
			Requests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvInt("RATE_LIMIT_WINDOW", 60),
		},
		Webhook: WebhookConfig{
			DefaultURL:   getEnv("DEFAULT_WEBHOOK_URL", ""),
			SlackURL:     getEnv("SLACK_WEBHOOK_URL", ""),
			SlackChannel: getEnv("SLACK_CHANNEL", ""),
			NotifyOnFail: getEnvBool("WEBHOOK_NOTIFY_ON_FAIL", true),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("ENABLE_METRICS", true),
			Port:    getEnvInt("METRICS_PORT", 9090),
		},
	}
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		c.Redis.Addr = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		// Store for CLI executor
	}
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.CLI.MaxConcurrency == 0 {
		c.CLI.MaxConcurrency = 5
	}
	if c.CLI.DefaultTimeout == 0 {
		c.CLI.DefaultTimeout = 600 * time.Second
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.CLI.MaxConcurrency < 1 {
		return fmt.Errorf("max concurrency must be at least 1")
	}
	return nil
}

// GetRedisDSN returns a Redis connection string
func (c *RedisConfig) GetDSN() string {
	if c.Password != "" {
		return fmt.Sprintf("redis://:%s@%s/%d", c.Password, c.Addr, c.DB)
	}
	return fmt.Sprintf("redis://%s/%d", c.Addr, c.DB)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}