package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/model"
)

// Agent Service Tests

func TestAgentModel(t *testing.T) {
	agent := &model.Agent{
		ID:           "test-agent",
		Name:         "Test Agent",
		Description:  "A test agent",
		Model:        "claude-sonnet-4-6",
		SystemPrompt: "You are a helpful assistant",
		MaxTokens:    4096,
		Enabled:      true,
	}

	if agent.ID != "test-agent" {
		t.Error("Agent ID mismatch")
	}

	if agent.Model != "claude-sonnet-4-6" {
		t.Error("Agent model mismatch")
	}

	if !agent.Enabled {
		t.Error("Agent should be enabled by default")
	}
}

func TestAgentWithSkills(t *testing.T) {
	agent := &model.Agent{
		ID:   "agent-with-skills",
		Name: "Multi-Skill Agent",
		Skills: []model.SkillRef{
			{
				SkillID: "skill-1",
				Alias:   "primary",
			},
			{
				SkillID: "skill-2",
				Alias:   "secondary",
			},
		},
		DefaultSkill: "primary",
	}

	if len(agent.Skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(agent.Skills))
	}

	if agent.DefaultSkill != "primary" {
		t.Error("Default skill should be 'primary'")
	}
}

func TestAgentWithTools(t *testing.T) {
	agent := &model.Agent{
		ID:   "agent-with-tools",
		Name: "Tooled Agent",
		Tools: []model.Tool{
			{Name: "read_file", Description: "Read file contents"},
			{Name: "write_file", Description: "Write file contents"},
		},
		Permissions: []model.Permission{
			{Resource: "file", Action: "read"},
			{Resource: "file", Action: "write"},
		},
	}

	if len(agent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
	}

	if len(agent.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(agent.Permissions))
	}
}

func TestAgentIsolation(t *testing.T) {
	agent := &model.Agent{
		ID:   "isolated-agent",
		Name: "Isolated Agent",
		Isolation: model.IsolationConfig{
			DataIsolation:    true,
			SessionIsolation: true,
			NetworkIsolation: false,
			FileIsolation:    true,
			Namespace:        "tenant-123",
		},
	}

	if !agent.Isolation.DataIsolation {
		t.Error("Data isolation should be enabled")
	}

	if !agent.Isolation.SessionIsolation {
		t.Error("Session isolation should be enabled")
	}

	if agent.Isolation.NetworkIsolation {
		t.Error("Network isolation should be disabled")
	}
}

func TestAgentCreateRequest(t *testing.T) {
	req := &model.AgentCreateRequest{
		Name:         "New Agent",
		Description:  "Created via API",
		Model:        "claude-opus-4-6",
		SystemPrompt: "System prompt",
		MaxTokens:    8192,
		Tags:         []string{"test", "api"},
		Category:     "analysis",
	}

	if req.Name != "New Agent" {
		t.Error("Name mismatch")
	}

	if req.Model != "claude-opus-4-6" {
		t.Error("Model mismatch")
	}

	if len(req.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(req.Tags))
	}
}

// Workflow Tests

func TestWorkflowModel(t *testing.T) {
	workflow := &model.Workflow{
		ID:          "test-workflow",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Mode:        model.ModeSerial,
		Enabled:     true,
	}

	if workflow.ID != "test-workflow" {
		t.Error("Workflow ID mismatch")
	}

	if workflow.Mode != model.ModeSerial {
		t.Error("Workflow mode should be serial")
	}
}

func TestWorkflowWithAgents(t *testing.T) {
	workflow := &model.Workflow{
		ID:   "multi-agent-workflow",
		Name: "Multi-Agent Workflow",
		Agents: []model.AgentNode{
			{
				ID:       "step-1",
				AgentID:  "agent-1",
				OutputAs: "result1",
			},
			{
				ID:        "step-2",
				AgentID:   "agent-2",
				InputFrom: map[string]string{"input": "result1"},
				DependsOn: []string{"step-1"},
				OutputAs:  "result2",
			},
		},
		Mode: model.ModeSerial,
	}

	if len(workflow.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(workflow.Agents))
	}

	if workflow.Agents[1].DependsOn[0] != "step-1" {
		t.Error("Dependency should be step-1")
	}
}

func TestWorkflowConnections(t *testing.T) {
	workflow := &model.Workflow{
		ID:   "connected-workflow",
		Name: "Connected Workflow",
		Connections: []model.Connection{
			{
				FromNode:   "step-1",
				FromOutput: "output",
				ToNode:     "step-2",
				ToInput:    "input",
			},
		},
	}

	if len(workflow.Connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(workflow.Connections))
	}
}

func TestExecutionModes(t *testing.T) {
	modes := []model.ExecutionMode{
		model.ModeSerial,
		model.ModeParallel,
		model.ModeHybrid,
	}

	for _, mode := range modes {
		workflow := &model.Workflow{
			ID:   string(mode) + "-workflow",
			Mode: mode,
		}

		if workflow.Mode != mode {
			t.Errorf("Mode mismatch: expected %s, got %s", mode, workflow.Mode)
		}
	}
}

// Execution Tests

func TestExecutionModel(t *testing.T) {
	exec := &model.Execution{
		ID:           "exec-123",
		WorkflowID:   "workflow-456",
		WorkflowName: "Test Workflow",
		Status:       model.ExecutionStatusPending,
		Progress:     0,
		TotalSteps:   3,
	}

	if exec.Status != model.ExecutionStatusPending {
		t.Error("Execution should be pending")
	}

	if exec.Progress != 0 {
		t.Error("Initial progress should be 0")
	}
}

func TestExecutionProgress(t *testing.T) {
	exec := &model.Execution{
		ID:              "exec-456",
		WorkflowID:      "workflow-789",
		Status:          model.ExecutionStatusRunning,
		Progress:        50,
		CurrentStep:     "step-2",
		TotalSteps:      4,
		CompletedSteps:  2,
	}

	if exec.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", exec.Progress)
	}

	if exec.CurrentStep != "step-2" {
		t.Error("Current step should be step-2")
	}
}

func TestExecutionStatuses(t *testing.T) {
	statuses := []model.ExecutionStatus{
		model.ExecutionStatusPending,
		model.ExecutionStatusRunning,
		model.ExecutionStatusCompleted,
		model.ExecutionStatusFailed,
		model.ExecutionStatusCancelled,
		model.ExecutionStatusPaused,
	}

	for _, status := range statuses {
		exec := &model.Execution{
			ID:     string(status) + "-exec",
			Status: status,
		}

		if exec.Status != status {
			t.Errorf("Status mismatch: expected %s", status)
		}
	}
}

func TestNodeResult(t *testing.T) {
	result := model.NodeResult{
		NodeID:    "node-1",
		AgentID:   "agent-1",
		Status:    model.ExecutionStatusCompleted,
		Duration:  1500,
		Retries:   0,
	}

	if result.NodeID != "node-1" {
		t.Error("Node ID mismatch")
	}

	if result.Status != model.ExecutionStatusCompleted {
		t.Error("Node should be completed")
	}
}

// Scheduled Job Tests

func TestScheduledJobModel(t *testing.T) {
	job := &model.ScheduledJob{
		ID:          "job-123",
		Name:        "Daily Scan",
		Description: "Run security scan daily",
		TargetType:  "workflow",
		TargetID:    "workflow-123",
		Cron:        "0 2 * * *",
		Timezone:    "Asia/Shanghai",
		Enabled:     true,
	}

	if job.TargetType != "workflow" {
		t.Error("Target type should be workflow")
	}

	if job.Cron != "0 2 * * *" {
		t.Error("Cron expression mismatch")
	}
}

func TestScheduledJobInput(t *testing.T) {
	job := &model.ScheduledJob{
		ID:   "job-with-input",
		Name: "Job with Input",
		Input: map[string]interface{}{
			"target": "src/",
			"depth":  "deep",
		},
	}

	if job.Input["target"] != "src/" {
		t.Error("Input target mismatch")
	}
}

func TestScheduledJobFailureHandling(t *testing.T) {
	testCases := []struct {
		onFailure string
	}{
		{"notify"},
		{"retry"},
		{"disable"},
	}

	for _, tc := range testCases {
		job := &model.ScheduledJob{
			ID:        "job-" + tc.onFailure,
			OnFailure: tc.onFailure,
		}

		if job.OnFailure != tc.onFailure {
			t.Errorf("OnFailure mismatch: expected %s", tc.onFailure)
		}
	}
}

// Agent Execute Request Tests

func TestAgentExecuteRequest(t *testing.T) {
	req := &model.AgentExecuteRequest{
		Input: map[string]interface{}{
			"code":  "package main",
			"file":  "main.go",
		},
		Context: map[string]interface{}{
			"work_dir": "/tmp/project",
		},
		Async:    true,
		Callback: "https://example.com/webhook",
	}

	if req.Input["code"] != "package main" {
		t.Error("Input code mismatch")
	}

	if !req.Async {
		t.Error("Should be async")
	}
}

// Workflow Create Request Tests

func TestWorkflowCreateRequest(t *testing.T) {
	req := &model.WorkflowCreateRequest{
		Name:        "New Workflow",
		Description: "Created via API",
		Agents: []model.AgentNode{
			{ID: "step-1", AgentID: "agent-1"},
		},
		Mode: model.ModeParallel,
	}

	if len(req.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(req.Agents))
	}

	if req.Mode != model.ModeParallel {
		t.Error("Mode should be parallel")
	}
}

// Execution Filter Tests

func TestExecutionFilter(t *testing.T) {
	filter := &model.ExecutionFilter{
		Status:     model.ExecutionStatusRunning,
		WorkflowID: "workflow-123",
		Page:       1,
		PageSize:   20,
	}

	if filter.Status != model.ExecutionStatusRunning {
		t.Error("Status filter mismatch")
	}

	if filter.PageSize != 20 {
		t.Error("Page size should be 20")
	}
}

// Context Test Helper

func TestContextHelper(t *testing.T) {
	ctx := context.Background()

	// Verify context is valid
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Verify context can be used
	select {
	case <-ctx.Done():
		t.Error("Context should not be done")
	default:
		// Expected
	}
}