package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/config"
	"github.com/company/claude-pipeline/internal/repository"
)

func TestSkillService(t *testing.T) {
	cfg := config.RedisConfig{
		Addr: "localhost:6379",
		DB:   1,
	}

	redisClient := repository.NewRedisClient(cfg)
	ctx := t.Context()

	if err := redisClient.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	svc := NewSkillService(redisClient, config.GitLabConfig{})

	t.Run("SyncFromGitLab", func(t *testing.T) {
		skills, err := svc.SyncFromGitLab(ctx)
		if err != nil {
			t.Errorf("Failed to sync skills: %v", err)
		}

		if len(skills) == 0 {
			t.Error("Expected at least one skill")
		}
	})

	t.Run("GetAllSkills", func(t *testing.T) {
		skills, err := svc.GetAllSkills(ctx)
		if err != nil {
			t.Errorf("Failed to get skills: %v", err)
		}

		if len(skills) == 0 {
			t.Error("Expected at least one skill")
		}
	})

	t.Run("GetSkill", func(t *testing.T) {
		skill, err := svc.GetSkill(ctx, "code-review")
		if err != nil {
			t.Errorf("Failed to get skill: %v", err)
		}

		if skill == nil {
			t.Error("Expected skill to be returned")
		}

		if skill.ID != "code-review" {
			t.Errorf("Expected skill ID 'code-review', got '%s'", skill.ID)
		}
	})
}