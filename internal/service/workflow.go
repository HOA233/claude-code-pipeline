package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// WorkflowEngine manages workflow executions
type WorkflowEngine struct {
	mu          sync.RWMutex
	workflows   map[string]*WorkflowDefinition
	instances   map[string]*WorkflowInstance
	executor    *WorkflowExecutor
	eventBus    *EventBus
}

// WorkflowDefinition defines a workflow
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	States      []StateDefinition       `json:"states"`
	Transitions []TransitionDefinition `json:"transitions"`
	Initial     string                 `json:"initial"`
	Timeout     int                    `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StateDefinition defines a workflow state
type StateDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // task, parallel, choice, wait, end
	Action      *ActionDefinition      `json:"action,omitempty"`
	Branches    []BranchDefinition     `json:"branches,omitempty"`
	Choices     []ChoiceDefinition     `json:"choices,omitempty"`
	WaitConfig  *WaitConfig            `json:"wait_config,omitempty"`
	OnError     string                 `json:"on_error,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	Retry       *RetryDefinition       `json:"retry,omitempty"`
}

// ActionDefinition defines an action
type ActionDefinition struct {
	Type       string                 `json:"type"` // task, skill, pipeline
	SkillID    string                 `json:"skill_id,omitempty"`
	PipelineID string                 `json:"pipeline_id,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// BranchDefinition defines a parallel branch
type BranchDefinition struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Steps []StateDefinition `json:"steps"`
}

// ChoiceDefinition defines a conditional branch
type ChoiceDefinition struct {
	Condition string `json:"condition"`
	Next      string `json:"next"`
}

// WaitConfig defines wait behavior
type WaitConfig struct {
	Type     string `json:"type"` // event, duration, callback
	Event    string `json:"event,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

// RetryDefinition defines retry behavior
type RetryDefinition struct {
	MaxAttempts int    `json:"max_attempts"`
	Backoff     string `json:"backoff"`
	Interval    int    `json:"interval"`
}

// TransitionDefinition defines state transitions
type TransitionDefinition struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Event  string `json:"event,omitempty"`
	Guard  string `json:"guard,omitempty"`
}

// WorkflowInstance represents a running workflow
type WorkflowInstance struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	Status       string                 `json:"status"` // pending, running, completed, failed, cancelled
	CurrentState string                 `json:"current_state"`
	Variables    map[string]interface{} `json:"variables"`
	History      []StateExecution       `json:"history"`
	Error        string                 `json:"error,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// StateExecution records a state execution
type StateExecution struct {
	StateID    string                 `json:"state_id"`
	Status     string                 `json:"status"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
}

// WorkflowExecutor executes workflows
type WorkflowExecutor struct {
	taskService *TaskService
	eventBus    *EventBus
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(eventBus *EventBus) *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]*WorkflowDefinition),
		instances: make(map[string]*WorkflowInstance),
		executor:  &WorkflowExecutor{eventBus: eventBus},
		eventBus:  eventBus,
	}
}

// RegisterWorkflow registers a workflow definition
func (e *WorkflowEngine) RegisterWorkflow(workflow *WorkflowDefinition) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if workflow.ID == "" {
		return errors.New("workflow ID is required")
	}

	if workflow.Initial == "" {
		return errors.New("initial state is required")
	}

	// Validate states exist
	stateIDs := make(map[string]bool)
	for _, state := range workflow.States {
		stateIDs[state.ID] = true
	}

	if !stateIDs[workflow.Initial] {
		return fmt.Errorf("initial state '%s' not found", workflow.Initial)
	}

	e.workflows[workflow.ID] = workflow
	return nil
}

// StartWorkflow starts a new workflow instance
func (e *WorkflowEngine) StartWorkflow(ctx context.Context, workflowID string, variables map[string]interface{}) (*WorkflowInstance, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	instance := &WorkflowInstance{
		ID:           generateID(),
		WorkflowID:   workflowID,
		Status:       "running",
		CurrentState: workflow.Initial,
		Variables:    variables,
		History:      []StateExecution{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	e.instances[instance.ID] = instance

	// Emit event
	e.eventBus.Publish(ctx, "workflow.started", CreateEvent(
		"workflow.started",
		"workflow-engine",
		instance.ID,
		map[string]interface{}{
			"workflow_id": workflowID,
		},
	))

	// Execute initial state
	go e.executeState(context.Background(), instance, workflow)

	return instance, nil
}

// executeState executes a state
func (e *WorkflowEngine) executeState(ctx context.Context, instance *WorkflowInstance, workflow *WorkflowDefinition) {
	var currentState *StateDefinition
	for _, state := range workflow.States {
		if state.ID == instance.CurrentState {
			currentState = &state
			break
		}
	}

	if currentState == nil {
		instance.Status = "failed"
		instance.Error = fmt.Sprintf("state not found: %s", instance.CurrentState)
		return
	}

	// Record execution start
	execution := StateExecution{
		StateID:   currentState.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}

	// Execute based on state type
	switch currentState.Type {
	case "task":
		err := e.executor.executeTaskState(ctx, instance, currentState)
		if err != nil {
			execution.Status = "failed"
			execution.Error = err.Error()
			instance.Status = "failed"
			instance.Error = err.Error()
		} else {
			execution.Status = "completed"
		}

	case "parallel":
		err := e.executor.executeParallelState(ctx, instance, currentState)
		if err != nil {
			execution.Status = "failed"
			execution.Error = err.Error()
		} else {
			execution.Status = "completed"
		}

	case "choice":
		nextState := e.executor.evaluateChoice(instance, currentState)
		instance.CurrentState = nextState
		execution.Status = "completed"

	case "wait":
		err := e.executor.executeWaitState(ctx, instance, currentState)
		if err != nil {
			execution.Status = "failed"
			execution.Error = err.Error()
		} else {
			execution.Status = "completed"
		}

	case "end":
		instance.Status = "completed"
		now := time.Now()
		instance.CompletedAt = &now
		execution.Status = "completed"

	default:
		execution.Status = "failed"
		execution.Error = fmt.Sprintf("unknown state type: %s", currentState.Type)
	}

	now := time.Now()
	execution.FinishedAt = &now
	instance.History = append(instance.History, execution)
	instance.UpdatedAt = now

	// Find next state
	if instance.Status == "running" {
		for _, transition := range workflow.Transitions {
			if transition.From == currentState.ID {
				instance.CurrentState = transition.To
				go e.executeState(ctx, instance, workflow)
				return
			}
		}
		// No transition found, check if this is an end state
		if currentState.Type != "end" {
			instance.Status = "completed"
			instance.CompletedAt = &now
		}
	}
}

// GetInstance gets a workflow instance
func (e *WorkflowEngine) GetInstance(instanceID string) (*WorkflowInstance, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	instance, exists := e.instances[instanceID]
	if !exists {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}
	return instance, nil
}

// CancelInstance cancels a workflow instance
func (e *WorkflowEngine) CancelInstance(instanceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	instance, exists := e.instances[instanceID]
	if !exists {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	if instance.Status != "running" && instance.Status != "pending" {
		return fmt.Errorf("cannot cancel instance in status: %s", instance.Status)
	}

	instance.Status = "cancelled"
	now := time.Now()
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	return nil
}

// ListInstances lists all workflow instances
func (e *WorkflowEngine) ListInstances(workflowID string) []*WorkflowInstance {
	e.mu.RLock()
	defer e.mu.RUnlock()

	instances := make([]*WorkflowInstance, 0)
	for _, instance := range e.instances {
		if workflowID == "" || instance.WorkflowID == workflowID {
			instances = append(instances, instance)
		}
	}
	return instances
}

// executeTaskState executes a task state
func (ex *WorkflowExecutor) executeTaskState(ctx context.Context, instance *WorkflowInstance, state *StateDefinition) error {
	if state.Action == nil {
		return errors.New("action is required for task state")
	}

	// Execute task based on action type
	switch state.Action.Type {
	case "skill":
		// Would call task service to execute skill
		return nil
	case "pipeline":
		// Would execute pipeline
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", state.Action.Type)
	}
}

// executeParallelState executes a parallel state
func (ex *WorkflowExecutor) executeParallelState(ctx context.Context, instance *WorkflowInstance, state *StateDefinition) error {
	if len(state.Branches) == 0 {
		return errors.New("parallel state requires branches")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(state.Branches))

	for _, branch := range state.Branches {
		wg.Add(1)
		go func(b BranchDefinition) {
			defer wg.Done()
			// Execute branch steps
		}(branch)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// evaluateChoice evaluates a choice state
func (ex *WorkflowExecutor) evaluateChoice(instance *WorkflowInstance, state *StateDefinition) string {
	for _, choice := range state.Choices {
		// Evaluate condition
		if evaluateCondition(choice.Condition, instance.Variables) {
			return choice.Next
		}
	}
	return ""
}

// executeWaitState executes a wait state
func (ex *WorkflowExecutor) executeWaitState(ctx context.Context, instance *WorkflowInstance, state *StateDefinition) error {
	if state.WaitConfig == nil {
		return errors.New("wait_config is required for wait state")
	}

	switch state.WaitConfig.Type {
	case "duration":
		time.Sleep(time.Duration(state.WaitConfig.Duration) * time.Second)
	case "event":
		// Would wait for event
	case "callback":
		// Would wait for callback
	}

	return nil
}

// evaluateCondition evaluates a condition expression
func evaluateCondition(condition string, variables map[string]interface{}) bool {
	// Simplified condition evaluation
	// In real implementation, would use expression language
	return condition == "" || condition == "true"
}

// ToJSON serializes workflow definition
func (w *WorkflowDefinition) ToJSON() ([]byte, error) {
	return json.MarshalIndent(w, "", "  ")
}

// ParseWorkflow parses workflow from JSON
func ParseWorkflow(data []byte) (*WorkflowDefinition, error) {
	var workflow WorkflowDefinition
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}
	return &workflow, nil
}