package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Batch Service Tests

func TestBatchService_New(t *testing.T) {
	bs := service.NewBatchService()
	if bs == nil {
		t.Fatal("Expected non-nil batch service")
	}
}

func TestBatchService_CreateJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Test Batch",
		TenantID: "tenant-1",
		Items:    []string{"item1", "item2", "item3"},
	}

	err := bs.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if job.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if job.Status != service.BatchStatusPending {
		t.Errorf("Expected status pending, got %s", job.Status)
	}
}

func TestBatchService_CreateJob_MissingName(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		TenantID: "tenant-1",
		Items:    []string{"item1"},
	}

	err := bs.CreateJob(job)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestBatchService_GetJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Get Test",
		TenantID: "tenant-get",
		Items:    []string{"item1"},
	}
	bs.CreateJob(job)

	retrieved, err := bs.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.Name != "Get Test" {
		t.Error("Job name mismatch")
	}
}

func TestBatchService_GetJob_NotFound(t *testing.T) {
	bs := service.NewBatchService()

	_, err := bs.GetJob("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job")
	}
}

func TestBatchService_StartJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Start Test",
		TenantID: "tenant-start",
		Items:    []string{"item1", "item2"},
	}
	bs.CreateJob(job)

	err := bs.StartJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	retrieved, _ := bs.GetJob(job.ID)
	if retrieved.Status != service.BatchStatusRunning {
		t.Error("Expected status to be running")
	}
}

func TestBatchService_CompleteItem(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Complete Item Test",
		TenantID: "tenant-complete",
		Items:    []string{"item1", "item2", "item3"},
	}
	bs.CreateJob(job)
	bs.StartJob(job.ID)

	err := bs.CompleteItem(job.ID, "item1", nil)
	if err != nil {
		t.Fatalf("Failed to complete item: %v", err)
	}

	retrieved, _ := bs.GetJob(job.ID)
	if retrieved.CompletedCount != 1 {
		t.Errorf("Expected 1 completed item, got %d", retrieved.CompletedCount)
	}
}

func TestBatchService_FailItem(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Fail Item Test",
		TenantID: "tenant-fail",
		Items:    []string{"item1", "item2"},
	}
	bs.CreateJob(job)
	bs.StartJob(job.ID)

	err := bs.FailItem(job.ID, "item1", "processing error")
	if err != nil {
		t.Fatalf("Failed to fail item: %v", err)
	}

	retrieved, _ := bs.GetJob(job.ID)
	if retrieved.FailedCount != 1 {
		t.Errorf("Expected 1 failed item, got %d", retrieved.FailedCount)
	}
}

func TestBatchService_CompleteJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Complete Test",
		TenantID: "tenant-complete",
		Items:    []string{"item1"},
	}
	bs.CreateJob(job)
	bs.StartJob(job.ID)
	bs.CompleteItem(job.ID, "item1", nil)

	// Job should auto-complete when all items are done
	retrieved, _ := bs.GetJob(job.ID)
	if retrieved.Status != service.BatchStatusCompleted {
		t.Error("Expected job to be completed")
	}
}

func TestBatchService_CancelJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Cancel Test",
		TenantID: "tenant-cancel",
		Items:    []string{"item1", "item2"},
	}
	bs.CreateJob(job)
	bs.StartJob(job.ID)

	err := bs.CancelJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to cancel job: %v", err)
	}

	retrieved, _ := bs.GetJob(job.ID)
	if retrieved.Status != service.BatchStatusCancelled {
		t.Error("Expected status to be cancelled")
	}
}

func TestBatchService_ListJobs(t *testing.T) {
	bs := service.NewBatchService()

	bs.CreateJob(&service.BatchJob{
		Name:     "List 1",
		TenantID: "tenant-list",
		Items:    []string{"item1"},
	})

	bs.CreateJob(&service.BatchJob{
		Name:     "List 2",
		TenantID: "tenant-list",
		Items:    []string{"item1"},
	})

	jobs := bs.ListJobs("tenant-list")
	if len(jobs) < 2 {
		t.Errorf("Expected at least 2 jobs, got %d", len(jobs))
	}
}

func TestBatchService_GetProgress(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Progress Test",
		TenantID: "tenant-progress",
		Items:    []string{"item1", "item2", "item3"},
	}
	bs.CreateJob(job)
	bs.StartJob(job.ID)
	bs.CompleteItem(job.ID, "item1", nil)

	progress := bs.GetProgress(job.ID)

	if progress.Total != 3 {
		t.Errorf("Expected total 3, got %d", progress.Total)
	}

	if progress.Completed != 1 {
		t.Errorf("Expected 1 completed, got %d", progress.Completed)
	}

	if progress.Percent < 33 || progress.Percent > 34 {
		t.Errorf("Expected ~33%%, got %.1f%%", progress.Percent)
	}
}

func TestBatchService_DeleteJob(t *testing.T) {
	bs := service.NewBatchService()

	job := &service.BatchJob{
		Name:     "Delete Test",
		TenantID: "tenant-delete",
		Items:    []string{"item1"},
	}
	bs.CreateJob(job)

	err := bs.DeleteJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	_, err = bs.GetJob(job.ID)
	if err == nil {
		t.Error("Expected error for deleted job")
	}
}

func TestBatchService_BatchStatuses(t *testing.T) {
	statuses := []service.BatchStatus{
		service.BatchStatusPending,
		service.BatchStatusRunning,
		service.BatchStatusCompleted,
		service.BatchStatusFailed,
		service.BatchStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestBatchService_BatchJobToJSON(t *testing.T) {
	job := &service.BatchJob{
		ID:            "job-1",
		Name:          "Test Job",
		TenantID:      "tenant-1",
		Status:        service.BatchStatusCompleted,
		Items:         []string{"item1", "item2"},
		CompletedCount: 2,
		TotalCount:    2,
		CreatedAt:     time.Now(),
	}

	data, err := job.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}