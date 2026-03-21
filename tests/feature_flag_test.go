package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestFeatureFlagService_New(t *testing.T) {
	ff := service.NewFeatureFlagService()
	if ff == nil {
		t.Fatal("Expected non-nil feature flag service")
	}
}

func TestFeatureFlagService_CreateFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	err := ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:   "test-flag",
		Name:  "Test Flag",
		Type:  service.FlagTypeBoolean,
		Default: false,
	})

	if err != nil {
		t.Fatalf("Failed to create flag: %v", err)
	}
}

func TestFeatureFlagService_CreateFlag_MissingKey(t *testing.T) {
	ff := service.NewFeatureFlagService()

	err := ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Name: "No Key",
	})

	if err == nil {
		t.Error("Expected error for missing key")
	}
}

func TestFeatureFlagService_CreateFlag_MissingName(t *testing.T) {
	ff := service.NewFeatureFlagService()

	err := ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key: "no-name",
	})

	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestFeatureFlagService_CreateFlag_DuplicateKey(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "duplicate",
		Name: "First",
	})

	err := ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "duplicate",
		Name: "Second",
	})

	if err == nil {
		t.Error("Expected error for duplicate key")
	}
}

func TestFeatureFlagService_GetFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	flag := &service.FeatureFlag{
		ID:   "flag-1",
		Key:  "get-test",
		Name: "Get Test",
		Type: service.FlagTypeBoolean,
	}
	ff.CreateFlag(context.Background(), flag)

	// Get by ID
	retrieved, err := ff.GetFlag("flag-1")
	if err != nil {
		t.Fatalf("Failed to get flag by ID: %v", err)
	}

	if retrieved.Key != "get-test" {
		t.Error("Flag key mismatch")
	}

	// Get by key
	retrieved, err = ff.GetFlag("get-test")
	if err != nil {
		t.Fatalf("Failed to get flag by key: %v", err)
	}
}

func TestFeatureFlagService_GetFlag_NotFound(t *testing.T) {
	ff := service.NewFeatureFlagService()

	_, err := ff.GetFlag("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent flag")
	}
}

func TestFeatureFlagService_UpdateFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	flag := &service.FeatureFlag{
		ID:   "update-test",
		Key:  "update",
		Name: "Original",
		Type: service.FlagTypeBoolean,
	}
	ff.CreateFlag(context.Background(), flag)

	updated, err := ff.UpdateFlag(context.Background(), "update-test", map[string]interface{}{
		"name":        "Updated",
		"description": "New description",
	})

	if err != nil {
		t.Fatalf("Failed to update flag: %v", err)
	}

	if updated.Name != "Updated" {
		t.Error("Name not updated")
	}
}

func TestFeatureFlagService_DeleteFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	flag := &service.FeatureFlag{
		ID:  "delete-test",
		Key: "delete",
		Name: "Delete",
	}
	ff.CreateFlag(context.Background(), flag)

	err := ff.DeleteFlag("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete flag: %v", err)
	}

	_, err = ff.GetFlag("delete-test")
	if err == nil {
		t.Error("Expected error for deleted flag")
	}
}

func TestFeatureFlagService_ToggleFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	flag := &service.FeatureFlag{
		ID:      "toggle-test",
		Key:     "toggle",
		Name:    "Toggle",
		Enabled: false,
	}
	ff.CreateFlag(context.Background(), flag)

	ff.ToggleFlag("toggle-test", true)

	retrieved, _ := ff.GetFlag("toggle-test")
	if !retrieved.Enabled {
		t.Error("Expected flag to be enabled")
	}
}

func TestFeatureFlagService_Evaluate_Disabled(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "disabled-flag",
		Name:    "Disabled",
		Enabled: false,
		Default: false,
	})

	result, err := ff.Evaluate(context.Background(), "disabled-flag", &service.EvaluationContext{
		UserID: "user-1",
	})

	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result.Value != false {
		t.Error("Expected default value for disabled flag")
	}
	if result.Reason != "flag_disabled" {
		t.Errorf("Expected reason 'flag_disabled', got '%s'", result.Reason)
	}
}

func TestFeatureFlagService_Evaluate_Enabled(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "enabled-flag",
		Name:    "Enabled",
		Enabled: true,
		Type:    service.FlagTypeBoolean,
		Default: false,
		Variations: []service.Variation{
			{ID: "true", Name: "True", Value: true},
			{ID: "false", Name: "False", Value: false},
		},
	})

	result, err := ff.Evaluate(context.Background(), "enabled-flag", &service.EvaluationContext{
		UserID: "user-1",
	})

	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result.Reason != "default" {
		t.Errorf("Expected reason 'default', got '%s'", result.Reason)
	}
}

func TestFeatureFlagService_Evaluate_TargetingRule(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "targeted-flag",
		Name:    "Targeted",
		Enabled: true,
		Type:    service.FlagTypeBoolean,
		Default: false,
		Variations: []service.Variation{
			{ID: "true", Name: "True", Value: true},
			{ID: "false", Name: "False", Value: false},
		},
	})

	// Add targeting rule for specific user
	ff.AddTargetingRule(&service.TargetingRule{
		FlagKey: "targeted-flag",
		Conditions: []service.Condition{
			{Attribute: "user_id", Operator: "eq", Value: "vip-user"},
		},
		Variation: "true",
		Enabled:   true,
	})

	// Evaluate for VIP user
	result, err := ff.Evaluate(context.Background(), "targeted-flag", &service.EvaluationContext{
		UserID: "vip-user",
	})

	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result.Value != true {
		t.Error("Expected true for VIP user")
	}
	if result.Reason != "targeting_rule" {
		t.Errorf("Expected reason 'targeting_rule', got '%s'", result.Reason)
	}

	// Evaluate for non-VIP user
	result2, _ := ff.Evaluate(context.Background(), "targeted-flag", &service.EvaluationContext{
		UserID: "regular-user",
	})

	if result2.Value != false {
		t.Error("Expected default for regular user")
	}
}

func TestFeatureFlagService_EvaluateAll(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "flag1",
		Name:    "Flag 1",
		Enabled: true,
		Default: true,
	})

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "flag2",
		Name:    "Flag 2",
		Enabled: false,
		Default: false,
	})

	results := ff.EvaluateAll(context.Background(), &service.EvaluationContext{
		UserID: "user-1",
	})

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}
}

func TestFeatureFlagService_AddTargetingRule(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "rule-test",
		Name: "Rule Test",
	})

	rule := &service.TargetingRule{
		FlagKey: "rule-test",
		Conditions: []service.Condition{
			{Attribute: "tenant_id", Operator: "eq", Value: "tenant-1"},
		},
		Variation: "variation-1",
		Enabled:   true,
	}

	err := ff.AddTargetingRule(rule)
	if err != nil {
		t.Fatalf("Failed to add targeting rule: %v", err)
	}
}

func TestFeatureFlagService_RemoveTargetingRule(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "remove-rule-test",
		Name: "Remove Rule Test",
	})

	rule := &service.TargetingRule{
		ID:      "rule-1",
		FlagKey: "remove-rule-test",
	}
	ff.AddTargetingRule(rule)

	err := ff.RemoveTargetingRule("remove-rule-test", "rule-1")
	if err != nil {
		t.Fatalf("Failed to remove rule: %v", err)
	}
}

func TestFeatureFlagService_ListFlags(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "list1",
		Name: "List 1",
		Tags: []string{"test"},
	})

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "list2",
		Name: "List 2",
		Tags: []string{"test"},
	})

	flags := ff.ListFlags("test")
	if len(flags) < 2 {
		t.Errorf("Expected at least 2 flags, got %d", len(flags))
	}
}

func TestFeatureFlagService_GetStats(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:  "stats-test",
		Name: "Stats Test",
	})

	// Evaluate a few times
	ff.Evaluate(context.Background(), "stats-test", &service.EvaluationContext{UserID: "user-1"})
	ff.Evaluate(context.Background(), "stats-test", &service.EvaluationContext{UserID: "user-2"})

	stats, err := ff.GetStats("stats-test")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.Evaluations < 2 {
		t.Errorf("Expected at least 2 evaluations, got %d", stats.Evaluations)
	}
}

func TestFeatureFlagService_MultivariateFlag(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "multivariate",
		Name:    "Multivariate Test",
		Type:    service.FlagTypeMultivariate,
		Enabled: true,
		Default: "control",
		Variations: []service.Variation{
			{ID: "control", Name: "Control", Value: "control"},
			{ID: "variant-a", Name: "Variant A", Value: "variant-a"},
			{ID: "variant-b", Name: "Variant B", Value: "variant-b"},
		},
	})

	result, err := ff.Evaluate(context.Background(), "multivariate", &service.EvaluationContext{
		UserID: "user-1",
	})

	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result.Value != "control" {
		t.Errorf("Expected 'control', got '%v'", result.Value)
	}
}

func TestFeatureFlagService_Conditions(t *testing.T) {
	ff := service.NewFeatureFlagService()

	ff.CreateFlag(context.Background(), &service.FeatureFlag{
		Key:     "cond-test",
		Name:    "Condition Test",
		Enabled: true,
		Default: false,
		Variations: []service.Variation{
			{ID: "true", Name: "True", Value: true},
		},
	})

	// Test "in" operator
	ff.AddTargetingRule(&service.TargetingRule{
		FlagKey: "cond-test",
		Conditions: []service.Condition{
			{Attribute: "user_id", Operator: "in", Value: []interface{}{"user1", "user2", "user3"}},
		},
		Variation: "true",
		Enabled:   true,
	})

	result, _ := ff.Evaluate(context.Background(), "cond-test", &service.EvaluationContext{
		UserID: "user2",
	})

	if result.Value != true {
		t.Error("Expected true for user in list")
	}
}