package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Task Service Tests

func TestTaskService_New(t *testing.T) {
	ts := service.NewTaskService()
	if ts == nil {
		t.Fatal("Expected non-nil task service")
	}
}

func TestTaskService_Create(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Test Task",
		Type:     "pipeline",
		TenantID: "tenant-1",
	}

	err := ts.Create(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if task.Status != service.TaskStatusPending {
		t.Errorf("Expected status pending, got %s", task.Status)
	}
}

func TestTaskService_Create_MissingName(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		TenantID: "tenant-1",
	}

	err := ts.Create(task)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestTaskService_Get(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Get Test",
		TenantID: "tenant-get",
	}
	ts.Create(task)

	retrieved, err := ts.Get(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.Name != "Get Test" {
		t.Error("Task name mismatch")
	}
}

func TestTaskService_Get_NotFound(t *testing.T) {
	ts := service.NewTaskService()

	_, err := ts.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}
}

func TestTaskService_Update(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Update Test",
		TenantID: "tenant-update",
	}
	ts.Create(task)

	err := ts.Update(task.ID, map[string]interface{}{
		"priority": 10,
		"name":     "Updated Task",
	})

	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Name != "Updated Task" {
		t.Error("Name not updated")
	}
}

func TestTaskService_Delete(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Delete Test",
		TenantID: "tenant-delete",
	}
	ts.Create(task)

	err := ts.Delete(task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	_, err = ts.Get(task.ID)
	if err == nil {
		t.Error("Expected error for deleted task")
	}
}

func TestTaskService_Start(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Start Test",
		TenantID: "tenant-start",
	}
	ts.Create(task)

	err := ts.Start(task.ID)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Status != service.TaskStatusRunning {
		t.Error("Expected status running")
	}
}

func TestTaskService_Complete(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Complete Test",
		TenantID: "tenant-complete",
	}
	ts.Create(task)
	ts.Start(task.ID)

	err := ts.Complete(task.ID, map[string]interface{}{"result": "success"})
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Status != service.TaskStatusCompleted {
		t.Error("Expected status completed")
	}
}

func TestTaskService_Fail(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Fail Test",
		TenantID: "tenant-fail",
	}
	ts.Create(task)
	ts.Start(task.ID)

	err := ts.Fail(task.ID, "Something went wrong")
	if err != nil {
		t.Fatalf("Failed to fail task: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Status != service.TaskStatusFailed {
		t.Error("Expected status failed")
	}

	if retrieved.Error != "Something went wrong" {
		t.Error("Error message mismatch")
	}
}

func TestTaskService_Cancel(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Cancel Test",
		TenantID: "tenant-cancel",
	}
	ts.Create(task)
	ts.Start(task.ID)

	err := ts.Cancel(task.ID)
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Status != service.TaskStatusCancelled {
		t.Error("Expected status cancelled")
	}
}

func TestTaskService_List(t *testing.T) {
	ts := service.NewTaskService()

	ts.Create(&service.Task{
		Name:     "List 1",
		TenantID: "tenant-list",
	})

	ts.Create(&service.Task{
		Name:     "List 2",
		TenantID: "tenant-list",
	})

	ts.Create(&service.Task{
		Name:     "Other",
		TenantID: "other-tenant",
	})

	tasks := ts.List("tenant-list")
	if len(tasks) < 2 {
		t.Errorf("Expected at least 2 tasks, got %d", len(tasks))
	}
}

func TestTaskService_ListByStatus(t *testing.T) {
	ts := service.NewTaskService()

	ts.Create(&service.Task{
		Name:     "Status Pending",
		TenantID: "tenant-status",
	})

	task := &service.Task{
		Name:     "Status Running",
		TenantID: "tenant-status",
	}
	ts.Create(task)
	ts.Start(task.ID)

	pendingTasks := ts.ListByStatus("tenant-status", service.TaskStatusPending)
	for _, t := range pendingTasks {
		if t.Status != service.TaskStatusPending {
			t.Error("Expected only pending tasks")
		}
	}
}

func TestTaskService_SetProgress(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Progress Test",
		TenantID: "tenant-progress",
	}
	ts.Create(task)
	ts.Start(task.ID)

	err := ts.SetProgress(task.ID, 50, "Processing...")
	if err != nil {
		t.Fatalf("Failed to set progress: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Progress != 50 {
		t.Error("Progress not set")
	}
}

func TestTaskService_SetPriority(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Priority Test",
		TenantID: "tenant-priority",
	}
	ts.Create(task)

	err := ts.SetPriority(task.ID, 100)
	if err != nil {
		t.Fatalf("Failed to set priority: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if retrieved.Priority != 100 {
		t.Error("Priority not set")
	}
}

func TestTaskService_SetTags(t *testing.T) {
	ts := service.NewTaskService()

	task := &service.Task{
		Name:     "Tags Test",
		TenantID: "tenant-tags",
	}
	ts.Create(task)

	err := ts.SetTags(task.ID, []string{"urgent", "production"})
	if err != nil {
		t.Fatalf("Failed to set tags: %v", err)
	}

	retrieved, _ := ts.Get(task.ID)
	if len(retrieved.Tags) != 2 {
		t.Error("Tags not set correctly")
	}
}

func TestTaskService_GetStats(t *testing.T) {
	ts := service.NewTaskService()

	ts.Create(&service.Task{Name: "Stats 1", TenantID: "tenant-stats"})
	ts.Create(&service.Task{Name: "Stats 2", TenantID: "tenant-stats"})

	stats := ts.GetStats("tenant-stats")

	if stats.TotalTasks < 2 {
		t.Errorf("Expected at least 2 tasks, got %d", stats.TotalTasks)
	}
}

func TestTaskService_TaskStatuses(t *testing.T) {
	statuses := []service.TaskStatus{
		service.TaskStatusPending,
		service.TaskStatusRunning,
		service.TaskStatusCompleted,
		service.TaskStatusFailed,
		service.TaskStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestTaskService_TaskToJSON(t *testing.T) {
	task := &service.Task{
		ID:        "json-task",
		Name:      "JSON Task",
		Type:      "pipeline",
		TenantID:  "tenant-1",
		Status:    service.TaskStatusCompleted,
		CreatedAt: time.Now(),
	}

	data, err := task.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}