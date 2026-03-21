package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Orchestrator Service Tests

func TestOrchestratorService_New(t *testing.T) {
	os := service.NewOrchestratorService()
	if os == nil {
		t.Fatal("Expected non-nil orchestrator service")
	}
}

func TestOrchestratorService_CreatePipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "pipeline-1",
		Name:     "Test Pipeline",
		TenantID: "tenant-1",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Build"},
			{ID: "stage-2", Name: "Test"},
			{ID: "stage-3", Name: "Deploy"},
		},
	}

	err := os.CreatePipeline(pipeline)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
}

func TestOrchestratorService_CreatePipeline_MissingName(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "no-name",
		TenantID: "tenant-1",
	}

	err := os.CreatePipeline(pipeline)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestOrchestratorService_GetPipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "get-pipeline",
		Name:     "Get Pipeline",
		TenantID: "tenant-get",
	}
	os.CreatePipeline(pipeline)

	retrieved, err := os.GetPipeline("get-pipeline")
	if err != nil {
		t.Fatalf("Failed to get pipeline: %v", err)
	}

	if retrieved.Name != "Get Pipeline" {
		t.Error("Pipeline name mismatch")
	}
}

func TestOrchestratorService_GetPipeline_NotFound(t *testing.T) {
	os := service.NewOrchestratorService()

	_, err := os.GetPipeline("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent pipeline")
	}
}

func TestOrchestratorService_ListPipelines(t *testing.T) {
	os := service.NewOrchestratorService()

	os.CreatePipeline(&service.Pipeline{
		ID:       "list-1",
		Name:     "List 1",
		TenantID: "tenant-list",
	})

	os.CreatePipeline(&service.Pipeline{
		ID:       "list-2",
		Name:     "List 2",
		TenantID: "tenant-list",
	})

	pipelines := os.ListPipelines("tenant-list")
	if len(pipelines) < 2 {
		t.Errorf("Expected at least 2 pipelines, got %d", len(pipelines))
	}
}

func TestOrchestratorService_DeletePipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "delete-pipeline",
		Name:     "Delete Pipeline",
		TenantID: "tenant-delete",
	}
	os.CreatePipeline(pipeline)

	err := os.DeletePipeline("delete-pipeline")
	if err != nil {
		t.Fatalf("Failed to delete pipeline: %v", err)
	}

	_, err = os.GetPipeline("delete-pipeline")
	if err == nil {
		t.Error("Expected error for deleted pipeline")
	}
}

func TestOrchestratorService_ExecutePipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "exec-pipeline",
		Name:     "Exec Pipeline",
		TenantID: "tenant-exec",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)

	execution, err := os.ExecutePipeline("exec-pipeline", map[string]interface{}{
		"param": "value",
	})

	if err != nil {
		t.Fatalf("Failed to execute pipeline: %v", err)
	}

	if execution.ID == "" {
		t.Error("Expected execution ID")
	}

	if execution.Status != service.PipelineStatusRunning {
		t.Error("Expected status running")
	}
}

func TestOrchestratorService_GetExecution(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "getexec-pipeline",
		Name:     "Get Exec Pipeline",
		TenantID: "tenant-getexec",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("getexec-pipeline", nil)

	retrieved, err := os.GetExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}

	if retrieved.PipelineID != "getexec-pipeline" {
		t.Error("Pipeline ID mismatch")
	}
}

func TestOrchestratorService_CancelExecution(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "cancel-pipeline",
		Name:     "Cancel Pipeline",
		TenantID: "tenant-cancel",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("cancel-pipeline", nil)

	err := os.CancelExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to cancel execution: %v", err)
	}

	retrieved, _ := os.GetExecution(execution.ID)
	if retrieved.Status != service.PipelineStatusCancelled {
		t.Error("Expected status cancelled")
	}
}

func TestOrchestratorService_PauseExecution(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "pause-pipeline",
		Name:     "Pause Pipeline",
		TenantID: "tenant-pause",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("pause-pipeline", nil)

	err := os.PauseExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to pause execution: %v", err)
	}

	retrieved, _ := os.GetExecution(execution.ID)
	if retrieved.Status != service.PipelineStatusPaused {
		t.Error("Expected status paused")
	}
}

func TestOrchestratorService_ResumeExecution(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "resume-pipeline",
		Name:     "Resume Pipeline",
		TenantID: "tenant-resume",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("resume-pipeline", nil)
	os.PauseExecution(execution.ID)

	err := os.ResumeExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to resume execution: %v", err)
	}

	retrieved, _ := os.GetExecution(execution.ID)
	if retrieved.Status != service.PipelineStatusRunning {
		t.Error("Expected status running after resume")
	}
}

func TestOrchestratorService_CompleteStage(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "completestage-pipeline",
		Name:     "Complete Stage Pipeline",
		TenantID: "tenant-completestage",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
			{ID: "stage-2", Name: "Stage 2"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("completestage-pipeline", nil)

	err := os.CompleteStage(execution.ID, "stage-1", map[string]interface{}{
		"result": "success",
	})

	if err != nil {
		t.Fatalf("Failed to complete stage: %v", err)
	}

	retrieved, _ := os.GetExecution(execution.ID)
	if retrieved.CurrentStage != "stage-2" {
		t.Error("Expected current stage to advance")
	}
}

func TestOrchestratorService_FailStage(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "failstage-pipeline",
		Name:     "Fail Stage Pipeline",
		TenantID: "tenant-failstage",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	execution, _ := os.ExecutePipeline("failstage-pipeline", nil)

	err := os.FailStage(execution.ID, "stage-1", "Stage failed")
	if err != nil {
		t.Fatalf("Failed to fail stage: %v", err)
	}

	retrieved, _ := os.GetExecution(execution.ID)
	if retrieved.Status != service.PipelineStatusFailed {
		t.Error("Expected status failed")
	}
}

func TestOrchestratorService_ListExecutions(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "listexec-pipeline",
		Name:     "List Exec Pipeline",
		TenantID: "tenant-listexec",
		Stages: []service.Stage{
			{ID: "stage-1", Name: "Stage 1"},
		},
	}
	os.CreatePipeline(pipeline)
	os.ExecutePipeline("listexec-pipeline", nil)
	os.ExecutePipeline("listexec-pipeline", nil)

	executions := os.ListExecutions("tenant-listexec")
	if len(executions) < 2 {
		t.Errorf("Expected at least 2 executions, got %d", len(executions))
	}
}

func TestOrchestratorService_EnablePipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "enable-pipeline",
		Name:     "Enable Pipeline",
		TenantID: "tenant-enable",
		Enabled:  false,
	}
	os.CreatePipeline(pipeline)

	err := os.EnablePipeline("enable-pipeline")
	if err != nil {
		t.Fatalf("Failed to enable pipeline: %v", err)
	}

	retrieved, _ := os.GetPipeline("enable-pipeline")
	if !retrieved.Enabled {
		t.Error("Pipeline should be enabled")
	}
}

func TestOrchestratorService_DisablePipeline(t *testing.T) {
	os := service.NewOrchestratorService()

	pipeline := &service.Pipeline{
		ID:       "disable-pipeline",
		Name:     "Disable Pipeline",
		TenantID: "tenant-disable",
		Enabled:  true,
	}
	os.CreatePipeline(pipeline)

	err := os.DisablePipeline("disable-pipeline")
	if err != nil {
		t.Fatalf("Failed to disable pipeline: %v", err)
	}

	retrieved, _ := os.GetPipeline("disable-pipeline")
	if retrieved.Enabled {
		t.Error("Pipeline should be disabled")
	}
}

func TestOrchestratorService_PipelineStatuses(t *testing.T) {
	statuses := []service.PipelineStatus{
		service.PipelineStatusPending,
		service.PipelineStatusRunning,
		service.PipelineStatusPaused,
		service.PipelineStatusCompleted,
		service.PipelineStatusFailed,
		service.PipelineStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestOrchestratorService_PipelineToJSON(t *testing.T) {
	pipeline := &service.Pipeline{
		ID:        "json-pipeline",
		Name:      "JSON Pipeline",
		TenantID:  "tenant-1",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	data, err := pipeline.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}