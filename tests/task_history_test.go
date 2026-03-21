package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// TaskHistory Service Tests

func TestTaskHistory_New(t *testing.T) {
	th := service.NewTaskHistoryService()
	if th == nil {
		t.Fatal("Expected non-nil task history service")
	}
}

func TestTaskHistory_Record(t *testing.T) {
	th := service.NewTaskHistoryService()

	record := &service.TaskHistoryRecord{
		TaskID:     "task-1",
		TenantID:   "tenant-1",
		StatusFrom: "pending",
		StatusTo:   "running",
		Message:    "Task started",
		ChangedBy:  "system",
	}

	err := th.Record(record)
	if err != nil {
		t.Fatalf("Failed to record: %v", err)
	}

	if record.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if record.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestTaskHistory_GetTaskHistory(t *testing.T) {
	th := service.NewTaskHistoryService()

	taskID := "history-task"

	th.Record(&service.TaskHistoryRecord{
		TaskID:     taskID,
		TenantID:   "tenant-history",
		StatusFrom: "pending",
		StatusTo:   "running",
		Message:    "Started",
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:     taskID,
		TenantID:   "tenant-history",
		StatusFrom: "running",
		StatusTo:   "completed",
		Message:    "Finished",
	})

	records := th.GetTaskHistory(taskID)
	if len(records) < 2 {
		t.Errorf("Expected at least 2 records, got %d", len(records))
	}
}

func TestTaskHistory_GetRecord(t *testing.T) {
	th := service.NewTaskHistoryService()

	record := &service.TaskHistoryRecord{
		TaskID:   "get-record-task",
		TenantID: "tenant-get",
		Message:  "Test record",
	}
	th.Record(record)

	retrieved, err := th.GetRecord(record.ID)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if retrieved.Message != "Test record" {
		t.Error("Record message mismatch")
	}
}

func TestTaskHistory_GetRecord_NotFound(t *testing.T) {
	th := service.NewTaskHistoryService()

	_, err := th.GetRecord("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent record")
	}
}

func TestTaskHistory_ListByTenant(t *testing.T) {
	th := service.NewTaskHistoryService()

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "tenant-list-1",
		TenantID: "tenant-list",
		Message:  "Record 1",
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "tenant-list-2",
		TenantID: "tenant-list",
		Message:  "Record 2",
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "other-task",
		TenantID: "other-tenant",
		Message:  "Other record",
	})

	records := th.ListByTenant("tenant-list")
	if len(records) < 2 {
		t.Errorf("Expected at least 2 records, got %d", len(records))
	}
}

func TestTaskHistory_ListByStatus(t *testing.T) {
	th := service.NewTaskHistoryService()

	th.Record(&service.TaskHistoryRecord{
		TaskID:     "status-1",
		TenantID:   "tenant-status",
		StatusTo:   "completed",
		Message:    "Completed task",
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:     "status-2",
		TenantID:   "tenant-status",
		StatusTo:   "failed",
		Message:    "Failed task",
	})

	completed := th.ListByStatus("tenant-status", "completed")
	for _, r := range completed {
		if r.StatusTo != "completed" {
			t.Error("Expected only completed status records")
		}
	}
}

func TestTaskHistory_ListByTimeRange(t *testing.T) {
	th := service.NewTaskHistoryService()

	now := time.Now()

	th.Record(&service.TaskHistoryRecord{
		TaskID:    "time-1",
		TenantID:  "tenant-time",
		Message:   "In range",
		Timestamp: now,
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:    "time-2",
		TenantID:  "tenant-time",
		Message:   "Old",
		Timestamp: now.Add(-48 * time.Hour),
	})

	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	records := th.ListByTimeRange("tenant-time", start, end)
	if len(records) < 1 {
		t.Error("Expected at least 1 record in time range")
	}
}

func TestTaskHistory_DeleteOldRecords(t *testing.T) {
	th := service.NewTaskHistoryService()

	// Record with old timestamp
	record := &service.TaskHistoryRecord{
		TaskID:    "old-record",
		TenantID:  "tenant-old",
		Message:   "Old",
		Timestamp: time.Now().Add(-365 * 24 * time.Hour),
	}
	th.Record(record)

	// Recent record
	th.Record(&service.TaskHistoryRecord{
		TaskID:    "recent-record",
		TenantID:  "tenant-old",
		Message:   "Recent",
		Timestamp: time.Now(),
	})

	deleted := th.DeleteOldRecords(30 * 24 * time.Hour)
	if deleted < 1 {
		t.Error("Expected at least 1 old record to be deleted")
	}
}

func TestTaskHistory_GetStats(t *testing.T) {
	th := service.NewTaskHistoryService()

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "stats-1",
		TenantID: "tenant-stats",
		StatusTo: "completed",
	})

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "stats-2",
		TenantID: "tenant-stats",
		StatusTo: "failed",
	})

	stats := th.GetStats("tenant-stats")

	if stats.TotalRecords < 2 {
		t.Errorf("Expected at least 2 records, got %d", stats.TotalRecords)
	}
}

func TestTaskHistory_ExportHistory(t *testing.T) {
	th := service.NewTaskHistoryService()

	th.Record(&service.TaskHistoryRecord{
		TaskID:   "export-1",
		TenantID: "tenant-export",
		Message:  "Export test",
	})

	data, err := th.ExportHistory("tenant-export", "json")
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected exported data")
	}
}

func TestTaskHistory_TaskHistoryRecordToJSON(t *testing.T) {
	record := &service.TaskHistoryRecord{
		ID:         "record-1",
		TaskID:     "task-1",
		TenantID:   "tenant-1",
		StatusFrom: "pending",
		StatusTo:   "running",
		Message:    "Task started",
		Timestamp:  time.Now(),
		ChangedBy:  "system",
	}

	data, err := record.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}