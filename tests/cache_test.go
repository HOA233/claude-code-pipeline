package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Cache Service Tests

func TestCacheService_New(t *testing.T) {
	cs := service.NewCacheService()
	if cs == nil {
		t.Fatal("Expected non-nil cache service")
	}
}

func TestCacheService_Set(t *testing.T) {
	cs := service.NewCacheService()

	err := cs.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
}

func TestCacheService_Get(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("key1", "value1", 5*time.Minute)

	value, found := cs.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
}

func TestCacheService_Get_NotFound(t *testing.T) {
	cs := service.NewCacheService()

	_, found := cs.Get("nonexistent")
	if found {
		t.Error("Expected not found for nonexistent key")
	}
}

func TestCacheService_Delete(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("delete-key", "value", 5*time.Minute)

	err := cs.Delete("delete-key")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found := cs.Get("delete-key")
	if found {
		t.Error("Expected key to be deleted")
	}
}

func TestCacheService_Delete_NotFound(t *testing.T) {
	cs := service.NewCacheService()

	err := cs.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent key")
	}
}

func TestCacheService_Has(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("has-key", "value", 5*time.Minute)

	if !cs.Has("has-key") {
		t.Error("Expected Has to return true")
	}

	if cs.Has("nonexistent") {
		t.Error("Expected Has to return false for nonexistent key")
	}
}

func TestCacheService_Clear(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("key1", "value1", 5*time.Minute)
	cs.Set("key2", "value2", 5*time.Minute)

	cs.Clear()

	if cs.Has("key1") || cs.Has("key2") {
		t.Error("Expected cache to be cleared")
	}
}

func TestCacheService_GetOrSet(t *testing.T) {
	cs := service.NewCacheService()

	callCount := 0
	loader := func() (interface{}, error) {
		callCount++
		return "loaded-value", nil
	}

	// First call should load
	value, err := cs.GetOrSet("load-key", loader, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to GetOrSet: %v", err)
	}

	if value != "loaded-value" {
		t.Errorf("Expected loaded-value, got %v", value)
	}

	if callCount != 1 {
		t.Errorf("Expected loader to be called once, got %d", callCount)
	}

	// Second call should use cache
	value, _ = cs.GetOrSet("load-key", loader, 5*time.Minute)
	if callCount != 1 {
		t.Error("Expected loader not to be called again")
	}
}

func TestCacheService_SetNX(t *testing.T) {
	cs := service.NewCacheService()

	// Key doesn't exist, should set
	set := cs.SetNX("nx-key", "value1", 5*time.Minute)
	if !set {
		t.Error("Expected SetNX to return true for new key")
	}

	// Key exists, should not set
	set = cs.SetNX("nx-key", "value2", 5*time.Minute)
	if set {
		t.Error("Expected SetNX to return false for existing key")
	}

	// Original value should remain
	value, _ := cs.Get("nx-key")
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
}

func TestCacheService_TTL(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("ttl-key", "value", 1*time.Second)

	ttl := cs.TTL("ttl-key")
	if ttl <= 0 {
		t.Error("Expected positive TTL")
	}

	// Nonexistent key
	ttl = cs.TTL("nonexistent")
	if ttl != 0 {
		t.Error("Expected 0 TTL for nonexistent key")
	}
}

func TestCacheService_Increment(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("counter", 0, 5*time.Minute)

	val, err := cs.Increment("counter", 1)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	val, _ = cs.Increment("counter", 5)
	if val != 6 {
		t.Errorf("Expected 6, got %d", val)
	}
}

func TestCacheService_Decrement(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("counter", 10, 5*time.Minute)

	val, err := cs.Decrement("counter", 3)
	if err != nil {
		t.Fatalf("Failed to decrement: %v", err)
	}

	if val != 7 {
		t.Errorf("Expected 7, got %d", val)
	}
}

func TestCacheService_Keys(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("prefix:key1", "value1", 5*time.Minute)
	cs.Set("prefix:key2", "value2", 5*time.Minute)
	cs.Set("other:key3", "value3", 5*time.Minute)

	keys := cs.Keys("prefix:")
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys with prefix, got %d", len(keys))
	}
}

func TestCacheService_Size(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("key1", "value1", 5*time.Minute)
	cs.Set("key2", "value2", 5*time.Minute)

	size := cs.Size()
	if size < 2 {
		t.Errorf("Expected at least 2 items, got %d", size)
	}
}

func TestCacheService_Expiration(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("expire-key", "value", 100*time.Millisecond)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, found := cs.Get("expire-key")
	if found {
		t.Error("Expected key to be expired")
	}
}

func TestCacheService_GetStats(t *testing.T) {
	cs := service.NewCacheService()

	cs.Set("key1", "value1", 5*time.Minute)
	cs.Get("key1") // hit
	cs.Get("nonexistent") // miss

	stats := cs.GetStats()

	if stats.Hits < 1 {
		t.Errorf("Expected at least 1 hit, got %d", stats.Hits)
	}

	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 miss, got %d", stats.Misses)
	}
}