package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Audit Service Tests

func TestAuditService_New(t *testing.T) {
	as := service.NewAuditService()
	if as == nil {
		t.Fatal("Expected non-nil audit service")
	}
}

func TestAuditService_LogEvent(t *testing.T) {
	as := service.NewAuditService()

	event := &service.AuditEvent{
		TenantID:    "tenant-1",
		UserID:      "user-1",
		Action:      "create",
		Resource:    "pipeline",
		ResourceID:  "pipeline-1",
		Details:     map[string]interface{}{"name": "test-pipeline"},
	}

	err := as.LogEvent(event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	if event.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if event.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestAuditService_GetEvent(t *testing.T) {
	as := service.NewAuditService()

	event := &service.AuditEvent{
		TenantID:   "tenant-get",
		UserID:     "user-1",
		Action:     "update",
		Resource:   "task",
		ResourceID: "task-1",
	}
	as.LogEvent(event)

	retrieved, err := as.GetEvent(event.ID)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	if retrieved.Action != "update" {
		t.Error("Action mismatch")
	}
}

func TestAuditService_GetEvent_NotFound(t *testing.T) {
	as := service.NewAuditService()

	_, err := as.GetEvent("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent event")
	}
}

func TestAuditService_ListEvents(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID:   "tenant-list",
		UserID:     "user-1",
		Action:     "create",
		Resource:   "pipeline",
		ResourceID: "p-1",
	})

	as.LogEvent(&service.AuditEvent{
		TenantID:   "tenant-list",
		UserID:     "user-1",
		Action:     "delete",
		Resource:   "pipeline",
		ResourceID: "p-2",
	})

	events := as.ListEvents("tenant-list", "user-1", "", "")
	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}
}

func TestAuditService_ListEvents_ByAction(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-filter",
		UserID:   "user-1",
		Action:   "create",
		Resource: "task",
	})

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-filter",
		UserID:   "user-1",
		Action:   "delete",
		Resource: "task",
	})

	events := as.ListEvents("tenant-filter", "", "create", "")
	for _, e := range events {
		if e.Action != "create" {
			t.Errorf("Expected only create events, got %s", e.Action)
		}
	}
}

func TestAuditService_ListEvents_ByResource(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-res",
		UserID:   "user-1",
		Action:   "create",
		Resource: "pipeline",
	})

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-res",
		UserID:   "user-1",
		Action:   "create",
		Resource: "task",
	})

	events := as.ListEvents("tenant-res", "", "", "pipeline")
	for _, e := range events {
		if e.Resource != "pipeline" {
			t.Errorf("Expected only pipeline events, got %s", e.Resource)
		}
	}
}

func TestAuditService_GetUserActivity(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-activity",
		UserID:   "user-activity",
		Action:   "login",
		Resource: "session",
	})

	activity := as.GetUserActivity("tenant-activity", "user-activity")
	if len(activity) < 1 {
		t.Error("Expected user activity")
	}
}

func TestAuditService_GetResourceHistory(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID:   "tenant-history",
		UserID:     "user-1",
		Action:     "create",
		Resource:   "pipeline",
		ResourceID: "pipeline-history",
	})

	as.LogEvent(&service.AuditEvent{
		TenantID:   "tenant-history",
		UserID:     "user-2",
		Action:     "update",
		Resource:   "pipeline",
		ResourceID: "pipeline-history",
	})

	history := as.GetResourceHistory("tenant-history", "pipeline", "pipeline-history")
	if len(history) < 2 {
		t.Errorf("Expected at least 2 history entries, got %d", len(history))
	}
}

func TestAuditService_DeleteOldEvents(t *testing.T) {
	as := service.NewAuditService()

	// Create old event
	event := &service.AuditEvent{
		TenantID: "tenant-old",
		UserID:   "user-1",
		Action:   "create",
		Resource: "task",
	}
	as.LogEvent(event)

	// Manually set old timestamp
	oldTime := time.Now().Add(-365 * 24 * time.Hour)
	as.SetTimestamp(event.ID, oldTime)

	deleted := as.DeleteOldEvents(30 * 24 * time.Hour)
	if deleted < 1 {
		t.Error("Expected at least 1 old event to be deleted")
	}
}

func TestAuditService_ExportEvents(t *testing.T) {
	as := service.NewAuditService()

	as.LogEvent(&service.AuditEvent{
		TenantID: "tenant-export",
		UserID:   "user-1",
		Action:   "create",
		Resource: "pipeline",
	})

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(24 * time.Hour)

	data, err := as.ExportEvents("tenant-export", start, end, "json")
	if err != nil {
		t.Fatalf("Failed to export events: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected exported data")
	}
}

func TestAuditService_AuditEventToJSON(t *testing.T) {
	event := &service.AuditEvent{
		ID:         "event-1",
		TenantID:   "tenant-1",
		UserID:     "user-1",
		Action:     "create",
		Resource:   "pipeline",
		ResourceID: "p-1",
		Timestamp:  time.Now(),
	}

	data, err := event.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}