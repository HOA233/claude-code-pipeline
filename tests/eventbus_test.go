package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// EventBus Service Tests

func TestEventBus_New(t *testing.T) {
	eb := service.NewEventBus()
	if eb == nil {
		t.Fatal("Expected non-nil event bus")
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	eb := service.NewEventBus()

	handler := func(event service.Event) {}
	unsubscribe := eb.Subscribe("test-topic", handler)

	if unsubscribe == nil {
		t.Error("Expected unsubscribe function")
	}
}

func TestEventBus_Publish(t *testing.T) {
	eb := service.NewEventBus()

	received := make(chan service.Event, 1)

	eb.Subscribe("test-topic", func(event service.Event) {
		received <- event
	})

	event := service.Event{
		Type:    "test-event",
		Payload: map[string]interface{}{"key": "value"},
	}

	eb.Publish("test-topic", event)

	select {
	case e := <-received:
		if e.Type != "test-event" {
			t.Error("Event type mismatch")
		}
	case <-time.After(time.Second):
		t.Error("Did not receive event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	eb := service.NewEventBus()

	count := 0
	handler := func(event service.Event) {
		count++
	}

	eb.Subscribe("multi-topic", handler)
	eb.Subscribe("multi-topic", handler)
	eb.Subscribe("multi-topic", handler)

	eb.Publish("multi-topic", service.Event{Type: "test"})
	time.Sleep(100 * time.Millisecond)

	if count != 3 {
		t.Errorf("Expected 3 handler calls, got %d", count)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := service.NewEventBus()

	count := 0
	handler := func(event service.Event) {
		count++
	}

	unsubscribe := eb.Subscribe("unsub-topic", handler)

	eb.Publish("unsub-topic", service.Event{Type: "test"})
	time.Sleep(100 * time.Millisecond)

	if count != 1 {
		t.Errorf("Expected 1 handler call, got %d", count)
	}

	// Unsubscribe
	unsubscribe()

	// Should not receive more events
	eb.Publish("unsub-topic", service.Event{Type: "test"})
	time.Sleep(100 * time.Millisecond)

	if count != 1 {
		t.Error("Expected no additional calls after unsubscribe")
	}
}

func TestEventBus_GetTopics(t *testing.T) {
	eb := service.NewEventBus()

	eb.Subscribe("topic-1", func(event service.Event) {})
	eb.Subscribe("topic-2", func(event service.Event) {})
	eb.Subscribe("topic-3", func(event service.Event) {})

	topics := eb.GetTopics()
	if len(topics) < 3 {
		t.Errorf("Expected at least 3 topics, got %d", len(topics))
	}
}

func TestEventBus_SubscriberCount(t *testing.T) {
	eb := service.NewEventBus()

	eb.Subscribe("count-topic", func(event service.Event) {})
	eb.Subscribe("count-topic", func(event service.Event) {})

	count := eb.SubscriberCount("count-topic")
	if count != 2 {
		t.Errorf("Expected 2 subscribers, got %d", count)
	}
}

func TestEventBus_ClearTopic(t *testing.T) {
	eb := service.NewEventBus()

	eb.Subscribe("clear-topic", func(event service.Event) {})

	eb.ClearTopic("clear-topic")

	count := eb.SubscriberCount("clear-topic")
	if count != 0 {
		t.Errorf("Expected 0 subscribers after clear, got %d", count)
	}
}

func TestEventBus_ClearAll(t *testing.T) {
	eb := service.NewEventBus()

	eb.Subscribe("topic-a", func(event service.Event) {})
	eb.Subscribe("topic-b", func(event service.Event) {})

	eb.ClearAll()

	topics := eb.GetTopics()
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics after clear, got %d", len(topics))
	}
}

func TestEventBus_AsyncPublish(t *testing.T) {
	eb := service.NewEventBus()

	received := make(chan service.Event, 1)

	eb.Subscribe("async-topic", func(event service.Event) {
		time.Sleep(50 * time.Millisecond)
		received <- event
	})

	eb.PublishAsync("async-topic", service.Event{Type: "async-test"})

	select {
	case <-received:
		// Success
	case <-time.After(time.Second):
		t.Error("Did not receive async event")
	}
}

func TestEventBus_EventToJSON(t *testing.T) {
	event := service.Event{
		ID:        "event-1",
		Type:      "test",
		Topic:     "test-topic",
		Payload:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}

	data, err := event.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}