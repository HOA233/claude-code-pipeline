package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Executor Service Tests

func TestExecutorService_New(t *testing.T) {
	es := service.NewExecutorService()
	if es == nil {
		t.Fatal("Expected non-nil executor service")
	}
}

func TestExecutorService_CreateTask(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:      "Test Task",
		Command:   "echo",
		Args:      []string{"hello"},
		TenantID:  "tenant-1",
	}

	err := es.CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestExecutorService_CreateTask_MissingCommand(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "No Command",
		TenantID: "tenant-1",
	}

	err := es.CreateTask(context.Background(), task)
	if err == nil {
		t.Error("Expected error for missing command")
	}
}

func TestExecutorService_GetTask(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:      "Get Test",
		Command:   "ls",
		TenantID:  "tenant-get",
	}
	es.CreateTask(context.Background(), task)

	retrieved, err := es.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.Name != "Get Test" {
		t.Error("Task name mismatch")
	}
}

func TestExecutorService_GetTask_NotFound(t *testing.T) {
	es := service.NewExecutorService()

	_, err := es.GetTask("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}
}

func TestExecutorService_Execute(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Execute Test",
		Command:  "echo",
		Args:     []string{"test-output"},
		TenantID: "tenant-exec",
		Timeout:  10 * time.Second,
	}
	es.CreateTask(context.Background(), task)

	err := es.Execute(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	retrieved, _ := es.GetTask(task.ID)
	if retrieved.Status != service.ExecutorStatusCompleted {
		t.Errorf("Expected status completed, got %s", retrieved.Status)
	}
}

func TestExecutorService_Execute_WithTimeout(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Timeout Test",
		Command:  "sleep",
		Args:     []string{"10"},
		TenantID: "tenant-timeout",
		Timeout:  100 * time.Millisecond,
	}
	es.CreateTask(context.Background(), task)

	err := es.Execute(context.Background(), task.ID)
	// Should timeout
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestExecutorService_CancelTask(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Cancel Test",
		Command:  "sleep",
		Args:     []string{"100"},
		TenantID: "tenant-cancel",
		Timeout:  1 * time.Minute,
	}
	es.CreateTask(context.Background(), task)

	// Start execution in background
	go es.Execute(context.Background(), task.ID)

	// Wait a bit for task to start
	time.Sleep(50 * time.Millisecond)

	err := es.CancelTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}
}

func TestExecutorService_ListTasks(t *testing.T) {
	es := service.NewExecutorService()

	es.CreateTask(context.Background(), &service.ExecutorTask{
		Name:     "List 1",
		Command:  "ls",
		TenantID: "tenant-list",
	})

	es.CreateTask(context.Background(), &service.ExecutorTask{
		Name:     "List 2",
		Command:  "ls",
		TenantID: "tenant-list",
	})

	tasks := es.ListTasks("tenant-list")
	if len(tasks) < 2 {
		t.Errorf("Expected at least 2 tasks, got %d", len(tasks))
	}
}

func TestExecutorService_GetOutput(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Output Test",
		Command:  "echo",
		Args:     []string{"test-output-line"},
		TenantID: "tenant-output",
		Timeout:  10 * time.Second,
	}
	es.CreateTask(context.Background(), task)
	es.Execute(context.Background(), task.ID)

	output := es.GetOutput(task.ID)
	if len(output) == 0 {
		t.Error("Expected output")
	}
}

func TestExecutorService_DeleteTask(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Delete Test",
		Command:  "ls",
		TenantID: "tenant-delete",
	}
	es.CreateTask(context.Background(), task)

	err := es.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	_, err = es.GetTask(task.ID)
	if err == nil {
		t.Error("Expected error for deleted task")
	}
}

func TestExecutorService_RetryTask(t *testing.T) {
	es := service.NewExecutorService()

	task := &service.ExecutorTask{
		Name:     "Retry Test",
		Command:  "echo",
		Args:     []string{"retry"},
		TenantID: "tenant-retry",
		MaxRetries: 3,
		Timeout:   10 * time.Second,
	}
	es.CreateTask(context.Background(), task)

	err := es.Retry(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to retry task: %v", err)
	}
}

func TestExecutorService_GetStats(t *testing.T) {
	es := service.NewExecutorService()

	es.CreateTask(context.Background(), &service.ExecutorTask{
		Name:     "Stats",
		Command:  "ls",
		TenantID: "tenant-stats",
	})

	stats := es.GetStats()
	if stats.TotalTasks < 1 {
		t.Errorf("Expected at least 1 task in stats, got %d", stats.TotalTasks)
	}
}

func TestExecutorService_ExecutorStatuses(t *testing.T) {
	statuses := []service.ExecutorStatus{
		service.ExecutorStatusPending,
		service.ExecutorStatusRunning,
		service.ExecutorStatusCompleted,
		service.ExecutorStatusFailed,
		service.ExecutorStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestExecutorService_ExecutorTaskToJSON(t *testing.T) {
	task := &service.ExecutorTask{
		ID:        "task-1",
		Name:      "Test",
		Command:   "echo",
		Args:      []string{"hello"},
		TenantID:  "tenant-1",
		Status:    service.ExecutorStatusCompleted,
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