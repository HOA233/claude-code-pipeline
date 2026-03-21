package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestEventBus_SubscribePublish(t *testing.T) {
	bus := service.NewEventBus()

	received := make(chan *service.Event, 1)

	bus.Subscribe("test.topic", func(ctx context.Context, event *service.Event) error {
		received <- event
		return nil
	})

	event := service.CreateEvent("test.event", "test", "subject", map[string]interface{}{
		"key": "value",
	})

	err := bus.Publish(context.Background(), "test.topic", event)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case e := <-received:
		if e.Type != "test.event" {
			t.Errorf("Expected type 'test.event', got '%s'", e.Type)
		}
	case <-context.Background().Done():
		t.Error("Timeout waiting for event")
	}
}

func TestEventBus_DeadLetterQueue(t *testing.T) {
	bus := service.NewEventBus()

	// Handler that always fails
	bus.Subscribe("fail.topic", func(ctx context.Context, event *service.Event) error {
		return nil
	})

	event := service.CreateEvent("test.event", "test", "subject", nil)
	event.RetryCount = 3 // Already at max retries

	bus.Publish(context.Background(), "fail.topic", event)

	dlq := bus.GetDeadLetterQueue()
	events := dlq.List()

	if len(events) != 1 {
		t.Errorf("Expected 1 event in DLQ, got %d", len(events))
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := service.NewEventBus()

	count := 0

	bus.Subscribe("multi.topic", func(ctx context.Context, event *service.Event) error {
		count++
		return nil
	})

	bus.Subscribe("multi.topic", func(ctx context.Context, event *service.Event) error {
		count++
		return nil
	})

	event := service.CreateEvent("test", "test", "test", nil)
	bus.Publish(context.Background(), "multi.topic", event)

	if count != 2 {
		t.Errorf("Expected 2 handler calls, got %d", count)
	}
}

func TestWorkflowEngine_RegisterWorkflow(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:      "test-workflow",
		Name:    "Test Workflow",
		Initial: "start",
		States: []service.StateDefinition{
			{ID: "start", Name: "Start", Type: "task"},
			{ID: "end", Name: "End", Type: "end"},
		},
	}

	err := engine.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}
}

func TestWorkflowEngine_RegisterWorkflow_NoInitial(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:   "no-initial",
		Name: "No Initial State",
		States: []service.StateDefinition{
			{ID: "start", Name: "Start", Type: "task"},
		},
	}

	err := engine.RegisterWorkflow(workflow)
	if err == nil {
		t.Error("Expected error for missing initial state")
	}
}

func TestWorkflowEngine_StartWorkflow(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:      "start-test",
		Name:    "Start Test",
		Initial: "end",
		States: []service.StateDefinition{
			{ID: "end", Name: "End", Type: "end"},
		},
	}

	engine.RegisterWorkflow(workflow)

	instance, err := engine.StartWorkflow(context.Background(), "start-test", map[string]interface{}{
		"input": "value",
	})

	if err != nil {
		t.Fatalf("Failed to start workflow: %v", err)
	}

	if instance.Status != "running" && instance.Status != "completed" {
		t.Errorf("Expected status 'running' or 'completed', got '%s'", instance.Status)
	}
}

func TestWorkflowEngine_GetInstance(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:      "get-test",
		Name:    "Get Test",
		Initial: "end",
		States: []service.StateDefinition{
			{ID: "end", Name: "End", Type: "end"},
		},
	}

	engine.RegisterWorkflow(workflow)
	instance, _ := engine.StartWorkflow(context.Background(), "get-test", nil)

	retrieved, err := engine.GetInstance(instance.ID)
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.ID != instance.ID {
		t.Error("Retrieved instance ID mismatch")
	}
}

func TestWorkflowEngine_CancelInstance(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:      "cancel-test",
		Name:    "Cancel Test",
		Initial: "wait",
		States: []service.StateDefinition{
			{ID: "wait", Name: "Wait", Type: "wait", WaitConfig: &service.WaitConfig{Type: "duration", Duration: 100}},
		},
	}

	engine.RegisterWorkflow(workflow)
	instance, _ := engine.StartWorkflow(context.Background(), "cancel-test", nil)

	err := engine.CancelInstance(instance.ID)
	if err != nil {
		t.Fatalf("Failed to cancel: %v", err)
	}

	retrieved, _ := engine.GetInstance(instance.ID)
	if retrieved.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", retrieved.Status)
	}
}

func TestWorkflowEngine_ListInstances(t *testing.T) {
	bus := service.NewEventBus()
	engine := service.NewWorkflowEngine(bus)

	workflow := &service.WorkflowDefinition{
		ID:      "list-test",
		Name:    "List Test",
		Initial: "end",
		States: []service.StateDefinition{
			{ID: "end", Name: "End", Type: "end"},
		},
	}

	engine.RegisterWorkflow(workflow)

	// Start multiple instances
	for i := 0; i < 3; i++ {
		engine.StartWorkflow(context.Background(), "list-test", nil)
	}

	instances := engine.ListInstances("list-test")
	if len(instances) < 3 {
		t.Errorf("Expected at least 3 instances, got %d", len(instances))
	}
}

func TestDeadLetterQueue_AddRemove(t *testing.T) {
	dlq := service.NewDeadLetterQueue(10)

	event1 := &service.Event{ID: "event-1"}
	event2 := &service.Event{ID: "event-2"}

	dlq.Add(event1)
	dlq.Add(event2)

	events := dlq.List()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	dlq.Remove("event-1")
	events = dlq.List()
	if len(events) != 1 {
		t.Errorf("Expected 1 event after removal, got %d", len(events))
	}
}

func TestDeadLetterQueue_MaxLength(t *testing.T) {
	dlq := service.NewDeadLetterQueue(3)

	for i := 0; i < 5; i++ {
		dlq.Add(&service.Event{ID: string(rune('0' + i))})
	}

	events := dlq.List()
	if len(events) > 3 {
		t.Errorf("Expected max 3 events, got %d", len(events))
	}
}

func TestEvent_ToJSON(t *testing.T) {
	event := service.CreateEvent("test.event", "source", "subject", map[string]interface{}{
		"key": "value",
	})

	data, err := event.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	parsed, err := service.EventFromJSON(data)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if parsed.Type != event.Type {
		t.Error("Type mismatch after serialization")
	}
}