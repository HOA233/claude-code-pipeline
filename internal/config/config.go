package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	GitLab GitLabConfig `mapstructure:"gitlab"`
	CLI    CLIConfig    `mapstructure:"cli"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
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
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}