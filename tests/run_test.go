package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/service"
)

func TestRunService_New(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()

	rs := service.NewRunService(executor, eventBus)
	if rs == nil {
		t.Fatal("Expected non-nil run service")
	}
}

func TestRunService_RegisterPipeline(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "test-pipeline",
		Name: "Test Pipeline",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", Action: "analyze"},
			{ID: "step2", CLI: "claude", Action: "review", DependsOn: []string{"step1"}},
		},
	}

	err := rs.RegisterPipeline(pipeline)
	if err != nil {
		t.Fatalf("Failed to register pipeline: %v", err)
	}
}

func TestRunService_RegisterPipeline_NoSteps(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:    "no-steps",
		Name:  "No Steps Pipeline",
		Mode:  model.ModeSerial,
		Steps: []model.Step{},
	}

	err := rs.RegisterPipeline(pipeline)
	if err == nil {
		t.Error("Expected error for pipeline with no steps")
	}
}

func TestRunService_RegisterPipeline_CircularDependency(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "circular",
		Name: "Circular Dependency",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", DependsOn: []string{"step2"}},
			{ID: "step2", CLI: "claude", DependsOn: []string{"step1"}},
		},
	}

	err := rs.RegisterPipeline(pipeline)
	if err == nil {
		t.Error("Expected error for circular dependency")
	}
}

func TestRunService_CreateRun(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "create-run-test",
		Name: "Create Run Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", Action: "test"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, err := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "create-run-test",
		Params: map[string]interface{}{
			"target": "src/",
		},
	})

	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	if run.Status != model.RunStatusPending {
		t.Errorf("Expected pending status, got: %s", run.Status)
	}
}

func TestRunService_CreateRun_PipelineNotFound(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	_, err := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "nonexistent",
	})

	if err == nil {
		t.Error("Expected error for nonexistent pipeline")
	}
}

func TestRunService_GetRun(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "get-run-test",
		Name: "Get Run Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "get-run-test",
	})

	retrieved, err := rs.GetRun(run.ID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if retrieved.ID != run.ID {
		t.Error("Run ID mismatch")
	}
}

func TestRunService_CancelRun(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "cancel-run-test",
		Name: "Cancel Run Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "cancel-run-test",
	})

	err := rs.CancelRun(run.ID)
	if err != nil {
		t.Fatalf("Failed to cancel run: %v", err)
	}

	updated, _ := rs.GetRun(run.ID)
	if updated.Status != model.RunStatusCancelled {
		t.Errorf("Expected cancelled status, got: %s", updated.Status)
	}
}

func TestRunService_ListRuns(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "list-runs-test",
		Name: "List Runs Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude"},
		},
	}
	rs.RegisterPipeline(pipeline)

	// Create multiple runs
	for i := 0; i < 3; i++ {
		rs.CreateRun(context.Background(), &model.RunCreateRequest{
			PipelineID: "list-runs-test",
		})
	}

	runs := rs.ListRuns("list-runs-test")
	if len(runs) < 3 {
		t.Errorf("Expected at least 3 runs, got %d", len(runs))
	}
}

func TestRunService_ListPipelines(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	// Register multiple pipelines
	for i := 0; i < 3; i++ {
		rs.RegisterPipeline(&model.Pipeline{
			ID:   string(rune('a' + i)),
			Name: "Pipeline",
			Mode: model.ModeSerial,
			Steps: []model.Step{
				{ID: "step1", CLI: "claude"},
			},
		})
	}

	pipelines := rs.ListPipelines()
	if len(pipelines) < 3 {
		t.Errorf("Expected at least 3 pipelines, got %d", len(pipelines))
	}
}

func TestRunService_DeletePipeline(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "delete-test",
		Name: "Delete Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude"},
		},
	}
	rs.RegisterPipeline(pipeline)

	err := rs.DeletePipeline("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete pipeline: %v", err)
	}

	_, err = rs.GetPipeline("delete-test")
	if err == nil {
		t.Error("Expected error for deleted pipeline")
	}
}

func TestRunService_StartRun_Serial(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "serial-run-test",
		Name: "Serial Run Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", Action: "analyze"},
			{ID: "step2", CLI: "claude", Action: "review"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "serial-run-test",
	})

	err := rs.StartRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	updated, _ := rs.GetRun(run.ID)
	if updated.Status != model.RunStatusCompleted {
		t.Errorf("Expected completed status, got: %s", updated.Status)
	}

	if len(updated.StepResults) != 2 {
		t.Errorf("Expected 2 step results, got %d", len(updated.StepResults))
	}
}

func TestRunService_StartRun_Parallel(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "parallel-run-test",
		Name: "Parallel Run Test",
		Mode: model.ModeParallel,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", Action: "analyze"},
			{ID: "step2", CLI: "claude", Action: "review"},
			{ID: "step3", CLI: "claude", Action: "test"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "parallel-run-test",
	})

	err := rs.StartRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	updated, _ := rs.GetRun(run.ID)
	if updated.Status != model.RunStatusCompleted {
		t.Errorf("Expected completed status, got: %s", updated.Status)
	}
}

func TestRunService_StartRun_Hybrid(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "hybrid-run-test",
		Name: "Hybrid Run Test",
		Mode: model.ModeHybrid,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude", Action: "init"},
			{ID: "step2", CLI: "claude", Action: "analyze", DependsOn: []string{"step1"}},
			{ID: "step3", CLI: "claude", Action: "review", DependsOn: []string{"step1"}},
			{ID: "step4", CLI: "claude", Action: "finalize", DependsOn: []string{"step2", "step3"}},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "hybrid-run-test",
	})

	err := rs.StartRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	updated, _ := rs.GetRun(run.ID)
	if updated.Status != model.RunStatusCompleted {
		t.Errorf("Expected completed status, got: %s", updated.Status)
	}
}

func TestRunService_GetSession(t *testing.T) {
	executor := service.NewCLIExecutor(nil, service.CLIConfig{})
	eventBus := service.NewEventBus()
	rs := service.NewRunService(executor, eventBus)

	pipeline := &model.Pipeline{
		ID:   "session-test",
		Name: "Session Test",
		Mode: model.ModeSerial,
		Steps: []model.Step{
			{ID: "step1", CLI: "claude"},
		},
	}
	rs.RegisterPipeline(pipeline)

	run, _ := rs.CreateRun(context.Background(), &model.RunCreateRequest{
		PipelineID: "session-test",
	})

	session, err := rs.GetSession(run.SessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.PipelineID != "session-test" {
		t.Error("Session pipeline ID mismatch")
	}
}