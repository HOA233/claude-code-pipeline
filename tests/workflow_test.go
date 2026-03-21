package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Workflow Service Tests

func TestWorkflowService_New(t *testing.T) {
	ws := service.NewWorkflowService()
	if ws == nil {
		t.Fatal("Expected non-nil workflow service")
	}
}

func TestWorkflowService_Create(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:          "workflow-1",
		Name:        "Test Workflow",
		Description: "A test workflow",
		TenantID:    "tenant-1",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "First Step", Type: "task"},
			{ID: "step-2", Name: "Second Step", Type: "task"},
		},
	}

	err := ws.Create(workflow)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
}

func TestWorkflowService_Create_MissingName(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "no-name",
		TenantID: "tenant-1",
	}

	err := ws.Create(workflow)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestWorkflowService_Get(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "get-workflow",
		Name:     "Get Workflow",
		TenantID: "tenant-get",
	}
	ws.Create(workflow)

	retrieved, err := ws.Get("get-workflow")
	if err != nil {
		t.Fatalf("Failed to get workflow: %v", err)
	}

	if retrieved.Name != "Get Workflow" {
		t.Error("Workflow name mismatch")
	}
}

func TestWorkflowService_Get_NotFound(t *testing.T) {
	ws := service.NewWorkflowService()

	_, err := ws.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
}

func TestWorkflowService_List(t *testing.T) {
	ws := service.NewWorkflowService()

	ws.Create(&service.Workflow{
		ID:       "list-1",
		Name:     "List 1",
		TenantID: "tenant-list",
	})

	ws.Create(&service.Workflow{
		ID:       "list-2",
		Name:     "List 2",
		TenantID: "tenant-list",
	})

	workflows := ws.List("tenant-list")
	if len(workflows) < 2 {
		t.Errorf("Expected at least 2 workflows, got %d", len(workflows))
	}
}

func TestWorkflowService_Update(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:          "update-workflow",
		Name:        "Update Workflow",
		Description: "Original description",
		TenantID:    "tenant-update",
	}
	ws.Create(workflow)

	err := ws.Update("update-workflow", &service.Workflow{
		Description: "Updated description",
	})

	if err != nil {
		t.Fatalf("Failed to update workflow: %v", err)
	}

	retrieved, _ := ws.Get("update-workflow")
	if retrieved.Description != "Updated description" {
		t.Error("Description not updated")
	}
}

func TestWorkflowService_Delete(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "delete-workflow",
		Name:     "Delete Workflow",
		TenantID: "tenant-delete",
	}
	ws.Create(workflow)

	err := ws.Delete("delete-workflow")
	if err != nil {
		t.Fatalf("Failed to delete workflow: %v", err)
	}

	_, err = ws.Get("delete-workflow")
	if err == nil {
		t.Error("Expected error for deleted workflow")
	}
}

func TestWorkflowService_Start(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "start-workflow",
		Name:     "Start Workflow",
		TenantID: "tenant-start",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)

	execution, err := ws.Start("start-workflow", map[string]interface{}{"param": "value"})
	if err != nil {
		t.Fatalf("Failed to start workflow: %v", err)
	}

	if execution.ID == "" {
		t.Error("Expected execution ID")
	}

	if execution.Status != service.WorkflowStatusRunning {
		t.Error("Expected status running")
	}
}

func TestWorkflowService_GetExecution(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "exec-workflow",
		Name:     "Exec Workflow",
		TenantID: "tenant-exec",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)
	execution, _ := ws.Start("exec-workflow", nil)

	retrieved, err := ws.GetExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}

	if retrieved.WorkflowID != "exec-workflow" {
		t.Error("Execution workflow ID mismatch")
	}
}

func TestWorkflowService_CancelExecution(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "cancel-workflow",
		Name:     "Cancel Workflow",
		TenantID: "tenant-cancel",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)
	execution, _ := ws.Start("cancel-workflow", nil)

	err := ws.CancelExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to cancel execution: %v", err)
	}

	retrieved, _ := ws.GetExecution(execution.ID)
	if retrieved.Status != service.WorkflowStatusCancelled {
		t.Error("Expected status cancelled")
	}
}

func TestWorkflowService_CompleteStep(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "step-complete-workflow",
		Name:     "Step Complete Workflow",
		TenantID: "tenant-step",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
			{ID: "step-2", Name: "Step 2", Type: "task"},
		},
	}
	ws.Create(workflow)
	execution, _ := ws.Start("step-complete-workflow", nil)

	err := ws.CompleteStep(execution.ID, "step-1", map[string]interface{}{"result": "success"})
	if err != nil {
		t.Fatalf("Failed to complete step: %v", err)
	}

	retrieved, _ := ws.GetExecution(execution.ID)
	if retrieved.CurrentStep != "step-2" {
		t.Error("Expected current step to advance")
	}
}

func TestWorkflowService_FailStep(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "fail-step-workflow",
		Name:     "Fail Step Workflow",
		TenantID: "tenant-failstep",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)
	execution, _ := ws.Start("fail-step-workflow", nil)

	err := ws.FailStep(execution.ID, "step-1", "Step failed")
	if err != nil {
		t.Fatalf("Failed to fail step: %v", err)
	}

	retrieved, _ := ws.GetExecution(execution.ID)
	if retrieved.Status != service.WorkflowStatusFailed {
		t.Error("Expected status failed")
	}
}

func TestWorkflowService_ListExecutions(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "list-exec-workflow",
		Name:     "List Exec Workflow",
		TenantID: "tenant-listexec",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)
	ws.Start("list-exec-workflow", nil)
	ws.Start("list-exec-workflow", nil)

	executions := ws.ListExecutions("tenant-listexec")
	if len(executions) < 2 {
		t.Errorf("Expected at least 2 executions, got %d", len(executions))
	}
}

func TestWorkflowService_Enable(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "enable-workflow",
		Name:     "Enable Workflow",
		TenantID: "tenant-enable",
		Enabled:  false,
	}
	ws.Create(workflow)

	err := ws.Enable("enable-workflow")
	if err != nil {
		t.Fatalf("Failed to enable workflow: %v", err)
	}

	retrieved, _ := ws.Get("enable-workflow")
	if !retrieved.Enabled {
		t.Error("Workflow should be enabled")
	}
}

func TestWorkflowService_Disable(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "disable-workflow",
		Name:     "Disable Workflow",
		TenantID: "tenant-disable",
		Enabled:  true,
	}
	ws.Create(workflow)

	err := ws.Disable("disable-workflow")
	if err != nil {
		t.Fatalf("Failed to disable workflow: %v", err)
	}

	retrieved, _ := ws.Get("disable-workflow")
	if retrieved.Enabled {
		t.Error("Workflow should be disabled")
	}
}

func TestWorkflowService_GetStats(t *testing.T) {
	ws := service.NewWorkflowService()

	workflow := &service.Workflow{
		ID:       "stats-workflow",
		Name:     "Stats Workflow",
		TenantID: "tenant-stats",
		Steps: []service.WorkflowStep{
			{ID: "step-1", Name: "Step 1", Type: "task"},
		},
	}
	ws.Create(workflow)
	ws.Start("stats-workflow", nil)

	stats := ws.GetStats("tenant-stats")
	if stats.TotalWorkflows < 1 {
		t.Errorf("Expected at least 1 workflow, got %d", stats.TotalWorkflows)
	}
}

func TestWorkflowService_WorkflowStatuses(t *testing.T) {
	statuses := []service.WorkflowStatus{
		service.WorkflowStatusPending,
		service.WorkflowStatusRunning,
		service.WorkflowStatusCompleted,
		service.WorkflowStatusFailed,
		service.WorkflowStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestWorkflowService_WorkflowToJSON(t *testing.T) {
	workflow := &service.Workflow{
		ID:          "json-workflow",
		Name:        "JSON Workflow",
		Description: "Test",
		TenantID:    "tenant-1",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	data, err := workflow.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}