package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

func TestCircuitBreaker_New(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		Name:             "test",
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          5 * time.Second,
	})

	if cb.GetState() != service.StateClosed {
		t.Error("Expected initial state to be closed")
	}
}

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	// Successful calls should work
	for i := 0; i < 5; i++ {
		err := cb.Execute(context.Background(), func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	}

	if cb.GetState() != service.StateClosed {
		t.Error("Expected state to remain closed")
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	stateChanges := 0
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
		OnStateChange: func(from, to service.CircuitState) {
			stateChanges++
		},
	})

	// Cause failures to trip the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != service.StateOpen {
		t.Errorf("Expected state to be open, got: %s", cb.GetState())
	}

	// Should reject calls when open
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if !errors.Is(err, service.ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got: %v", err)
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})

	// Trip the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != service.StateOpen {
		t.Fatal("Expected state to be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open and succeed
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected success in half-open, got: %v", err)
	}

	if cb.GetState() != service.StateHalfOpen {
		t.Fatal("Expected state to be half-open")
	}

	// Another success to close
	err = cb.Execute(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	if cb.GetState() != service.StateClosed {
		t.Errorf("Expected state to be closed, got: %s", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})

	// Trip the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// This should transition to half-open
	cb.Execute(context.Background(), func() error {
		return nil
	})

	// Failure in half-open should open again
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})

	if cb.GetState() != service.StateOpen {
		t.Errorf("Expected state to be open after half-open failure")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 2,
	})

	// Trip the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != service.StateOpen {
		t.Fatal("Expected state to be open")
	}

	// Reset
	cb.Reset()

	if cb.GetState() != service.StateClosed {
		t.Error("Expected state to be closed after reset")
	}
}

func TestCircuitBreaker_ExecuteWithFallback(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		FailureThreshold: 1,
	})

	// Trip the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("failure")
	})

	fallbackCalled := false
	err := cb.ExecuteWithFallback(context.Background(), func() error {
		return errors.New("should not be called")
	}, func(e error) error {
		fallbackCalled = true
		return nil
	})

	if !fallbackCalled {
		t.Error("Expected fallback to be called")
	}
	if err != nil {
		t.Errorf("Expected fallback to handle error, got: %v", err)
	}
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := service.NewCircuitBreaker(&service.CircuitBreakerConfig{
		Name:             "stats-test",
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          10 * time.Second,
	})

	stats := cb.GetStats()

	if stats["name"] != "stats-test" {
		t.Error("Expected name in stats")
	}
	if stats["state"] != "closed" {
		t.Error("Expected closed state in stats")
	}
	if stats["failure_threshold"] != 5 {
		t.Error("Expected failure threshold in stats")
	}
}

func TestCircuitBreakerPool_Get(t *testing.T) {
	pool := service.NewCircuitBreakerPool(service.CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	cb1 := pool.Get("service-a")
	cb2 := pool.Get("service-b")
	cb1Again := pool.Get("service-a")

	if cb1 == cb2 {
		t.Error("Different services should have different breakers")
	}

	if cb1 != cb1Again {
		t.Error("Same service should return same breaker")
	}
}

func TestCircuitBreakerPool_Execute(t *testing.T) {
	pool := service.NewCircuitBreakerPool(service.CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	err := pool.Execute(context.Background(), "test-service", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCircuitBreakerPool_GetAllStats(t *testing.T) {
	pool := service.NewCircuitBreakerPool(service.CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	pool.Get("service-a")
	pool.Get("service-b")

	stats := pool.GetAllStats()

	if len(stats) != 2 {
		t.Errorf("Expected 2 services in stats, got %d", len(stats))
	}
}

func TestCircuitBreakerPool_Remove(t *testing.T) {
	pool := service.NewCircuitBreakerPool(service.CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	pool.Get("service-a")
	pool.Remove("service-a")

	stats := pool.GetAllStats()

	if len(stats) != 0 {
		t.Error("Expected no services after removal")
	}
}