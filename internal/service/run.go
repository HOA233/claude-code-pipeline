package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/model"
)

// RunService manages pipeline runs
type RunService struct {
	mu          sync.RWMutex
	runs        map[string]*model.Run
	pipelines   map[string]*model.Pipeline
	sessions    map[string]*model.PipelineSession
	executor    *CLIExecutor
	eventBus    *EventBus
	circuit     *CircuitBreaker
}

// NewRunService creates a new run service
func NewRunService(executor *CLIExecutor, eventBus *EventBus) *RunService {
	return &RunService{
		runs:      make(map[string]*model.Run),
		pipelines: make(map[string]*model.Pipeline),
		sessions:  make(map[string]*model.PipelineSession),
		executor:  executor,
		eventBus:  eventBus,
		circuit: NewCircuitBreaker(&CircuitBreakerConfig{
			Name:             "run-service",
			FailureThreshold: 5,
			Timeout:          30 * time.Second,
		}),
	}
}

// RegisterPipeline registers a pipeline for execution
func (s *RunService) RegisterPipeline(pipeline *model.Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pipeline.ID == "" {
		return errors.New("pipeline ID is required")
	}

	if len(pipeline.Steps) == 0 {
		return errors.New("pipeline must have at least one step")
	}

	// Validate step dependencies
	if err := s.validateDependencies(pipeline); err != nil {
		return err
	}

	pipeline.CreatedAt = time.Now()
	pipeline.UpdatedAt = time.Now()
	s.pipelines[pipeline.ID] = pipeline

	return nil
}

// validateDependencies validates step dependencies
func (s *RunService) validateDependencies(pipeline *model.Pipeline) error {
	stepIDs := make(map[string]bool)
	for _, step := range pipeline.Steps {
		stepIDs[step.ID] = true
	}

	for _, step := range pipeline.Steps {
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, step := range pipeline.Steps {
		if s.hasCycle(step.ID, pipeline.Steps, visited, recStack) {
			return fmt.Errorf("circular dependency detected")
		}
	}

	return nil
}

// hasCycle detects circular dependencies using DFS
func (s *RunService) hasCycle(stepID string, steps []model.Step, visited, recStack map[string]bool) bool {
	visited[stepID] = true
	recStack[stepID] = true

	// Find step and check dependencies
	for _, step := range steps {
		if step.ID == stepID {
			for _, dep := range step.DependsOn {
				if !visited[dep] {
					if s.hasCycle(dep, steps, visited, recStack) {
						return true
					}
				} else if recStack[dep] {
					return true
				}
			}
		}
	}

	recStack[stepID] = false
	return false
}

// CreateRun creates a new pipeline run
func (s *RunService) CreateRun(ctx context.Context, req *model.RunCreateRequest) (*model.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pipeline, exists := s.pipelines[req.PipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", req.PipelineID)
	}

	// Create session
	session := &model.PipelineSession{
		ID:            generateID(),
		PipelineID:    req.PipelineID,
		TenantID:      pipeline.TenantID,
		Data:          make(map[string]interface{}),
		SkillRegistry: pipeline.Skills,
		StepHistory:   []model.StepResult{},
		SharedData:    make(map[string]json.RawMessage),
		CreatedAt:     time.Now(),
	}

	// Copy context to session data
	for k, v := range req.Context {
		session.Data[k] = v
	}
	for k, v := range req.Params {
		session.Data[k] = v
	}

	s.sessions[session.ID] = session

	// Create run
	now := time.Now()
	run := &model.Run{
		ID:          generateID(),
		PipelineID:  req.PipelineID,
		SessionID:   session.ID,
		TenantID:    pipeline.TenantID,
		Status:      model.RunStatusPending,
		StepResults: []model.StepResult{},
		CreatedAt:   now,
	}

	s.runs[run.ID] = run

	// Publish event
	s.eventBus.Publish(ctx, "run.created", CreateEvent(
		"run.created",
		"run-service",
		run.ID,
		map[string]interface{}{
			"pipeline_id": req.PipelineID,
		},
	))

	return run, nil
}

// StartRun starts a pipeline run
func (s *RunService) StartRun(ctx context.Context, runID string) error {
	s.mu.Lock()
	run, exists := s.runs[runID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("run not found: %s", runID)
	}

	if run.Status != model.RunStatusPending {
		s.mu.Unlock()
		return fmt.Errorf("run is not in pending state")
	}

	pipeline := s.pipelines[run.PipelineID]
	session := s.sessions[run.SessionID]
	s.mu.Unlock()

	// Update status
	now := time.Now()
	run.Status = model.RunStatusRunning
	run.StartedAt = &now

	// Publish event
	s.eventBus.Publish(ctx, "run.started", CreateEvent(
		"run.started",
		"run-service",
		run.ID,
		map[string]interface{}{
			"pipeline_id": pipeline.ID,
			"mode":        pipeline.Mode,
		},
	))

	// Execute based on mode
	var err error
	switch pipeline.Mode {
	case model.ModeSerial:
		err = s.executeSerial(ctx, run, pipeline, session)
	case model.ModeParallel:
		err = s.executeParallel(ctx, run, pipeline, session)
	case model.ModeHybrid:
		err = s.executeHybrid(ctx, run, pipeline, session)
	default:
		err = s.executeSerial(ctx, run, pipeline, session)
	}

	// Update final status
	completedAt := time.Now()
	run.CompletedAt = &completedAt
	run.Duration = completedAt.Sub(now).Milliseconds()

	if err != nil {
		run.Status = model.RunStatusFailed
		run.Error = err.Error()
	} else {
		run.Status = model.RunStatusCompleted
	}

	// Publish completion event
	s.eventBus.Publish(ctx, "run.completed", CreateEvent(
		"run.completed",
		"run-service",
		run.ID,
		map[string]interface{}{
			"status":   run.Status,
			"duration": run.Duration,
		},
	))

	return nil
}

// executeSerial executes steps one by one
func (s *RunService) executeSerial(ctx context.Context, run *model.Run, pipeline *model.Pipeline, session *model.PipelineSession) error {
	for _, step := range pipeline.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result := s.executeStep(ctx, step, session)
		run.StepResults = append(run.StepResults, result)
		session.StepHistory = append(session.StepHistory, result)

		// Store output in session
		if step.OutputTo != "" && len(result.Output) > 0 {
			session.SharedData[step.OutputTo] = result.Output
		}

		if result.Status == model.RunStatusFailed {
			if step.OnError == model.ErrorStop {
				return fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			}
		}

		// Publish step event
		s.eventBus.Publish(ctx, "run.step", CreateEvent(
			"run.step",
			"run-service",
			run.ID,
			map[string]interface{}{
				"step_id": step.ID,
				"status":  result.Status,
			},
		))
	}

	return nil
}

// executeParallel executes all steps concurrently
func (s *RunService) executeParallel(ctx context.Context, run *model.Run, pipeline *model.Pipeline, session *model.PipelineSession) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(pipeline.Steps))
	results := make([]model.StepResult, len(pipeline.Steps))
	var resultsMu sync.Mutex

	for i, step := range pipeline.Steps {
		wg.Add(1)
		go func(idx int, st model.Step) {
			defer wg.Done()

			result := s.executeStep(ctx, st, session)

			resultsMu.Lock()
			results[idx] = result
			resultsMu.Unlock()

			if result.Status == model.RunStatusFailed && st.OnError == model.ErrorStop {
				errChan <- fmt.Errorf("step %s failed: %s", st.ID, result.Error)
			}
		}(i, step)
	}

	wg.Wait()
	close(errChan)

	run.StepResults = results

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeHybrid executes steps based on dependencies
func (s *RunService) executeHybrid(ctx context.Context, run *model.Run, pipeline *model.Pipeline, session *model.PipelineSession) error {
	// Build dependency graph
	dependents := make(map[string][]string)
	dependencies := make(map[string]int)
	stepMap := make(map[string]model.Step)

	for _, step := range pipeline.Steps {
		stepMap[step.ID] = step
		dependencies[step.ID] = len(step.DependsOn)
		for _, dep := range step.DependsOn {
			dependents[dep] = append(dependents[dep], step.ID)
		}
	}

	// Execute using topological sort with parallel execution
	completed := make(map[string]bool)
	var mu sync.Mutex

	for len(completed) < len(pipeline.Steps) {
		// Find ready steps (all dependencies completed)
		var ready []model.Step
		for _, step := range pipeline.Steps {
			if completed[step.ID] {
				continue
			}
			if dependencies[step.ID] == 0 {
				ready = append(ready, step)
			}
		}

		if len(ready) == 0 && len(completed) < len(pipeline.Steps) {
			return errors.New("deadlock detected in step dependencies")
		}

		// Execute ready steps in parallel
		var wg sync.WaitGroup
		for _, step := range ready {
			wg.Add(1)
			go func(st model.Step) {
				defer wg.Done()

				result := s.executeStep(ctx, st, session)
				run.StepResults = append(run.StepResults, result)

				mu.Lock()
				completed[st.ID] = true
				// Update dependency counts
				for _, dep := range dependents[st.ID] {
					dependencies[dep]--
				}
				mu.Unlock()

				// Publish event
				s.eventBus.Publish(ctx, "run.step", CreateEvent(
					"run.step",
					"run-service",
					run.ID,
					map[string]interface{}{
						"step_id": st.ID,
						"status":  result.Status,
					},
				))
			}(step)
		}
		wg.Wait()
	}

	return nil
}

// executeStep executes a single step
func (s *RunService) executeStep(ctx context.Context, step model.Step, session *model.PipelineSession) model.StepResult {
	now := time.Now()
	result := model.StepResult{
		StepID:    step.ID,
		Status:    model.RunStatusRunning,
		StartedAt: &now,
	}

	// Check condition if present
	if step.Condition != "" && !s.evaluateCondition(step.Condition, session.Data) {
		result.Status = model.RunStatusCompleted
		result.Output = json.RawMessage(`{"skipped": true, "reason": "condition not met"}`)
		endedAt := time.Now()
		result.EndedAt = &endedAt
		return result
	}

	// Execute with retries
	var err error
	for i := 0; i <= step.RetryCount; i++ {
		err = s.circuit.Execute(ctx, func() error {
			// Simulate CLI execution
			output := map[string]interface{}{
				"cli":     step.CLI,
				"action":  step.Action,
				"command": step.Command,
				"params":  step.Params,
			}
			data, _ := json.Marshal(output)
			result.Output = data
			return nil
		})

		if err == nil {
			break
		}

		result.Retries = i + 1
		time.Sleep(time.Second * time.Duration(i+1))
	}

	endedAt := time.Now()
	result.EndedAt = &endedAt
	result.Duration = endedAt.Sub(now).Milliseconds()

	if err != nil {
		result.Status = model.RunStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = model.RunStatusCompleted
	}

	return result
}

// evaluateCondition evaluates a step condition
func (s *RunService) evaluateCondition(condition string, data map[string]interface{}) bool {
	// Simplified condition evaluation
	// In production, use an expression evaluator
	return condition == "" || condition == "true"
}

// GetRun retrieves a run by ID
func (s *RunService) GetRun(runID string) (*model.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, exists := s.runs[runID]
	if !exists {
		return nil, fmt.Errorf("run not found: %s", runID)
	}
	return run, nil
}

// CancelRun cancels a running pipeline
func (s *RunService) CancelRun(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, exists := s.runs[runID]
	if !exists {
		return fmt.Errorf("run not found: %s", runID)
	}

	if run.Status != model.RunStatusRunning && run.Status != model.RunStatusPending {
		return fmt.Errorf("cannot cancel run in status: %s", run.Status)
	}

	run.Status = model.RunStatusCancelled
	now := time.Now()
	run.CompletedAt = &now

	return nil
}

// ListRuns lists all runs for a pipeline
func (s *RunService) ListRuns(pipelineID string) []*model.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := make([]*model.Run, 0)
	for _, run := range s.runs {
		if pipelineID == "" || run.PipelineID == pipelineID {
			runs = append(runs, run)
		}
	}
	return runs
}

// GetPipeline retrieves a pipeline by ID
func (s *RunService) GetPipeline(pipelineID string) (*model.Pipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, exists := s.pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}
	return pipeline, nil
}

// ListPipelines lists all pipelines
func (s *RunService) ListPipelines() []*model.Pipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipelines := make([]*model.Pipeline, 0, len(s.pipelines))
	for _, p := range s.pipelines {
		pipelines = append(pipelines, p)
	}
	return pipelines
}

// DeletePipeline deletes a pipeline
func (s *RunService) DeletePipeline(pipelineID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pipelines, pipelineID)
	return nil
}

// GetSession retrieves a session by ID
func (s *RunService) GetSession(sessionID string) (*model.PipelineSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}