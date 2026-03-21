package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestTracingService_CreateNoop(t *testing.T) {
	cfg := &service.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	ts, err := service.NewTracingService(cfg)
	if err != nil {
		t.Fatalf("Failed to create tracing service: %v", err)
	}

	if ts == nil {
		t.Fatal("Expected non-nil tracing service")
	}
}

func TestTracingService_TaskSpan(t *testing.T) {
	ts, _ := service.NewTracingService(&service.TracingConfig{Enabled: false})

	ctx, span := ts.TraceTask(nil, "task-001", "code-review")
	if span == nil {
		t.Fatal("Expected non-nil task span")
	}

	span.SetStatus("running")
	span.SetProgress(1, 5)
	span.End()
}

func TestTracingService_PipelineSpan(t *testing.T) {
	ts, _ := service.NewTracingService(&service.TracingConfig{Enabled: false})

	ctx, span := ts.TracePipeline(nil, "pipeline-001", "serial")
	if span == nil {
		t.Fatal("Expected non-nil pipeline span")
	}

	span.AddStep("step1", "completed", 1000)
	span.SetStatus("running")
	span.End()
}

func TestTracingService_StepSpan(t *testing.T) {
	ts, _ := service.NewTracingService(&service.TracingConfig{Enabled: false})

	ctx, span := ts.TraceStep(nil, "step-001", "claude")
	if span == nil {
		t.Fatal("Expected non-nil step span")
	}

	span.SetAction("analyze")
	span.SetStatus("completed")
	span.End()
}

func TestGitOpsManager_ValidatePipeline_ValidModes(t *testing.T) {
	gm := service.NewGitOpsManager()

	modes := []string{"serial", "parallel", "hybrid"}

	for _, mode := range modes {
		pipeline := &service.PipelineDefinition{
			Name: "test-pipeline",
			Mode: mode,
			Steps: []service.StepDefinition{
				{ID: "step1", CLI: "claude"},
			},
		}

		err := gm.ValidatePipeline(pipeline)
		if err != nil {
			t.Errorf("Mode '%s' should be valid: %v", mode, err)
		}
	}
}

func TestTenantService_DefaultPlan(t *testing.T) {
	ts := service.NewTenantService()

	tenant, err := ts.CreateTenant(nil, &service.TenantCreateRequest{
		Name: "Default Plan Test",
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	if tenant.Plan != "free" {
		t.Errorf("Expected default plan 'free', got '%s'", tenant.Plan)
	}
}

func TestTenantService_PlanQuotas(t *testing.T) {
	ts := service.NewTenantService()

	planTests := []struct {
		plan           string
		expectedTasks  int
		expectedConcur int
	}{
		{"free", 100, 2},
		{"pro", 1000, 10},
		{"enterprise", 10000, 50},
	}

	for _, tt := range planTests {
		t.Run(tt.plan, func(t *testing.T) {
			tenant, err := ts.CreateTenant(nil, &service.TenantCreateRequest{
				Name: "Plan Test " + tt.plan,
				Plan: tt.plan,
			})
			if err != nil {
				t.Fatalf("Failed to create tenant: %v", err)
			}

			if tenant.Quotas.MaxTasks != tt.expectedTasks {
				t.Errorf("Expected MaxTasks %d, got %d", tt.expectedTasks, tenant.Quotas.MaxTasks)
			}

			if tenant.Quotas.MaxConcurrent != tt.expectedConcur {
				t.Errorf("Expected MaxConcurrent %d, got %d", tt.expectedConcur, tenant.Quotas.MaxConcurrent)
			}
		})
	}
}

func TestTenantService_UsageTracking(t *testing.T) {
	ts := service.NewTenantService()

	tenant, _ := ts.CreateTenant(nil, &service.TenantCreateRequest{
		Name: "Usage Test",
		Plan: "pro",
	})

	// Increment usage
	ts.IncrementUsage(nil, tenant.ID, "tasks", 5)
	ts.IncrementUsage(nil, tenant.ID, "pipelines", 2)
	ts.IncrementUsage(nil, tenant.ID, "concurrent", 1)

	// Get updated tenant
	updated, _ := ts.GetTenant(nil, tenant.ID)

	if updated.Usage.Tasks != 5 {
		t.Errorf("Expected tasks usage 5, got %d", updated.Usage.Tasks)
	}

	if updated.Usage.Pipelines != 2 {
		t.Errorf("Expected pipelines usage 2, got %d", updated.Usage.Pipelines)
	}

	if updated.Usage.Concurrent != 1 {
		t.Errorf("Expected concurrent usage 1, got %d", updated.Usage.Concurrent)
	}
}