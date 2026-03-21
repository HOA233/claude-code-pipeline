package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestPipelineBuilder_New(t *testing.T) {
	builder := service.NewPipelineBuilder()
	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

func TestPipelineBuilder_CompletePipeline(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("test-pipeline").
		WithName("Test Pipeline").
		WithDescription("A test pipeline").
		WithVersion("1.0.0").
		WithMode("serial").
		WithTimeout(300).
		AddClaudeStep("analyze", "analyze", map[string]interface{}{
			"target": "src/",
		}).
		AddClaudeStep("review", "review", map[string]interface{}{
			"target": "src/",
		}).
		AddVariable("env", "development").
		WithMetadata("author", "test").
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if pipeline.ID != "test-pipeline" {
		t.Errorf("Expected ID 'test-pipeline', got '%s'", pipeline.ID)
	}
	if pipeline.Name != "Test Pipeline" {
		t.Errorf("Expected name 'Test Pipeline', got '%s'", pipeline.Name)
	}
	if pipeline.Mode != "serial" {
		t.Errorf("Expected mode 'serial', got '%s'", pipeline.Mode)
	}
	if len(pipeline.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(pipeline.Steps))
	}
	if pipeline.Variables["env"] != "development" {
		t.Error("Expected variable 'env' to be 'development'")
	}
}

func TestPipelineBuilder_MissingID(t *testing.T) {
	_, err := service.NewPipelineBuilder().
		WithName("Test Pipeline").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestPipelineBuilder_MissingName(t *testing.T) {
	_, err := service.NewPipelineBuilder().
		WithID("test").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestPipelineBuilder_NoSteps(t *testing.T) {
	_, err := service.NewPipelineBuilder().
		WithID("test").
		WithName("Test").
		Build()

	if err == nil {
		t.Error("Expected error for no steps")
	}
}

func TestPipelineBuilder_InvalidMode(t *testing.T) {
	_, err := service.NewPipelineBuilder().
		WithID("test").
		WithName("Test").
		WithMode("invalid-mode").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	if err == nil {
		t.Error("Expected error for invalid mode")
	}
}

func TestPipelineBuilder_WithDependencies(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("dep-test").
		WithName("Dependency Test").
		WithMode("hybrid").
		AddClaudeStep("step1", "analyze", nil).
		AddStep(service.StepDefinition{
			ID:        "step2",
			CLI:       "claude",
			Action:    "review",
			DependsOn: []string{"step1"},
		}).
		AddStep(service.StepDefinition{
			ID:        "step3",
			CLI:       "claude",
			Action:    "report",
			DependsOn: []string{"step1"},
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if len(pipeline.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(pipeline.Steps))
	}

	// Check dependencies
	if len(pipeline.Steps[1].DependsOn) != 1 || pipeline.Steps[1].DependsOn[0] != "step1" {
		t.Error("Step 2 should depend on step1")
	}
}

func TestPipelineBuilder_InvalidDependency(t *testing.T) {
	_, err := service.NewPipelineBuilder().
		WithID("invalid-dep").
		WithName("Invalid Dependency").
		AddStep(service.StepDefinition{
			ID:        "step1",
			CLI:       "claude",
			Action:    "analyze",
			DependsOn: []string{"nonexistent"},
		}).
		Build()

	if err == nil {
		t.Error("Expected error for invalid dependency")
	}
}

func TestPipelineBuilder_GitStep(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("git-test").
		WithName("Git Test").
		AddGitStep("clone", "clone", map[string]interface{}{
			"repo": "https://github.com/test/repo.git",
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if pipeline.Steps[0].CLI != "git" {
		t.Errorf("Expected CLI 'git', got '%s'", pipeline.Steps[0].CLI)
	}
}

func TestPipelineBuilder_NPMStep(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("npm-test").
		WithName("NPM Test").
		AddNPMStep("install", "install", nil).
		AddNPMStep("test", "test", nil).
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if len(pipeline.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(pipeline.Steps))
	}
}

func TestPipelineBuilder_DockerStep(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("docker-test").
		WithName("Docker Test").
		AddDockerStep("build", "build", map[string]interface{}{
			"tag": "myapp:latest",
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if pipeline.Steps[0].CLI != "docker" {
		t.Errorf("Expected CLI 'docker', got '%s'", pipeline.Steps[0].CLI)
	}
}

func TestPipelineBuilder_Triggers(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("trigger-test").
		WithName("Trigger Test").
		AddClaudeStep("step1", "analyze", nil).
		AddWebhookTrigger("/webhook/deploy").
		AddScheduleTrigger("0 * * * *").
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if len(pipeline.Triggers) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(pipeline.Triggers))
	}
}

func TestPipelineBuilder_Defaults(t *testing.T) {
	pipeline, err := service.NewPipelineBuilder().
		WithID("defaults-test").
		WithName("Defaults Test").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	if err != nil {
		t.Fatalf("Failed to build pipeline: %v", err)
	}

	if pipeline.Mode != "serial" {
		t.Errorf("Expected default mode 'serial', got '%s'", pipeline.Mode)
	}
	if pipeline.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", pipeline.Version)
	}
}

func TestPipelineRegistry_Register(t *testing.T) {
	registry := service.NewPipelineRegistry()

	pipeline, _ := service.NewPipelineBuilder().
		WithID("registry-test").
		WithName("Registry Test").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	err := registry.Register(pipeline)
	if err != nil {
		t.Fatalf("Failed to register pipeline: %v", err)
	}
}

func TestPipelineRegistry_Get(t *testing.T) {
	registry := service.NewPipelineRegistry()

	pipeline, _ := service.NewPipelineBuilder().
		WithID("get-test").
		WithName("Get Test").
		AddClaudeStep("step1", "analyze", nil).
		Build()
	registry.Register(pipeline)

	retrieved, err := registry.Get("get-test")
	if err != nil {
		t.Fatalf("Failed to get pipeline: %v", err)
	}

	if retrieved.ID != "get-test" {
		t.Error("Retrieved pipeline ID mismatch")
	}
}

func TestPipelineRegistry_GetNotFound(t *testing.T) {
	registry := service.NewPipelineRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent pipeline")
	}
}

func TestPipelineRegistry_List(t *testing.T) {
	registry := service.NewPipelineRegistry()

	for i := 0; i < 3; i++ {
		pipeline, _ := service.NewPipelineBuilder().
			WithID(string(rune('a' + i))).
			WithName("Pipeline").
			AddClaudeStep("step1", "analyze", nil).
			Build()
		registry.Register(pipeline)
	}

	pipelines := registry.List()
	if len(pipelines) < 3 {
		t.Errorf("Expected at least 3 pipelines, got %d", len(pipelines))
	}
}

func TestPipelineRegistry_Delete(t *testing.T) {
	registry := service.NewPipelineRegistry()

	pipeline, _ := service.NewPipelineBuilder().
		WithID("delete-test").
		WithName("Delete Test").
		AddClaudeStep("step1", "analyze", nil).
		Build()
	registry.Register(pipeline)

	registry.Delete("delete-test")

	_, err := registry.Get("delete-test")
	if err == nil {
		t.Error("Expected error for deleted pipeline")
	}
}

func TestPipelineDefinition_ToJSON(t *testing.T) {
	pipeline, _ := service.NewPipelineBuilder().
		WithID("json-test").
		WithName("JSON Test").
		AddClaudeStep("step1", "analyze", nil).
		Build()

	data, err := pipeline.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestPipelineDefinition_Validate(t *testing.T) {
	pipeline := &service.PipelineDefinition{
		ID:   "validate-test",
		Name: "Validate Test",
		Steps: []service.StepDefinition{
			{ID: "step1", CLI: "claude", Action: "analyze"},
		},
	}

	err := pipeline.Validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
}

func TestPipelineDefinition_ValidateMissingID(t *testing.T) {
	pipeline := &service.PipelineDefinition{
		Name: "No ID",
		Steps: []service.StepDefinition{
			{ID: "step1", CLI: "claude"},
		},
	}

	err := pipeline.Validate()
	if err == nil {
		t.Error("Expected validation error for missing ID")
	}
}

func TestPipelineDefinition_ValidateDuplicateStepID(t *testing.T) {
	pipeline := &service.PipelineDefinition{
		ID:   "duplicate-test",
		Name: "Duplicate Step ID",
		Steps: []service.StepDefinition{
			{ID: "step1", CLI: "claude"},
			{ID: "step1", CLI: "claude"},
		},
	}

	err := pipeline.Validate()
	if err == nil {
		t.Error("Expected validation error for duplicate step ID")
	}
}

func TestGetTemplate(t *testing.T) {
	template, err := service.GetTemplate("code-review")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if template.ID != "code-review" {
		t.Errorf("Expected template ID 'code-review', got '%s'", template.ID)
	}
}

func TestGetTemplateNotFound(t *testing.T) {
	_, err := service.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
}

func TestListTemplates(t *testing.T) {
	templates := service.ListTemplates()

	if len(templates) == 0 {
		t.Error("Expected at least one template")
	}
}

func TestBuildFromTemplate(t *testing.T) {
	pipeline, err := service.BuildFromTemplate("code-review", map[string]interface{}{
		"target": "src/",
		"depth":  "deep",
	})

	if err != nil {
		t.Fatalf("Failed to build from template: %v", err)
	}

	if pipeline == nil {
		t.Fatal("Expected non-nil pipeline")
	}

	if len(pipeline.Steps) == 0 {
		t.Error("Expected at least one step from template")
	}
}