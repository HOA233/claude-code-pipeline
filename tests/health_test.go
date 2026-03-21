package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Health Service Tests

func TestHealthService_New(t *testing.T) {
	hs := service.NewHealthService()
	if hs == nil {
		t.Fatal("Expected non-nil health service")
	}
}

func TestHealthService_RegisterCheck(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:       "check-1",
		Name:     "Database",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Enabled:  true,
	}

	err := hs.RegisterCheck(check)
	if err != nil {
		t.Fatalf("Failed to register check: %v", err)
	}
}

func TestHealthService_RegisterCheck_MissingName(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID: "no-name",
	}

	err := hs.RegisterCheck(check)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestHealthService_GetCheck(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:       "get-check",
		Name:     "Get Check",
		Interval: time.Minute,
	}
	hs.RegisterCheck(check)

	retrieved, err := hs.GetCheck("get-check")
	if err != nil {
		t.Fatalf("Failed to get check: %v", err)
	}

	if retrieved.Name != "Get Check" {
		t.Error("Check name mismatch")
	}
}

func TestHealthService_GetCheck_NotFound(t *testing.T) {
	hs := service.NewHealthService()

	_, err := hs.GetCheck("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent check")
	}
}

func TestHealthService_ListChecks(t *testing.T) {
	hs := service.NewHealthService()

	hs.RegisterCheck(&service.HealthCheck{
		ID:   "list-1",
		Name: "List 1",
	})

	hs.RegisterCheck(&service.HealthCheck{
		ID:   "list-2",
		Name: "List 2",
	})

	checks := hs.ListChecks()
	if len(checks) < 2 {
		t.Errorf("Expected at least 2 checks, got %d", len(checks))
	}
}

func TestHealthService_DeleteCheck(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:   "delete-check",
		Name: "Delete Check",
	}
	hs.RegisterCheck(check)

	err := hs.DeleteCheck("delete-check")
	if err != nil {
		t.Fatalf("Failed to delete check: %v", err)
	}

	_, err = hs.GetCheck("delete-check")
	if err == nil {
		t.Error("Expected error for deleted check")
	}
}

func TestHealthService_EnableCheck(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:      "enable-check",
		Name:    "Enable Check",
		Enabled: false,
	}
	hs.RegisterCheck(check)

	err := hs.EnableCheck("enable-check")
	if err != nil {
		t.Fatalf("Failed to enable check: %v", err)
	}

	retrieved, _ := hs.GetCheck("enable-check")
	if !retrieved.Enabled {
		t.Error("Check should be enabled")
	}
}

func TestHealthService_DisableCheck(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:      "disable-check",
		Name:    "Disable Check",
		Enabled: true,
	}
	hs.RegisterCheck(check)

	err := hs.DisableCheck("disable-check")
	if err != nil {
		t.Fatalf("Failed to disable check: %v", err)
	}

	retrieved, _ := hs.GetCheck("disable-check")
	if retrieved.Enabled {
		t.Error("Check should be disabled")
	}
}

func TestHealthService_RecordResult(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:   "result-check",
		Name: "Result Check",
	}
	hs.RegisterCheck(check)

	result := &service.HealthCheckResult{
		CheckID:    "result-check",
		Status:     service.HealthStatusHealthy,
		Message:    "All good",
		Timestamp:  time.Now(),
		Duration:   100 * time.Millisecond,
	}

	err := hs.RecordResult(result)
	if err != nil {
		t.Fatalf("Failed to record result: %v", err)
	}
}

func TestHealthService_GetLastResult(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:   "lastresult-check",
		Name: "Last Result Check",
	}
	hs.RegisterCheck(check)

	hs.RecordResult(&service.HealthCheckResult{
		CheckID: "lastresult-check",
		Status:  service.HealthStatusHealthy,
	})

	result := hs.GetLastResult("lastresult-check")
	if result == nil {
		t.Fatal("Expected result")
	}

	if result.Status != service.HealthStatusHealthy {
		t.Error("Status mismatch")
	}
}

func TestHealthService_GetStatus(t *testing.T) {
	hs := service.NewHealthService()

	hs.RegisterCheck(&service.HealthCheck{
		ID:      "status-check-1",
		Name:    "Status Check 1",
		Enabled: true,
	})

	hs.RecordResult(&service.HealthCheckResult{
		CheckID: "status-check-1",
		Status:  service.HealthStatusHealthy,
	})

	status := hs.GetStatus()
	if status == nil {
		t.Fatal("Expected status")
	}
}

func TestHealthService_GetOverallStatus(t *testing.T) {
	hs := service.NewHealthService()

	hs.RegisterCheck(&service.HealthCheck{
		ID:      "overall-1",
		Name:    "Overall 1",
		Enabled: true,
	})

	hs.RecordResult(&service.HealthCheckResult{
		CheckID: "overall-1",
		Status:  service.HealthStatusHealthy,
	})

	overall := hs.GetOverallStatus()
	if overall != service.HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", overall)
	}
}

func TestHealthService_GetHistory(t *testing.T) {
	hs := service.NewHealthService()

	check := &service.HealthCheck{
		ID:   "history-check",
		Name: "History Check",
	}
	hs.RegisterCheck(check)

	// Record multiple results
	hs.RecordResult(&service.HealthCheckResult{
		CheckID: "history-check",
		Status:  service.HealthStatusHealthy,
	})

	hs.RecordResult(&service.HealthCheckResult{
		CheckID: "history-check",
		Status:  service.HealthStatusDegraded,
	})

	history := hs.GetHistory("history-check", 10)
	if len(history) < 2 {
		t.Errorf("Expected at least 2 history items, got %d", len(history))
	}
}

func TestHealthService_HealthStatuses(t *testing.T) {
	statuses := []service.HealthStatus{
		service.HealthStatusHealthy,
		service.HealthStatusDegraded,
		service.HealthStatusUnhealthy,
		service.HealthStatusUnknown,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestHealthService_HealthCheckToJSON(t *testing.T) {
	check := &service.HealthCheck{
		ID:        "json-check",
		Name:      "JSON Check",
		Interval:  time.Minute,
		Timeout:   5 * time.Second,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	data, err := check.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestHealthService_HealthCheckResultToJSON(t *testing.T) {
	result := &service.HealthCheckResult{
		CheckID:   "result-json",
		Status:    service.HealthStatusHealthy,
		Message:   "All systems operational",
		Timestamp: time.Now(),
		Duration:  50 * time.Millisecond,
	}

	data, err := result.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}