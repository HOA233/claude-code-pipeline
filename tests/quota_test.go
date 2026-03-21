package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Quota Service Tests

func TestQuotaService_New(t *testing.T) {
	qs := service.NewQuotaService()
	if qs == nil {
		t.Fatal("Expected non-nil quota service")
	}
}

func TestQuotaService_Create(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "quota-1",
		Name:     "API Calls",
		TenantID: "tenant-1",
		Resource: "api.calls",
		Limit:    10000,
		Period:   "monthly",
	}

	err := qs.Create(quota)
	if err != nil {
		t.Fatalf("Failed to create quota: %v", err)
	}
}

func TestQuotaService_Create_MissingName(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "no-name",
		TenantID: "tenant-1",
		Resource: "api.calls",
	}

	err := qs.Create(quota)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestQuotaService_Get(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "get-quota",
		Name:     "Get Quota",
		TenantID: "tenant-get",
		Resource: "test.resource",
	}
	qs.Create(quota)

	retrieved, err := qs.Get("get-quota")
	if err != nil {
		t.Fatalf("Failed to get quota: %v", err)
	}

	if retrieved.Name != "Get Quota" {
		t.Error("Quota name mismatch")
	}
}

func TestQuotaService_Get_NotFound(t *testing.T) {
	qs := service.NewQuotaService()

	_, err := qs.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent quota")
	}
}

func TestQuotaService_List(t *testing.T) {
	qs := service.NewQuotaService()

	qs.Create(&service.Quota{
		ID:       "list-1",
		Name:     "List 1",
		TenantID: "tenant-list",
		Resource: "res1",
	})

	qs.Create(&service.Quota{
		ID:       "list-2",
		Name:     "List 2",
		TenantID: "tenant-list",
		Resource: "res2",
	})

	quotas := qs.List("tenant-list")
	if len(quotas) < 2 {
		t.Errorf("Expected at least 2 quotas, got %d", len(quotas))
	}
}

func TestQuotaService_Update(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "update-quota",
		Name:     "Update Quota",
		TenantID: "tenant-update",
		Resource: "update.resource",
		Limit:    1000,
	}
	qs.Create(quota)

	err := qs.Update("update-quota", map[string]interface{}{
		"limit": 5000,
	})

	if err != nil {
		t.Fatalf("Failed to update quota: %v", err)
	}

	retrieved, _ := qs.Get("update-quota")
	if retrieved.Limit != 5000 {
		t.Error("Limit not updated")
	}
}

func TestQuotaService_Delete(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "delete-quota",
		Name:     "Delete Quota",
		TenantID: "tenant-delete",
		Resource: "delete.resource",
	}
	qs.Create(quota)

	err := qs.Delete("delete-quota")
	if err != nil {
		t.Fatalf("Failed to delete quota: %v", err)
	}

	_, err = qs.Get("delete-quota")
	if err == nil {
		t.Error("Expected error for deleted quota")
	}
}

func TestQuotaService_Consume(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "consume-quota",
		Name:     "Consume Quota",
		TenantID: "tenant-consume",
		Resource: "consume.resource",
		Limit:    100,
	}
	qs.Create(quota)

	err := qs.Consume("consume-quota", 10)
	if err != nil {
		t.Fatalf("Failed to consume quota: %v", err)
	}

	retrieved, _ := qs.Get("consume-quota")
	if retrieved.Used != 10 {
		t.Errorf("Expected used 10, got %d", retrieved.Used)
	}
}

func TestQuotaService_Consume_ExceedsLimit(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "exceed-quota",
		Name:     "Exceed Quota",
		TenantID: "tenant-exceed",
		Resource: "exceed.resource",
		Limit:    50,
	}
	qs.Create(quota)

	// First consume should work
	qs.Consume("exceed-quota", 40)

	// This should fail as it would exceed the limit
	err := qs.Consume("exceed-quota", 20)
	if err == nil {
		t.Error("Expected error for exceeding quota limit")
	}
}

func TestQuotaService_Release(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "release-quota",
		Name:     "Release Quota",
		TenantID: "tenant-release",
		Resource: "release.resource",
		Limit:    100,
	}
	qs.Create(quota)
	qs.Consume("release-quota", 50)

	err := qs.Release("release-quota", 20)
	if err != nil {
		t.Fatalf("Failed to release quota: %v", err)
	}

	retrieved, _ := qs.Get("release-quota")
	if retrieved.Used != 30 {
		t.Errorf("Expected used 30, got %d", retrieved.Used)
	}
}

func TestQuotaService_Reset(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "reset-quota",
		Name:     "Reset Quota",
		TenantID: "tenant-reset",
		Resource: "reset.resource",
		Limit:    100,
	}
	qs.Create(quota)
	qs.Consume("reset-quota", 50)

	err := qs.Reset("reset-quota")
	if err != nil {
		t.Fatalf("Failed to reset quota: %v", err)
	}

	retrieved, _ := qs.Get("reset-quota")
	if retrieved.Used != 0 {
		t.Error("Expected used to be 0 after reset")
	}
}

func TestQuotaService_CheckLimit(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "check-quota",
		Name:     "Check Quota",
		TenantID: "tenant-check",
		Resource: "check.resource",
		Limit:    100,
	}
	qs.Create(quota)

	// Should be within limit
	if !qs.CheckLimit("check-quota", 50) {
		t.Error("Expected 50 to be within limit")
	}

	qs.Consume("check-quota", 90)

	// Should not be within limit
	if qs.CheckLimit("check-quota", 20) {
		t.Error("Expected 20 to exceed limit")
	}
}

func TestQuotaService_GetUsage(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "usage-quota",
		Name:     "Usage Quota",
		TenantID: "tenant-usage",
		Resource: "usage.resource",
		Limit:    100,
	}
	qs.Create(quota)
	qs.Consume("usage-quota", 75)

	usage := qs.GetUsage("usage-quota")
	if usage != 75 {
		t.Errorf("Expected usage 75, got %d", usage)
	}
}

func TestQuotaService_GetRemaining(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "remaining-quota",
		Name:     "Remaining Quota",
		TenantID: "tenant-remaining",
		Resource: "remaining.resource",
		Limit:    100,
	}
	qs.Create(quota)
	qs.Consume("remaining-quota", 30)

	remaining := qs.GetRemaining("remaining-quota")
	if remaining != 70 {
		t.Errorf("Expected remaining 70, got %d", remaining)
	}
}

func TestQuotaService_GetPercentage(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "percent-quota",
		Name:     "Percent Quota",
		TenantID: "tenant-percent",
		Resource: "percent.resource",
		Limit:    100,
	}
	qs.Create(quota)
	qs.Consume("percent-quota", 75)

	percent := qs.GetPercentage("percent-quota")
	if percent != 75.0 {
		t.Errorf("Expected 75%%, got %f%%", percent)
	}
}

func TestQuotaService_SetAlertThreshold(t *testing.T) {
	qs := service.NewQuotaService()

	quota := &service.Quota{
		ID:       "alert-quota",
		Name:     "Alert Quota",
		TenantID: "tenant-alert",
		Resource: "alert.resource",
		Limit:    100,
	}
	qs.Create(quota)

	err := qs.SetAlertThreshold("alert-quota", 80)
	if err != nil {
		t.Fatalf("Failed to set alert threshold: %v", err)
	}
}

func TestQuotaService_GetStats(t *testing.T) {
	qs := service.NewQuotaService()

	qs.Create(&service.Quota{
		ID:       "stats-quota",
		Name:     "Stats Quota",
		TenantID: "tenant-stats",
		Resource: "stats.resource",
		Limit:    100,
	})

	stats := qs.GetStats("tenant-stats")
	if stats.TotalQuotas < 1 {
		t.Errorf("Expected at least 1 quota in stats, got %d", stats.TotalQuotas)
	}
}

func TestQuotaService_QuotaToJSON(t *testing.T) {
	quota := &service.Quota{
		ID:        "json-quota",
		Name:      "JSON Quota",
		TenantID:  "tenant-1",
		Resource:  "json.resource",
		Limit:     1000,
		Used:      250,
		Period:    "monthly",
		CreatedAt: time.Now(),
	}

	data, err := quota.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}