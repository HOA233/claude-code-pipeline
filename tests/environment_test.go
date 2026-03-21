package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Environment Service Tests

func TestEnvironmentService_New(t *testing.T) {
	es := service.NewEnvironmentService()
	if es == nil {
		t.Fatal("Expected non-nil environment service")
	}
}

func TestEnvironmentService_CreateEnvironment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "production",
		TenantID: "tenant-1",
		Type:     service.EnvTypeProduction,
	}

	err := es.CreateEnvironment(context.Background(), env)
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	if env.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if env.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", env.Status)
	}
}

func TestEnvironmentService_CreateEnvironment_MissingName(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		TenantID: "tenant-1",
		Type:     service.EnvTypeDevelopment,
	}

	err := es.CreateEnvironment(context.Background(), env)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestEnvironmentService_CreateEnvironment_MissingTenant(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name: "dev",
		Type: service.EnvTypeDevelopment,
	}

	err := es.CreateEnvironment(context.Background(), env)
	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestEnvironmentService_GetEnvironment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "staging",
		TenantID: "tenant-get",
		Type:     service.EnvTypeStaging,
	}
	es.CreateEnvironment(context.Background(), env)

	retrieved, err := es.GetEnvironment(env.ID)
	if err != nil {
		t.Fatalf("Failed to get environment: %v", err)
	}

	if retrieved.Name != "staging" {
		t.Error("Environment name mismatch")
	}
}

func TestEnvironmentService_GetEnvironment_NotFound(t *testing.T) {
	es := service.NewEnvironmentService()

	_, err := es.GetEnvironment("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent environment")
	}
}

func TestEnvironmentService_GetEnvironmentByName(t *testing.T) {
	es := service.NewEnvironmentService()

	es.CreateEnvironment(context.Background(), &service.Environment{
		Name:     "dev",
		TenantID: "tenant-name",
		Type:     service.EnvTypeDevelopment,
	})

	env, err := es.GetEnvironmentByName("tenant-name", "dev")
	if err != nil {
		t.Fatalf("Failed to get environment by name: %v", err)
	}

	if env.Name != "dev" {
		t.Error("Environment name mismatch")
	}
}

func TestEnvironmentService_ListEnvironments(t *testing.T) {
	es := service.NewEnvironmentService()

	es.CreateEnvironment(context.Background(), &service.Environment{
		Name:     "env1",
		TenantID: "tenant-list",
		Type:     service.EnvTypeDevelopment,
	})

	es.CreateEnvironment(context.Background(), &service.Environment{
		Name:     "env2",
		TenantID: "tenant-list",
		Type:     service.EnvTypeStaging,
	})

	es.CreateEnvironment(context.Background(), &service.Environment{
		Name:     "env3",
		TenantID: "other-tenant",
		Type:     service.EnvTypeProduction,
	})

	envs := es.ListEnvironments("tenant-list")
	if len(envs) < 2 {
		t.Errorf("Expected at least 2 environments, got %d", len(envs))
	}
}

func TestEnvironmentService_UpdateEnvironment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "original",
		TenantID: "tenant-update",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	updated, err := es.UpdateEnvironment(context.Background(), env.ID, map[string]interface{}{
		"name":        "updated",
		"description": "Updated description",
	})

	if err != nil {
		t.Fatalf("Failed to update environment: %v", err)
	}

	if updated.Name != "updated" {
		t.Error("Name not updated")
	}
}

func TestEnvironmentService_DeleteEnvironment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "to-delete",
		TenantID: "tenant-delete",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	err := es.DeleteEnvironment(env.ID)
	if err != nil {
		t.Fatalf("Failed to delete environment: %v", err)
	}

	_, err = es.GetEnvironment(env.ID)
	if err == nil {
		t.Error("Expected error for deleted environment")
	}
}

func TestEnvironmentService_SetVariable(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "vars",
		TenantID: "tenant-vars",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	err := es.SetVariable(env.ID, "API_KEY", "secret-key", true)
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}
}

func TestEnvironmentService_GetVariable(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "get-var",
		TenantID: "tenant-getvar",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	es.SetVariable(env.ID, "DB_URL", "postgres://localhost", false)

	v, err := es.GetVariable(env.ID, "DB_URL")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if v.Value != "postgres://localhost" {
		t.Error("Variable value mismatch")
	}
}

func TestEnvironmentService_ListVariables(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "list-vars",
		TenantID: "tenant-listvars",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	es.SetVariable(env.ID, "VAR1", "value1", false)
	es.SetVariable(env.ID, "VAR2", "value2", false)
	es.SetVariable(env.ID, "SECRET", "secret-value", true)

	vars, err := es.ListVariables(env.ID, false)
	if err != nil {
		t.Fatalf("Failed to list variables: %v", err)
	}

	// Should not include secrets
	for _, v := range vars {
		if v.Secret {
			t.Error("Should not include secret variables")
		}
	}
}

func TestEnvironmentService_DeleteVariable(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "del-var",
		TenantID: "tenant-delvar",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	es.SetVariable(env.ID, "TO_DELETE", "value", false)

	err := es.DeleteVariable(env.ID, "TO_DELETE")
	if err != nil {
		t.Fatalf("Failed to delete variable: %v", err)
	}

	_, err = es.GetVariable(env.ID, "TO_DELETE")
	if err == nil {
		t.Error("Expected error for deleted variable")
	}
}

func TestEnvironmentService_CreateDeployment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "deploy",
		TenantID: "tenant-deploy",
		Type:     service.EnvTypeStaging,
	}
	es.CreateEnvironment(context.Background(), env)

	deployment := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v1.0.0",
		TriggeredBy:   "user-1",
	}

	err := es.CreateDeployment(context.Background(), deployment)
	if err != nil {
		t.Fatalf("Failed to create deployment: %v", err)
	}

	if deployment.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if deployment.Status != service.DeploymentStatusPending {
		t.Errorf("Expected status '%s', got '%s'", service.DeploymentStatusPending, deployment.Status)
	}
}

func TestEnvironmentService_StartDeployment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "start-deploy",
		TenantID: "tenant-start",
		Type:     service.EnvTypeStaging,
	}
	es.CreateEnvironment(context.Background(), env)

	deployment := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v1.0.0",
	}
	es.CreateDeployment(context.Background(), deployment)

	err := es.StartDeployment(deployment.ID)
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}

	d, _ := es.GetDeployment(deployment.ID)
	if d.Status != service.DeploymentStatusDeploying {
		t.Error("Expected status to be deploying")
	}
}

func TestEnvironmentService_CompleteDeployment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "complete-deploy",
		TenantID: "tenant-complete",
		Type:     service.EnvTypeProduction,
	}
	es.CreateEnvironment(context.Background(), env)

	deployment := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v1.0.0",
	}
	es.CreateDeployment(context.Background(), deployment)
	es.StartDeployment(deployment.ID)

	err := es.CompleteDeployment(deployment.ID, true)
	if err != nil {
		t.Fatalf("Failed to complete deployment: %v", err)
	}

	d, _ := es.GetDeployment(deployment.ID)
	if d.Status != service.DeploymentStatusSucceeded {
		t.Error("Expected status to be succeeded")
	}

	if d.CompletedAt == nil {
		t.Error("Expected completed_at to be set")
	}
}

func TestEnvironmentService_RollbackDeployment(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "rollback-deploy",
		TenantID: "tenant-rollback",
		Type:     service.EnvTypeProduction,
	}
	es.CreateEnvironment(context.Background(), env)

	// Create first deployment
	dep1 := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v1.0.0",
	}
	es.CreateDeployment(context.Background(), dep1)
	es.StartDeployment(dep1.ID)
	es.CompleteDeployment(dep1.ID, true)

	// Create second deployment
	dep2 := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v2.0.0",
	}
	es.CreateDeployment(context.Background(), dep2)
	es.StartDeployment(dep2.ID)
	es.CompleteDeployment(dep2.ID, true)

	// Rollback to first deployment
	err := es.RollbackDeployment(dep2.ID, dep1.ID)
	if err != nil {
		t.Fatalf("Failed to rollback deployment: %v", err)
	}

	d, _ := es.GetDeployment(dep2.ID)
	if d.Status != service.DeploymentStatusRolledBack {
		t.Error("Expected status to be rolled_back")
	}
}

func TestEnvironmentService_GetDeployments(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "list-deploys",
		TenantID: "tenant-listdep",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	for i := 0; i < 5; i++ {
		es.CreateDeployment(context.Background(), &service.Deployment{
			EnvironmentID: env.ID,
			Version:       "v0.0." + string(rune('0'+i)),
		})
	}

	deployments := es.GetDeployments(env.ID, 3)
	if len(deployments) > 3 {
		t.Errorf("Expected at most 3 deployments, got %d", len(deployments))
	}
}

func TestEnvironmentService_AddDeploymentLog(t *testing.T) {
	es := service.NewEnvironmentService()

	env := &service.Environment{
		Name:     "log-deploy",
		TenantID: "tenant-log",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), env)

	deployment := &service.Deployment{
		EnvironmentID: env.ID,
		Version:       "v1.0.0",
	}
	es.CreateDeployment(context.Background(), deployment)

	err := es.AddDeploymentLog(deployment.ID, "Starting deployment...")
	if err != nil {
		t.Fatalf("Failed to add deployment log: %v", err)
	}

	d, _ := es.GetDeployment(deployment.ID)
	if len(d.Logs) == 0 {
		t.Error("Expected logs to be added")
	}
}

func TestEnvironmentService_PromoteEnvironment(t *testing.T) {
	es := service.NewEnvironmentService()

	devEnv := &service.Environment{
		Name:     "dev",
		TenantID: "tenant-promote",
		Type:     service.EnvTypeDevelopment,
	}
	es.CreateEnvironment(context.Background(), devEnv)

	prodEnv := &service.Environment{
		Name:     "prod",
		TenantID: "tenant-promote",
		Type:     service.EnvTypeProduction,
	}
	es.CreateEnvironment(context.Background(), prodEnv)

	deployment, err := es.PromoteEnvironment(context.Background(), devEnv.ID, prodEnv.ID)
	if err != nil {
		t.Fatalf("Failed to promote environment: %v", err)
	}

	if deployment.EnvironmentID != prodEnv.ID {
		t.Error("Deployment should be for target environment")
	}
}

func TestEnvironmentService_EnvironmentTypes(t *testing.T) {
	types := []service.EnvironmentType{
		service.EnvTypeDevelopment,
		service.EnvTypeStaging,
		service.EnvTypeProduction,
		service.EnvTypePreview,
	}

	es := service.NewEnvironmentService()

	for i, envType := range types {
		env := &service.Environment{
			Name:     "env-" + string(rune('a'+i)),
			TenantID: "tenant-types",
			Type:     envType,
		}

		err := es.CreateEnvironment(context.Background(), env)
		if err != nil {
			t.Errorf("Failed to create %s environment: %v", envType, err)
		}
	}
}

func TestEnvironmentService_EnvironmentToJSON(t *testing.T) {
	env := &service.Environment{
		ID:       "env-1",
		Name:     "production",
		TenantID: "tenant-1",
		Type:     service.EnvTypeProduction,
		Status:   "active",
	}

	data, err := env.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestEnvironmentService_DeploymentToJSON(t *testing.T) {
	now := time.Now()
	deployment := &service.Deployment{
		ID:            "dep-1",
		EnvironmentID: "env-1",
		Version:       "v1.0.0",
		Status:        service.DeploymentStatusSucceeded,
		StartedAt:     now,
	}

	data, err := deployment.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}