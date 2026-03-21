package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// EventBus provides pub/sub event handling
type EventBus struct {
	mu        sync.RWMutex
	handlers  map[string][]EventHandler
	deadletter *DeadLetterQueue
	emitter   EventEmitter
}

// EventHandler handles events
type EventHandler func(ctx context.Context, event *Event) error

// Event represents an event
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Subject     string                 `json:"subject"`
	Time        time.Time              `json:"time"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata"`
	RetryCount  int                    `json:"retry_count"`
	SpecVersion string                 `json:"specversion"`
}

// EventSubscription represents a subscription
type EventSubscription struct {
	ID        string   `json:"id"`
	Topic     string   `json:"topic"`
	Handler   string   `json:"handler"`
	Filter    string   `json:"filter"`
	Enabled   bool     `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// EventEmitter emits events to external systems
type EventEmitter interface {
	Emit(ctx context.Context, event *Event) error
}

// DeadLetterQueue stores failed events
type DeadLetterQueue struct {
	mu     sync.Mutex
	events []*Event
	maxLen int
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(maxLen int) *DeadLetterQueue {
	return &DeadLetterQueue{
		events: make([]*Event, 0),
		maxLen: maxLen,
	}
}

// Add adds an event to the queue
func (q *DeadLetterQueue) Add(event *Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.events) >= q.maxLen {
		q.events = q.events[1:]
	}
	q.events = append(q.events, event)
}

// List returns all events in the queue
func (q *DeadLetterQueue) List() []*Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]*Event, len(q.events))
	copy(result, q.events)
	return result
}

// Remove removes an event from the queue
func (q *DeadLetterQueue) Remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, e := range q.events {
		if e.ID == id {
			q.events = append(q.events[:i], q.events[i+1:]...)
			break
		}
	}
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers:  make(map[string][]EventHandler),
		deadletter: NewDeadLetterQueue(1000),
	}
}

// Subscribe subscribes to a topic
func (b *EventBus) Subscribe(topic string, handler EventHandler) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	subID := generateID()
	b.handlers[topic] = append(b.handlers[topic], handler)

	return subID
}

// Unsubscribe removes a subscription
func (b *EventBus) Unsubscribe(topic, subID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// In real implementation, would track subscription IDs
}

// Publish publishes an event to a topic
func (b *EventBus) Publish(ctx context.Context, topic string, event *Event) error {
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers[topic]))
	copy(handlers, b.handlers[topic])
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	event.Time = time.Now()
	event.SpecVersion = "1.0"

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			event.RetryCount++
			if event.RetryCount >= 3 {
				b.deadletter.Add(event)
			}
		}
	}

	return nil
}

// Emit emits an event to external systems
func (b *EventBus) Emit(ctx context.Context, event *Event) error {
	if b.emitter != nil {
		return b.emitter.Emit(ctx, event)
	}
	return nil
}

// GetDeadLetterQueue returns the dead letter queue
func (b *EventBus) GetDeadLetterQueue() *DeadLetterQueue {
	return b.deadletter
}

// Event types
const (
	EventTaskCreated   = "task.created"
	EventTaskStarted   = "task.started"
	EventTaskProgress  = "task.progress"
	EventTaskCompleted = "task.completed"
	EventTaskFailed    = "task.failed"

	EventRunCreated   = "run.created"
	EventRunStarted   = "run.started"
	EventRunStep      = "run.step"
	EventRunCompleted = "run.completed"
	EventRunFailed    = "run.failed"

	EventPipelineCreated = "pipeline.created"
	EventPipelineUpdated = "pipeline.updated"
	EventPipelineDeleted = "pipeline.deleted"

	EventScheduleTriggered = "schedule.triggered"
)

// CreateEvent creates a new event
func CreateEvent(eventType, source, subject string, data map[string]interface{}) *Event {
	return &Event{
		ID:       generateID(),
		Type:     eventType,
		Source:   source,
		Subject:  subject,
		Time:     time.Now(),
		Data:     data,
		Metadata: make(map[string]string),
	}
}

// ToJSON serializes the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an event from JSON
func EventFromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}