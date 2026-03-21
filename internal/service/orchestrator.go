package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/model"
	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
	"github.com/google/uuid"
)

// Orchestrator manages pipeline execution
type Orchestrator struct {
	redis    *repository.RedisClient
	executor *CLIExecutor
	runs     sync.Map // active runs
	stopChan chan struct{}
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(redis *repository.RedisClient, executor *CLIExecutor) *Orchestrator {
	return &Orchestrator{
		redis:    redis,
		executor: executor,
		stopChan: make(chan struct{}),
	}
}

// CreatePipeline creates a new pipeline
func (o *Orchestrator) CreatePipeline(ctx context.Context, req *model.PipelineCreateRequest) (*model.Pipeline, error) {
	now := time.Now()
	pipeline := &model.Pipeline{
		ID:          "pipeline-" + uuid.New().String()[:8],
		Name:        req.Name,
		Description: req.Description,
		Mode:        req.Mode,
		Steps:       req.Steps,
		ErrorConfig: req.ErrorConfig,
		Output:      req.Output,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Validate steps
	if err := o.validateSteps(pipeline); err != nil {
		return nil, err
	}

	if err := o.redis.SavePipeline(ctx, pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// GetPipeline retrieves a pipeline
func (o *Orchestrator) GetPipeline(ctx context.Context, id string) (*model.Pipeline, error) {
	return o.redis.GetPipeline(ctx, id)
}

// ListPipelines lists all pipelines
func (o *Orchestrator) ListPipelines(ctx context.Context) ([]*model.Pipeline, error) {
	return o.redis.GetAllPipelines(ctx)
}

// DeletePipeline deletes a pipeline
func (o *Orchestrator) DeletePipeline(ctx context.Context, id string) error {
	return o.redis.DeletePipeline(ctx, id)
}

// RunPipeline executes a pipeline
func (o *Orchestrator) RunPipeline(ctx context.Context, req *model.RunCreateRequest) (*model.Run, error) {
	pipeline, err := o.redis.GetPipeline(ctx, req.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %s", req.PipelineID)
	}

	now := time.Now()
	run := &model.Run{
		ID:          "run-" + uuid.New().String()[:8],
		PipelineID:  req.PipelineID,
		Status:      model.RunStatusPending,
		StepResults: make([]model.StepResult, 0),
		CreatedAt:   now,
	}

	// Save run
	if err := o.redis.SaveRun(ctx, run); err != nil {
		return nil, err
	}

	// Add to execution queue
	o.redis.PushRunQueue(ctx, run.ID)

	// Store run reference
	o.runs.Store(run.ID, run)

	return run, nil
}

// GetRun retrieves a run
func (o *Orchestrator) GetRun(ctx context.Context, id string) (*model.Run, error) {
	return o.redis.GetRun(ctx, id)
}

// ListRuns lists all runs
func (o *Orchestrator) ListRuns(ctx context.Context) ([]*model.Run, error) {
	return o.redis.GetAllRuns(ctx)
}

// CancelRun cancels a running pipeline
func (o *Orchestrator) CancelRun(ctx context.Context, id string) error {
	run, err := o.redis.GetRun(ctx, id)
	if err != nil {
		return err
	}

	if run.Status != model.RunStatusRunning && run.Status != model.RunStatusPending {
		return fmt.Errorf("cannot cancel run in status: %s", run.Status)
	}

	run.Status = model.RunStatusCancelled
	now := time.Now()
	run.CompletedAt = &now
	o.redis.SaveRun(ctx, run)

	return nil
}

// StartConsumer starts the pipeline execution consumer
func (o *Orchestrator) StartConsumer(ctx context.Context) {
	logger.Info("Pipeline orchestrator started")

	for {
		select {
		case <-o.stopChan:
			return
		default:
			runID, err := o.redis.PopRunQueue(ctx)
			if err != nil {
				logger.Error("Failed to get run: ", err)
				time.Sleep(time.Second)
				continue
			}

			if runID == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			go o.executeRun(ctx, runID)
		}
	}
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	close(o.stopChan)
}

// executeRun executes a pipeline run
func (o *Orchestrator) executeRun(ctx context.Context, runID string) {
	run, err := o.redis.GetRun(ctx, runID)
	if err != nil {
		logger.Error("Failed to get run: ", err)
		return
	}

	pipeline, err := o.redis.GetPipeline(ctx, run.PipelineID)
	if err != nil {
		o.failRun(ctx, run, "Pipeline not found")
		return
	}

	// Update status
	now := time.Now()
	run.Status = model.RunStatusRunning
	run.StartedAt = &now
	o.redis.SaveRun(ctx, run)

	// Execute based on mode
	var execErr error
	switch pipeline.Mode {
	case model.ModeSerial:
		execErr = o.executeSerial(ctx, pipeline, run)
	case model.ModeParallel:
		execErr = o.executeParallel(ctx, pipeline, run)
	case model.ModeHybrid:
		execErr = o.executeHybrid(ctx, pipeline, run)
	}

	// Update final status
	completedAt := time.Now()
	run.CompletedAt = &completedAt
	if execErr != nil {
		run.Status = model.RunStatusFailed
		run.Error = execErr.Error()
	} else {
		run.Status = model.RunStatusCompleted
	}

	if run.StartedAt != nil {
		run.Duration = completedAt.Sub(*run.StartedAt).Milliseconds()
	}

	o.redis.SaveRun(ctx, run)

	// Merge outputs
	o.mergeOutputs(ctx, pipeline, run)

	logger.Infof("Pipeline run completed: %s, status: %s, duration: %dms",
		runID, run.Status, run.Duration)
}

// executeSerial executes steps one by one
func (o *Orchestrator) executeSerial(ctx context.Context, pipeline *model.Pipeline, run *model.Run) error {
	stepOutputs := make(map[string]json.RawMessage)

	for _, step := range pipeline.Steps {
		// Check if cancelled
		if run.Status == model.RunStatusCancelled {
			return fmt.Errorf("run cancelled")
		}

		// Check dependencies
		if !o.checkDependencies(step, stepOutputs) {
			continue
		}

		// Render params with previous outputs
		params := o.renderParams(step.Params, stepOutputs)

		// Execute step
		result := o.executeStep(ctx, step, params)
		run.StepResults = append(run.StepResults, result)
		o.redis.SaveRun(ctx, run)

		stepOutputs[step.ID] = result.Output

		if result.Status == model.RunStatusFailed {
			if step.OnError == model.ErrorStop {
				return fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			}
		}
	}

	return nil
}

// executeParallel executes all steps concurrently
func (o *Orchestrator) executeParallel(ctx context.Context, pipeline *model.Pipeline, run *model.Run) error {
	var wg sync.WaitGroup
	results := make(chan model.StepResult, len(pipeline.Steps))
	stepOutputs := sync.Map{}

	for _, step := range pipeline.Steps {
		wg.Add(1)
		go func(s model.Step) {
			defer wg.Done()

			result := o.executeStep(ctx, s, s.Params)
			results <- result
			stepOutputs.Store(s.ID, result.Output)
		}(step)
	}

	wg.Wait()
	close(results)

	for result := range results {
		run.StepResults = append(run.StepResults, result)
	}

	o.redis.SaveRun(ctx, run)
	return nil
}

// executeHybrid executes with dependency graph
func (o *Orchestrator) executeHybrid(ctx context.Context, pipeline *model.Pipeline, run *model.Run) error {
	stepOutputs := make(map[string]json.RawMessage)
	completed := make(map[string]bool)

	// Build dependency graph
	for len(completed) < len(pipeline.Steps) {
		for _, step := range pipeline.Steps {
			if completed[step.ID] {
				continue
			}

			// Check if all dependencies are completed
			ready := true
			for _, dep := range step.DependsOn {
				if !completed[dep] {
					ready = false
					break
				}
			}

			if !ready {
				continue
			}

			// Execute step
			params := o.renderParams(step.Params, stepOutputs)
			result := o.executeStep(ctx, step, params)
			run.StepResults = append(run.StepResults, result)
			o.redis.SaveRun(ctx, run)

			stepOutputs[step.ID] = result.Output
			completed[step.ID] = true

			if result.Status == model.RunStatusFailed && step.OnError == model.ErrorStop {
				return fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			}
		}
	}

	return nil
}

// executeStep executes a single step
func (o *Orchestrator) executeStep(ctx context.Context, step model.Step, params map[string]interface{}) model.StepResult {
	now := time.Now()
	result := model.StepResult{
		StepID:    step.ID,
		Status:    model.RunStatusRunning,
		StartedAt: &now,
	}

	// Execute CLI
	output, err := o.executor.ExecuteCLI(ctx, step.CLI, step.Action, step.Command, params)
	if err != nil {
		result.Status = model.RunStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = model.RunStatusCompleted
		result.Output = output
	}

	endTime := time.Now()
	result.EndedAt = &endTime
	if result.StartedAt != nil {
		result.Duration = endTime.Sub(*result.StartedAt).Milliseconds()
	}

	return result
}

// Helper methods

func (o *Orchestrator) validateSteps(pipeline *model.Pipeline) error {
	stepIDs := make(map[string]bool)

	for _, step := range pipeline.Steps {
		if step.ID == "" {
			return fmt.Errorf("step must have an ID")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true

		// Validate dependencies
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on non-existent step: %s", step.ID, dep)
			}
		}
	}

	return nil
}

func (o *Orchestrator) checkDependencies(step model.Step, outputs map[string]json.RawMessage) bool {
	for _, dep := range step.DependsOn {
		if _, ok := outputs[dep]; !ok {
			return false
		}
	}
	return true
}

func (o *Orchestrator) renderParams(params map[string]interface{}, outputs map[string]json.RawMessage) map[string]interface{} {
	// TODO: Implement template rendering with previous outputs
	return params
}

func (o *Orchestrator) mergeOutputs(ctx context.Context, pipeline *model.Pipeline, run *model.Run) {
	// Merge all step outputs into run output
	merged := make(map[string]interface{})
	for _, result := range run.StepResults {
		if len(result.Output) > 0 {
			var output map[string]interface{}
			if err := json.Unmarshal(result.Output, &output); err == nil {
				merged[result.StepID] = output
			}
		}
	}

	output, _ := json.Marshal(merged)
	run.Output = output
	o.redis.SaveRun(ctx, run)
}

func (o *Orchestrator) failRun(ctx context.Context, run *model.Run, errMsg string) {
	run.Status = model.RunStatusFailed
	run.Error = errMsg
	now := time.Now()
	run.CompletedAt = &now
	o.redis.SaveRun(ctx, run)
}

// SubscribeRunUpdates subscribes to run updates
func (o *Orchestrator) SubscribeRunUpdates(ctx context.Context, runID string) interface{} {
	return o.redis.SubscribeRunUpdates(ctx, runID)
}