package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Tenant Service Tests

func TestTenantService_New(t *testing.T) {
	ts := service.NewTenantService()
	if ts == nil {
		t.Fatal("Expected non-nil tenant service")
	}
}

func TestTenantService_Create(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "tenant-1",
		Name:     "Test Tenant",
		Plan:     "pro",
		Region:   "us-east-1",
	}

	err := ts.Create(tenant)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
}

func TestTenantService_Create_MissingName(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:   "no-name",
		Plan: "free",
	}

	err := ts.Create(tenant)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestTenantService_Get(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "get-tenant",
		Name:     "Get Tenant",
		Plan:     "pro",
	}
	ts.Create(tenant)

	retrieved, err := ts.Get("get-tenant")
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if retrieved.Name != "Get Tenant" {
		t.Error("Tenant name mismatch")
	}
}

func TestTenantService_Get_NotFound(t *testing.T) {
	ts := service.NewTenantService()

	_, err := ts.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tenant")
	}
}

func TestTenantService_List(t *testing.T) {
	ts := service.NewTenantService()

	ts.Create(&service.Tenant{
		ID:   "list-1",
		Name: "List 1",
		Plan: "free",
	})

	ts.Create(&service.Tenant{
		ID:   "list-2",
		Name: "List 2",
		Plan: "pro",
	})

	tenants := ts.List()
	if len(tenants) < 2 {
		t.Errorf("Expected at least 2 tenants, got %d", len(tenants))
	}
}

func TestTenantService_ListByPlan(t *testing.T) {
	ts := service.NewTenantService()

	ts.Create(&service.Tenant{
		ID:   "plan-free-1",
		Name: "Free Tenant",
		Plan: "free",
	})

	ts.Create(&service.Tenant{
		ID:   "plan-pro-1",
		Name: "Pro Tenant",
		Plan: "pro",
	})

	freeTenants := ts.ListByPlan("free")
	for _, t := range freeTenants {
		if t.Plan != "free" {
			t.Error("Expected only free tenants")
		}
	}
}

func TestTenantService_Update(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "update-tenant",
		Name:     "Update Tenant",
		Plan:     "free",
	}
	ts.Create(tenant)

	err := ts.Update("update-tenant", map[string]interface{}{
		"plan":  "pro",
		"region": "eu-west-1",
	})

	if err != nil {
		t.Fatalf("Failed to update tenant: %v", err)
	}

	retrieved, _ := ts.Get("update-tenant")
	if retrieved.Plan != "pro" {
		t.Error("Plan not updated")
	}
}

func TestTenantService_Delete(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "delete-tenant",
		Name:     "Delete Tenant",
	}
	ts.Create(tenant)

	err := ts.Delete("delete-tenant")
	if err != nil {
		t.Fatalf("Failed to delete tenant: %v", err)
	}

	_, err = ts.Get("delete-tenant")
	if err == nil {
		t.Error("Expected error for deleted tenant")
	}
}

func TestTenantService_Suspend(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "suspend-tenant",
		Name:     "Suspend Tenant",
		Status:   "active",
	}
	ts.Create(tenant)

	err := ts.Suspend("suspend-tenant", "Payment overdue")
	if err != nil {
		t.Fatalf("Failed to suspend tenant: %v", err)
	}

	retrieved, _ := ts.Get("suspend-tenant")
	if retrieved.Status != "suspended" {
		t.Error("Expected status suspended")
	}
}

func TestTenantService_Activate(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "activate-tenant",
		Name:     "Activate Tenant",
		Status:   "suspended",
	}
	ts.Create(tenant)

	err := ts.Activate("activate-tenant")
	if err != nil {
		t.Fatalf("Failed to activate tenant: %v", err)
	}

	retrieved, _ := ts.Get("activate-tenant")
	if retrieved.Status != "active" {
		t.Error("Expected status active")
	}
}

func TestTenantService_SetPlan(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "plan-tenant",
		Name:     "Plan Tenant",
		Plan:     "free",
	}
	ts.Create(tenant)

	err := ts.SetPlan("plan-tenant", "enterprise")
	if err != nil {
		t.Fatalf("Failed to set plan: %v", err)
	}

	retrieved, _ := ts.Get("plan-tenant")
	if retrieved.Plan != "enterprise" {
		t.Error("Plan not set")
	}
}

func TestTenantService_SetSettings(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "settings-tenant",
		Name:     "Settings Tenant",
	}
	ts.Create(tenant)

	settings := map[string]interface{}{
		"max_users":     100,
		"features":      []string{"api", "webhooks"},
		"notifications": true,
	}

	err := ts.SetSettings("settings-tenant", settings)
	if err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}
}

func TestTenantService_GetSettings(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "getsettings-tenant",
		Name:     "Get Settings Tenant",
		Settings: map[string]interface{}{
			"custom_domain": "example.com",
		},
	}
	ts.Create(tenant)

	settings := ts.GetSettings("getsettings-tenant")
	if settings == nil {
		t.Fatal("Expected settings")
	}

	if settings["custom_domain"] != "example.com" {
		t.Error("Settings value mismatch")
	}
}

func TestTenantService_SetQuota(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "quota-tenant",
		Name:     "Quota Tenant",
	}
	ts.Create(tenant)

	quota := map[string]int{
		"api_calls":    10000,
		"storage_mb":   1024,
		"users":        50,
	}

	err := ts.SetQuota("quota-tenant", quota)
	if err != nil {
		t.Fatalf("Failed to set quota: %v", err)
	}
}

func TestTenantService_GetQuota(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "getquota-tenant",
		Name:     "Get Quota Tenant",
		Quota: map[string]int{
			"api_calls": 5000,
		},
	}
	ts.Create(tenant)

	quota := ts.GetQuota("getquota-tenant")
	if quota == nil {
		t.Fatal("Expected quota")
	}

	if quota["api_calls"] != 5000 {
		t.Error("Quota value mismatch")
	}
}

func TestTenantBasic_CheckQuota(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "checkquota-tenant",
		Name:     "Check Quota Tenant",
		Quota: map[string]int{
			"api_calls": 100,
		},
		Usage: map[string]int{
			"api_calls": 75,
		},
	}
	ts.Create(tenant)

	// Should be within quota
	if !ts.CheckQuota("checkquota-tenant", "api_calls", 20) {
		t.Error("Expected 20 to be within quota")
	}

	// Should exceed quota
	if ts.CheckQuota("checkquota-tenant", "api_calls", 30) {
		t.Error("Expected 30 to exceed quota")
	}
}

func TestTenantService_UpdateUsage(t *testing.T) {
	ts := service.NewTenantService()

	tenant := &service.Tenant{
		ID:       "usage-tenant",
		Name:     "Usage Tenant",
		Usage: map[string]int{
			"api_calls": 10,
		},
	}
	ts.Create(tenant)

	err := ts.UpdateUsage("usage-tenant", "api_calls", 100)
	if err != nil {
		t.Fatalf("Failed to update usage: %v", err)
	}

	retrieved, _ := ts.Get("usage-tenant")
	if retrieved.Usage["api_calls"] != 110 {
		t.Error("Usage not updated correctly")
	}
}

func TestTenantService_GetStats(t *testing.T) {
	ts := service.NewTenantService()

	ts.Create(&service.Tenant{
		ID:   "stats-1",
		Name: "Stats 1",
		Plan: "free",
	})

	ts.Create(&service.Tenant{
		ID:   "stats-2",
		Name: "Stats 2",
		Plan: "pro",
	})

	stats := ts.GetStats()

	if stats.TotalTenants < 2 {
		t.Errorf("Expected at least 2 tenants, got %d", stats.TotalTenants)
	}
}

func TestTenantService_TenantToJSON(t *testing.T) {
	tenant := &service.Tenant{
		ID:        "json-tenant",
		Name:      "JSON Tenant",
		Plan:      "pro",
		Region:    "us-east-1",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	data, err := tenant.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}