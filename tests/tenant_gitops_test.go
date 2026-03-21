package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestTenantService_CreateTenant(t *testing.T) {
	ts := service.NewTenantService()

	tests := []struct {
		name    string
		req     *service.TenantCreateRequest
		wantErr bool
	}{
		{
			name: "valid tenant",
			req: &service.TenantCreateRequest{
				Name: "Test Tenant",
				Slug: "test-tenant",
				Plan: "pro",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: &service.TenantCreateRequest{
				Slug: "no-name",
				Plan: "free",
			},
			wantErr: true,
		},
		{
			name: "auto-generate slug",
			req: &service.TenantCreateRequest{
				Name: "Auto Slug Tenant",
				Plan: "free",
			},
			wantErr: false,
		},
		{
			name: "invalid plan",
			req: &service.TenantCreateRequest{
				Name: "Invalid Plan",
				Slug: "invalid-plan",
				Plan: "premium",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := ts.CreateTenant(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tenant == nil {
				t.Error("Expected non-nil tenant")
			}
		})
	}
}

func TestTenantService_GetTenant(t *testing.T) {
	ts := service.NewTenantService()

	// Create a tenant
	created, err := ts.CreateTenant(context.Background(), &service.TenantCreateRequest{
		Name: "Test",
		Slug: "test",
		Plan: "pro",
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Get the tenant
	tenant, err := ts.GetTenant(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if tenant.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, tenant.ID)
	}

	if tenant.Plan != "pro" {
		t.Errorf("Expected plan 'pro', got %s", tenant.Plan)
	}
}

func TestTenantService_CheckQuota(t *testing.T) {
	ts := service.NewTenantService()

	// Create tenant with free plan
	tenant, err := ts.CreateTenant(context.Background(), &service.TenantCreateRequest{
		Name: "Quota Test",
		Slug: "quota-test",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Check initial quota
	if err := ts.CheckQuota(context.Background(), tenant.ID, "tasks"); err != nil {
		t.Errorf("Expected quota check to pass: %v", err)
	}

	// Increment usage to limit
	ts.IncrementUsage(context.Background(), tenant.ID, "tasks", 100)

	// Check quota exceeded
	if err := ts.CheckQuota(context.Background(), tenant.ID, "tasks"); err == nil {
		t.Error("Expected quota exceeded error")
	}
}

func TestTenantService_UpdateTenant(t *testing.T) {
	ts := service.NewTenantService()

	tenant, err := ts.CreateTenant(context.Background(), &service.TenantCreateRequest{
		Name: "Update Test",
		Slug: "update-test",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Update to pro plan
	err = ts.UpdateTenant(context.Background(), tenant.ID, map[string]interface{}{
		"plan": "pro",
	})
	if err != nil {
		t.Fatalf("Failed to update tenant: %v", err)
	}

	// Verify update
	updated, err := ts.GetTenant(context.Background(), tenant.ID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if updated.Plan != "pro" {
		t.Errorf("Expected plan 'pro', got %s", updated.Plan)
	}
}

func TestTenantService_DeleteTenant(t *testing.T) {
	ts := service.NewTenantService()

	tenant, err := ts.CreateTenant(context.Background(), &service.TenantCreateRequest{
		Name: "Delete Test",
		Slug: "delete-test",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Delete tenant
	err = ts.DeleteTenant(context.Background(), tenant.ID)
	if err != nil {
		t.Fatalf("Failed to delete tenant: %v", err)
	}

	// Verify soft delete
	deleted, err := ts.GetTenant(context.Background(), tenant.ID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if deleted.Status != "deleted" {
		t.Errorf("Expected status 'deleted', got %s", deleted.Status)
	}
}

func TestTenantService_ListTenants(t *testing.T) {
	ts := service.NewTenantService()

	// Create multiple tenants
	for i := 0; i < 3; i++ {
		_, err := ts.CreateTenant(context.Background(), &service.TenantCreateRequest{
			Name: "List Test",
			Plan: "free",
		})
		if err != nil {
			t.Fatalf("Failed to create tenant: %v", err)
		}
	}

	// List tenants
	tenants := ts.ListTenants(context.Background())

	// Should have at least 4 (3 created + 1 default)
	if len(tenants) < 4 {
		t.Errorf("Expected at least 4 tenants, got %d", len(tenants))
	}
}

func TestGitOpsManager_AddRepo(t *testing.T) {
	gm := service.NewGitOpsManager()

	tests := []struct {
		name    string
		repo    *service.GitOpsRepo
		wantErr bool
	}{
		{
			name: "valid repo",
			repo: &service.GitOpsRepo{
				Name:   "Test Repo",
				URL:    "https://github.com/test/pipelines",
				Branch: "main",
				Path:   "pipelines",
			},
			wantErr: false,
		},
		{
			name: "repo with defaults",
			repo: &service.GitOpsRepo{
				Name: "Default Branch",
				URL:  "https://github.com/test/default",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gm.AddRepo(context.Background(), tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitOpsManager_ValidatePipeline(t *testing.T) {
	gm := service.NewGitOpsManager()

	tests := []struct {
		name     string
		pipeline *service.PipelineDefinition
		wantErr  bool
	}{
		{
			name: "valid pipeline",
			pipeline: &service.PipelineDefinition{
				Name: "test-pipeline",
				Mode: "serial",
				Steps: []service.StepDefinition{
					{ID: "step1", CLI: "claude", Action: "analyze"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			pipeline: &service.PipelineDefinition{
				Mode: "serial",
				Steps: []service.StepDefinition{
					{ID: "step1", CLI: "claude"},
				},
			},
			wantErr: true,
		},
		{
			name: "no steps",
			pipeline: &service.PipelineDefinition{
				Name:  "no-steps",
				Mode:  "serial",
				Steps: []service.StepDefinition{},
			},
			wantErr: true,
		},
		{
			name: "step missing CLI",
			pipeline: &service.PipelineDefinition{
				Name: "missing-cli",
				Mode: "serial",
				Steps: []service.StepDefinition{
					{ID: "step1"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			pipeline: &service.PipelineDefinition{
				Name: "invalid-mode",
				Mode: "invalid",
				Steps: []service.StepDefinition{
					{ID: "step1", CLI: "claude"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate step IDs",
			pipeline: &service.PipelineDefinition{
				Name: "duplicate-ids",
				Mode: "serial",
				Steps: []service.StepDefinition{
					{ID: "step1", CLI: "claude"},
					{ID: "step1", CLI: "npm"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid dependency",
			pipeline: &service.PipelineDefinition{
				Name: "invalid-dep",
				Mode: "hybrid",
				Steps: []service.StepDefinition{
					{ID: "step1", CLI: "claude", DependsOn: []string{"nonexistent"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gm.ValidatePipeline(tt.pipeline)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePipeline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitOpsManager_SyncRepo(t *testing.T) {
	gm := service.NewGitOpsManager()

	repo := &service.GitOpsRepo{
		Name:   "Sync Test",
		URL:    "https://github.com/test/sync",
		Branch: "main",
	}

	if err := gm.AddRepo(context.Background(), repo); err != nil {
		t.Fatalf("Failed to add repo: %v", err)
	}

	// Sync the repo
	err := gm.SyncRepo(context.Background(), repo.ID)
	if err != nil {
		t.Errorf("SyncRepo() error = %v", err)
	}

	// Check status
	updated, err := gm.GetRepo(context.Background(), repo.ID)
	if err != nil {
		t.Fatalf("Failed to get repo: %v", err)
	}

	if updated.Status != "synced" {
		t.Errorf("Expected status 'synced', got %s", updated.Status)
	}
}

func TestGitOpsManager_ListRepos(t *testing.T) {
	gm := service.NewGitOpsManager()

	// Add multiple repos
	for i := 0; i < 3; i++ {
		err := gm.AddRepo(context.Background(), &service.GitOpsRepo{
			Name: "List Test",
			URL:  "https://github.com/test/list" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Failed to add repo: %v", err)
		}
	}

	repos := gm.ListRepos(context.Background())
	if len(repos) < 3 {
		t.Errorf("Expected at least 3 repos, got %d", len(repos))
	}
}