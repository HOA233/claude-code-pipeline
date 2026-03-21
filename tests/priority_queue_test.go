package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// PriorityQueue Service Tests

func TestPriorityQueue_New(t *testing.T) {
	pq := service.NewPriorityQueue()
	if pq == nil {
		t.Fatal("Expected non-nil priority queue")
	}
}

func TestPriorityQueue_Enqueue(t *testing.T) {
	pq := service.NewPriorityQueue()

	item := &service.QueueItem{
		ID:       "item-1",
		Data:     "test-data",
		Priority: 1,
	}

	err := pq.Enqueue(item)
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	if item.EnqueuedAt.IsZero() {
		t.Error("Expected EnqueuedAt to be set")
	}
}

func TestPriorityQueue_Enqueue_MissingID(t *testing.T) {
	pq := service.NewPriorityQueue()

	item := &service.QueueItem{
		Data:     "test",
		Priority: 1,
	}

	err := pq.Enqueue(item)
	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestPriorityQueue_Dequeue(t *testing.T) {
	pq := service.NewPriorityQueue()

	// Enqueue items with different priorities
	pq.Enqueue(&service.QueueItem{ID: "low", Data: "low-priority", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "high", Data: "high-priority", Priority: 10})
	pq.Enqueue(&service.QueueItem{ID: "medium", Data: "medium-priority", Priority: 5})

	// Should dequeue highest priority first
	item := pq.Dequeue()
	if item.ID != "high" {
		t.Errorf("Expected high priority item, got %s", item.ID)
	}
}

func TestPriorityQueue_Dequeue_Empty(t *testing.T) {
	pq := service.NewPriorityQueue()

	item := pq.Dequeue()
	if item != nil {
		t.Error("Expected nil for empty queue")
	}
}

func TestPriorityQueue_Peek(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "peek-1", Data: "test", Priority: 5})

	item := pq.Peek()
	if item == nil {
		t.Fatal("Expected item from peek")
	}

	if item.ID != "peek-1" {
		t.Error("Item ID mismatch")
	}

	// Peek should not remove the item
	if pq.Size() != 1 {
		t.Error("Peek should not remove item from queue")
	}
}

func TestPriorityQueue_Peek_Empty(t *testing.T) {
	pq := service.NewPriorityQueue()

	item := pq.Peek()
	if item != nil {
		t.Error("Expected nil for empty queue peek")
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "remove-1", Data: "test", Priority: 1})

	err := pq.Remove("remove-1")
	if err != nil {
		t.Fatalf("Failed to remove: %v", err)
	}

	if pq.Size() != 0 {
		t.Error("Queue should be empty after remove")
	}
}

func TestPriorityQueue_Remove_NotFound(t *testing.T) {
	pq := service.NewPriorityQueue()

	err := pq.Remove("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent item")
	}
}

func TestPriorityQueue_UpdatePriority(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "update-1", Data: "test", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "update-2", Data: "test", Priority: 5})

	err := pq.UpdatePriority("update-1", 10)
	if err != nil {
		t.Fatalf("Failed to update priority: %v", err)
	}

	// update-1 should now be first
	item := pq.Dequeue()
	if item.ID != "update-1" {
		t.Error("Expected update-1 to be highest priority after update")
	}
}

func TestPriorityQueue_Size(t *testing.T) {
	pq := service.NewPriorityQueue()

	if pq.Size() != 0 {
		t.Error("Expected empty queue to have size 0")
	}

	pq.Enqueue(&service.QueueItem{ID: "size-1", Data: "test", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "size-2", Data: "test", Priority: 1})

	if pq.Size() != 2 {
		t.Errorf("Expected size 2, got %d", pq.Size())
	}
}

func TestPriorityQueue_Clear(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "clear-1", Data: "test", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "clear-2", Data: "test", Priority: 1})

	pq.Clear()

	if pq.Size() != 0 {
		t.Error("Queue should be empty after clear")
	}
}

func TestPriorityQueue_Get(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "get-1", Data: "test-data", Priority: 1})

	item := pq.Get("get-1")
	if item == nil {
		t.Fatal("Expected item")
	}

	if item.Data != "test-data" {
		t.Error("Item data mismatch")
	}
}

func TestPriorityQueue_Get_NotFound(t *testing.T) {
	pq := service.NewPriorityQueue()

	item := pq.Get("nonexistent")
	if item != nil {
		t.Error("Expected nil for nonexistent item")
	}
}

func TestPriorityQueue_List(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "list-1", Data: "test", Priority: 5})
	pq.Enqueue(&service.QueueItem{ID: "list-2", Data: "test", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "list-3", Data: "test", Priority: 10})

	items := pq.List()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Should be sorted by priority (descending)
	if items[0].ID != "list-3" || items[2].ID != "list-2" {
		t.Error("Items should be sorted by priority")
	}
}

func TestPriorityQueue_DequeueBatch(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "batch-1", Data: "test", Priority: 5})
	pq.Enqueue(&service.QueueItem{ID: "batch-2", Data: "test", Priority: 3})
	pq.Enqueue(&service.QueueItem{ID: "batch-3", Data: "test", Priority: 1})

	items := pq.DequeueBatch(2)
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Should get highest priority items
	if items[0].ID != "batch-1" || items[1].ID != "batch-2" {
		t.Error("Should get highest priority items first")
	}
}

func TestPriorityQueue_GetStats(t *testing.T) {
	pq := service.NewPriorityQueue()

	pq.Enqueue(&service.QueueItem{ID: "stats-1", Data: "test", Priority: 1})
	pq.Enqueue(&service.QueueItem{ID: "stats-2", Data: "test", Priority: 5})
	pq.Enqueue(&service.QueueItem{ID: "stats-3", Data: "test", Priority: 10})

	stats := pq.GetStats()

	if stats.TotalItems != 3 {
		t.Errorf("Expected 3 items, got %d", stats.TotalItems)
	}
}

func TestPriorityQueue_QueueItemToJSON(t *testing.T) {
	item := &service.QueueItem{
		ID:         "json-1",
		Data:       "test-data",
		Priority:   5,
		EnqueuedAt: time.Now(),
	}

	data, err := item.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}