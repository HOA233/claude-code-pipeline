package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/api"
)

func TestSSEHandler_New(t *testing.T) {
	handler := api.NewSSEHandler()
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

func TestSSEHandler_GetClientCount(t *testing.T) {
	handler := api.NewSSEHandler()

	count := handler.GetClientCount("nonexistent")
	if count != 0 {
		t.Errorf("Expected 0 clients for nonexistent channel, got %d", count)
	}
}

func TestSSEHandler_GetTotalClients(t *testing.T) {
	handler := api.NewSSEHandler()

	total := handler.GetTotalClients()
	if total != 0 {
		t.Errorf("Expected 0 total clients, got %d", total)
	}
}

func TestSSEHandler_GetStats(t *testing.T) {
	handler := api.NewSSEHandler()

	stats := handler.GetStats()

	if stats["total_clients"] != 0 {
		t.Error("Expected 0 total clients in stats")
	}

	channels, ok := stats["channels"].(map[string]int)
	if !ok {
		t.Error("Expected channels map in stats")
		return
	}

	if len(channels) != 0 {
		t.Error("Expected empty channels map")
	}
}

func TestSSEHandler_Publish(t *testing.T) {
	handler := api.NewSSEHandler()

	// Publish should not panic even with no clients
	handler.Publish("test-channel", "test-event", map[string]interface{}{
		"message": "hello",
	})
}

func TestSSEHandler_BroadcastTaskUpdate(t *testing.T) {
	handler := api.NewSSEHandler()

	// Should not panic
	handler.BroadcastTaskUpdate("task-123", "running", map[string]interface{}{
		"progress": 50,
	})
}

func TestSSEHandler_BroadcastPipelineUpdate(t *testing.T) {
	handler := api.NewSSEHandler()

	// Should not panic
	handler.BroadcastPipelineUpdate("pipeline-456", "completed", map[string]interface{}{
		"duration": 5000,
	})
}

func TestSSEHandler_BroadcastRunUpdate(t *testing.T) {
	handler := api.NewSSEHandler()

	// Should not panic
	handler.BroadcastRunUpdate("run-789", "step-1", map[string]interface{}{
		"status": "success",
	})
}

func TestSSEHandler_BroadcastLog(t *testing.T) {
	handler := api.NewSSEHandler()

	// Should not panic
	handler.BroadcastLog("logs-channel", "info", "Application started")
}