package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

func TestQuotaService_New(t *testing.T) {
	qs := service.NewQuotaService()
	if qs == nil {
		t.Fatal("Expected non-nil quota service")
	}
}

func TestQuotaService_SetQuota(t *testing.T) {
	qs := service.NewQuotaService()

	err := qs.SetQuota(context.Background(), &service.Quota{
		Name:     "test-quota",
		TenantID: "tenant-1",
		Resources: map[string]int64{
			"tasks":       100,
			"concurrent":  5,
			"storage_mb":  1024,
		},
		Period:   service.QuotaPeriodDaily,
		Enforced: true,
	})

	if err != nil {
		t.Fatalf("Failed to set quota: %v", err)
	}
}

func TestQuotaService_SetQuota_MissingTenant(t *testing.T) {
	qs := service.NewQuotaService()

	err := qs.SetQuota(context.Background(), &service.Quota{
		Name: "test",
	})

	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestQuotaService_GetQuota(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		ID:       "quota-1",
		Name:     "Test",
		TenantID: "tenant-1",
		Period:   service.QuotaPeriodDaily,
	})

	quota, err := qs.GetQuota("quota-1")
	if err != nil {
		t.Fatalf("Failed to get quota: %v", err)
	}

	if quota.Name != "Test" {
		t.Error("Quota name mismatch")
	}
}

func TestQuotaService_GetQuota_NotFound(t *testing.T) {
	qs := service.NewQuotaService()

	_, err := qs.GetQuota("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent quota")
	}
}

func TestQuotaService_GetQuotaByTenant(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		Name:     "Tenant Quota",
		TenantID: "tenant-get",
		Period:   service.QuotaPeriodDaily,
	})

	quota, err := qs.GetQuotaByTenant("tenant-get")
	if err != nil {
		t.Fatalf("Failed to get quota: %v", err)
	}

	if quota.TenantID != "tenant-get" {
		t.Error("Tenant ID mismatch")
	}
}

func TestQuotaService_CheckQuota(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-check",
		Resources: map[string]int64{
			"tasks": 10,
		},
		Period:   service.QuotaPeriodDaily,
		Enforced: false,
	})

	// Should be allowed
	result, err := qs.CheckQuota(context.Background(), "tenant-check", "tasks", 5)
	if err != nil {
		t.Fatalf("Check quota failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected quota to be allowed")
	}
}

func TestQuotaService_CheckQuota_Exceeded(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-exceeded",
		Resources: map[string]int64{
			"tasks": 10,
		},
		Period:   service.QuotaPeriodDaily,
		Enforced: true,
	})

	// Consume 10 tasks
	qs.ConsumeQuota(context.Background(), "tenant-exceeded", "tasks", 10)

	// Should not be allowed
	result, err := qs.CheckQuota(context.Background(), "tenant-exceeded", "tasks", 1)
	if err == nil {
		t.Error("Expected error for quota exceeded")
	}

	if result.Allowed {
		t.Error("Expected quota to not be allowed")
	}
}

func TestQuotaService_ConsumeQuota(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-consume",
		Resources: map[string]int64{
			"tasks": 100,
		},
		Period:   service.QuotaPeriodDaily,
	})

	err := qs.ConsumeQuota(context.Background(), "tenant-consume", "tasks", 10)
	if err != nil {
		t.Fatalf("Failed to consume quota: %v", err)
	}

	usage, _ := qs.GetUsage("tenant-consume")
	if usage.Resources["tasks"] != 10 {
		t.Errorf("Expected usage 10, got %d", usage.Resources["tasks"])
	}
}

func TestQuotaService_ReleaseQuota(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-release",
		Resources: map[string]int64{
			"tasks": 100,
		},
		Period: service.QuotaPeriodDaily,
	})

	qs.ConsumeQuota(context.Background(), "tenant-release", "tasks", 20)
	qs.ReleaseQuota(context.Background(), "tenant-release", "tasks", 5)

	usage, _ := qs.GetUsage("tenant-release")
	if usage.Resources["tasks"] != 15 {
		t.Errorf("Expected usage 15, got %d", usage.Resources["tasks"])
	}
}

func TestQuotaService_GetUsage(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-usage",
		Period:   service.QuotaPeriodDaily,
	})

	usage, err := qs.GetUsage("tenant-usage")
	if err != nil {
		t.Fatalf("Failed to get usage: %v", err)
	}

	if usage.TenantID != "tenant-usage" {
		t.Error("Usage tenant ID mismatch")
	}
}

func TestQuotaService_ResetUsage(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-reset",
		Resources: map[string]int64{
			"tasks": 100,
		},
		Period: service.QuotaPeriodDaily,
	})

	qs.ConsumeQuota(context.Background(), "tenant-reset", "tasks", 50)
	qs.ResetUsage("tenant-reset")

	usage, _ := qs.GetUsage("tenant-reset")
	if usage.Resources["tasks"] != 0 {
		t.Error("Expected usage to be reset to 0")
	}
}

func TestQuotaService_ListQuotas(t *testing.T) {
	qs := service.NewQuotaService()

	for i := 0; i < 3; i++ {
		qs.SetQuota(context.Background(), &service.Quota{
			Name:     string(rune('a' + i)),
			TenantID: string(rune('a' + i)),
			Period:   service.QuotaPeriodDaily,
		})
	}

	quotas := qs.ListQuotas()
	if len(quotas) < 3 {
		t.Errorf("Expected at least 3 quotas, got %d", len(quotas))
	}
}

func TestQuotaService_DeleteQuota(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		ID:       "delete-test",
		TenantID: "tenant-delete",
		Period:   service.QuotaPeriodDaily,
	})

	err := qs.DeleteQuota("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete quota: %v", err)
	}

	_, err = qs.GetQuota("delete-test")
	if err == nil {
		t.Error("Expected error for deleted quota")
	}
}

func TestQuotaService_GetQuotaStatus(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-status",
		Resources: map[string]int64{
			"tasks": 100,
		},
		Period: service.QuotaPeriodDaily,
	})

	qs.ConsumeQuota(context.Background(), "tenant-status", "tasks", 25)

	status, err := qs.GetQuotaStatus("tenant-status")
	if err != nil {
		t.Fatalf("Failed to get quota status: %v", err)
	}

	if status["tenant_id"] != "tenant-status" {
		t.Error("Status tenant ID mismatch")
	}
}

func TestQuotaService_GetStats(t *testing.T) {
	qs := service.NewQuotaService()

	qs.SetQuota(context.Background(), &service.Quota{
		TenantID: "tenant-stats",
		Period:   service.QuotaPeriodDaily,
	})

	stats := qs.GetStats()

	if stats["total_quotas"].(int) < 1 {
		t.Error("Expected at least 1 quota")
	}
}

func TestQuotaService_Periods(t *testing.T) {
	qs := service.NewQuotaService()

	periods := []service.QuotaPeriod{
		service.QuotaPeriodDaily,
		service.QuotaPeriodWeekly,
		service.QuotaPeriodMonthly,
		service.QuotaPeriodYearly,
	}

	for _, period := range periods {
		err := qs.SetQuota(context.Background(), &service.Quota{
			TenantID: "tenant-" + string(period),
			Period:   period,
		})

		if err != nil {
			t.Errorf("Failed to set quota for period %s: %v", period, err)
		}
	}
}

// Health Service Tests

func TestQuotaHealthService_New(t *testing.T) {
	hs := service.NewHealthService("1.0.0")
	if hs == nil {
		t.Fatal("Expected non-nil health service")
	}
}

func TestQuotaHealthService_RegisterCheck(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("database", func() error {
		return nil
	}, 30*time.Second)
}

func TestHealthService_RunChecks(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("check1", func() error {
		return nil
	}, 30*time.Second)

	hs.RunChecks(context.Background())

	status := hs.GetStatus()
	if status.Checks["check1"].Status != service.HealthStatusHealthy {
		t.Error("Expected check1 to be healthy")
	}
}

func TestHealthService_RunChecks_Failing(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("failing", func() error {
		return errors.New("check failed")
	}, 30*time.Second)

	hs.RunChecks(context.Background())

	status := hs.GetStatus()
	if status.Checks["failing"].Status != service.HealthStatusUnhealthy {
		t.Error("Expected failing check to be unhealthy")
	}
}

func TestQuotaHealthService_GetStatus(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	status := hs.GetStatus()

	if status.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
	if status.Status == "" {
		t.Error("Expected non-empty status")
	}
	if status.Uptime < 0 {
		t.Error("Expected non-negative uptime")
	}
}

func TestHealthService_IsHealthy(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("healthy", func() error {
		return nil
	}, 30*time.Second)

	hs.RunChecks(context.Background())

	if !hs.IsHealthy() {
		t.Error("Expected service to be healthy")
	}
}

func TestHealthService_AddDependency(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.AddDependency("redis", service.DependencyHealth{
		Name:   "redis",
		Status: service.HealthStatusHealthy,
	})

	status := hs.GetStatus()
	if status.Dependencies["redis"].Status != service.HealthStatusHealthy {
		t.Error("Expected redis dependency to be healthy")
	}
}

func TestHealthService_SetDependencyHealth(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.SetDependencyHealth("postgres", service.HealthStatusHealthy, 5, nil)

	status := hs.GetStatus()
	if status.Dependencies["postgres"].Status != service.HealthStatusHealthy {
		t.Error("Expected postgres dependency to be healthy")
	}
}

func TestHealthService_GetLive(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	live := hs.GetLive()

	if live["status"] != "up" {
		t.Error("Expected status to be up")
	}
}

func TestHealthService_GetReady(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("ready", func() error {
		return nil
	}, 30*time.Second)

	hs.RunChecks(context.Background())

	ready, err := hs.GetReady()
	if err != nil {
		t.Fatalf("Service not ready: %v", err)
	}

	if ready["ready"] != true {
		t.Error("Expected service to be ready")
	}
}

func TestHealthService_GetReady_NotReady(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	hs.RegisterCheck("notready", func() error {
		return errors.New("not ready")
	}, 30*time.Second)

	hs.RunChecks(context.Background())

	_, err := hs.GetReady()
	if err == nil {
		t.Error("Expected error for not ready service")
	}
}

func TestHealthService_SystemInfo(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	status := hs.GetStatus()

	if status.System.GoVersion == "" {
		t.Error("Expected go version")
	}
	if status.System.NumCPU == 0 {
		t.Error("Expected num cpu")
	}
}

func TestHealthService_ToJSON(t *testing.T) {
	hs := service.NewHealthService("1.0.0")

	status := hs.GetStatus()
	data, err := status.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestLivenessProbe_Check(t *testing.T) {
	probe := service.NewLivenessProbe("test", func() bool {
		return true
	}, 3)

	if !probe.Check() {
		t.Error("Expected check to return true")
	}

	if !probe.IsAlive() {
		t.Error("Expected probe to be alive")
	}
}

func TestLivenessProbe_Failing(t *testing.T) {
	failures := 0
	probe := service.NewLivenessProbe("failing", func() bool {
		failures++
		return failures > 4
	}, 3)

	for i := 0; i < 5; i++ {
		probe.Check()
	}

	if probe.IsAlive() {
		t.Error("Expected probe to not be alive after failures")
	}
}

func TestReadinessProbe(t *testing.T) {
	probe := service.NewReadinessProbe("test", func() bool {
		return true
	}, true)

	if probe == nil {
		t.Fatal("Expected non-nil probe")
	}
}