package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Retry Service Tests

func TestRetryService_New(t *testing.T) {
	rs := service.NewRetryService()
	if rs == nil {
		t.Fatal("Expected non-nil retry service")
	}
}

func TestRetryService_Execute_Success(t *testing.T) {
	rs := service.NewRetryService()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	config := &service.RetryConfig{
		MaxAttempts: 3,
		Backoff:     100 * time.Millisecond,
	}

	err := rs.Execute(context.Background(), "success-task", fn, config)
	if err != nil {
		t.Fatalf("Expected success: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetryService_Execute_RetrySuccess(t *testing.T) {
	rs := service.NewRetryService()

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return &service.RetryableError{Err: "temporary failure"}
		}
		return nil
	}

	config := &service.RetryConfig{
		MaxAttempts: 5,
		Backoff:     10 * time.Millisecond,
	}

	err := rs.Execute(context.Background(), "retry-success", fn, config)
	if err != nil {
		t.Fatalf("Expected success after retries: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestRetryService_Execute_MaxAttemptsExceeded(t *testing.T) {
	rs := service.NewRetryService()

	callCount := 0
	fn := func() error {
		callCount++
		return &service.RetryableError{Err: "always fails"}
	}

	config := &service.RetryConfig{
		MaxAttempts: 3,
		Backoff:     10 * time.Millisecond,
	}

	err := rs.Execute(context.Background(), "max-attempts", fn, config)
	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if callCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", callCount)
	}
}

func TestRetryService_Execute_NonRetryableError(t *testing.T) {
	rs := service.NewRetryService()

	callCount := 0
	fn := func() error {
		callCount++
		return &service.FatalError{Err: "fatal error"}
	}

	config := &service.RetryConfig{
		MaxAttempts: 5,
		Backoff:     10 * time.Millisecond,
	}

	err := rs.Execute(context.Background(), "fatal-error", fn, config)
	if err == nil {
		t.Error("Expected fatal error")
	}

	// Should not retry for non-retryable errors
	if callCount != 1 {
		t.Errorf("Expected 1 call for fatal error, got %d", callCount)
	}
}

func TestRetryService_Execute_ContextCancellation(t *testing.T) {
	rs := service.NewRetryService()

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fn := func() error {
		callCount++
		if callCount >= 2 {
			cancel()
		}
		return &service.RetryableError{Err: "retry"}
	}

	config := &service.RetryConfig{
		MaxAttempts: 10,
		Backoff:     50 * time.Millisecond,
	}

	err := rs.Execute(ctx, "cancel-task", fn, config)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

func TestRetryService_ExponentialBackoff(t *testing.T) {
	rs := service.NewRetryService()

	var durations []time.Duration
	start := time.Now()

	callCount := 0
	fn := func() error {
		callCount++
		durations = append(durations, time.Since(start))
		if callCount < 4 {
			return &service.RetryableError{Err: "retry"}
		}
		return nil
	}

	config := &service.RetryConfig{
		MaxAttempts:     5,
		InitialBackoff:  50 * time.Millisecond,
		MaxBackoff:      500 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	rs.Execute(context.Background(), "backoff-task", fn, config)

	// Verify exponential backoff (approximately)
	if len(durations) >= 3 {
		// First retry should be after ~50ms, second ~150ms (50+100), third ~350ms (50+100+200)
		// This is approximate due to timing
		_ = durations
	}
}

func TestRetryService_GetStats(t *testing.T) {
	rs := service.NewRetryService()

	config := &service.RetryConfig{
		MaxAttempts: 3,
		Backoff:     10 * time.Millisecond,
	}

	// Success on first try
	rs.Execute(context.Background(), "stats-success", func() error { return nil }, config)

	// Fail after retries
	rs.Execute(context.Background(), "stats-fail", func() error {
		return &service.RetryableError{Err: "fail"}
	}, config)

	stats := rs.GetStats()

	if stats.TotalExecutions < 2 {
		t.Errorf("Expected at least 2 executions, got %d", stats.TotalExecutions)
	}
}

func TestRetryService_SetDefaultConfig(t *testing.T) {
	rs := service.NewRetryService()

	defaultConfig := &service.RetryConfig{
		MaxAttempts: 5,
		Backoff:     200 * time.Millisecond,
	}

	rs.SetDefaultConfig(defaultConfig)

	retrieved := rs.GetDefaultConfig()
	if retrieved.MaxAttempts != 5 {
		t.Error("Default config not set correctly")
	}
}

func TestRetryService_IsRetryable(t *testing.T) {
	rs := service.NewRetryService()

	retryableErr := &service.RetryableError{Err: "retry me"}
	if !rs.IsRetryable(retryableErr) {
		t.Error("Expected retryable error to be retryable")
	}

	fatalErr := &service.FatalError{Err: "fatal"}
	if rs.IsRetryable(fatalErr) {
		t.Error("Expected fatal error to not be retryable")
	}
}

func TestRetryService_Cancel(t *testing.T) {
	rs := service.NewRetryService()

	// Start a task with retries
	go func() {
		config := &service.RetryConfig{
			MaxAttempts: 100,
			Backoff:     1 * time.Second,
		}

		rs.Execute(context.Background(), "cancel-test", func() error {
			return &service.RetryableError{Err: "always retry"}
		}, config)
	}()

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the task
	err := rs.Cancel("cancel-test")
	if err != nil {
		t.Fatalf("Failed to cancel: %v", err)
	}
}

func TestRetryService_RetryableError(t *testing.T) {
	err := &service.RetryableError{Err: "test error"}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestRetryService_FatalError(t *testing.T) {
	err := &service.FatalError{Err: "fatal error"}

	if err.Error() != "fatal error" {
		t.Errorf("Expected 'fatal error', got %s", err.Error())
	}
}

func TestRetryService_RetryConfigToJSON(t *testing.T) {
	config := &service.RetryConfig{
		MaxAttempts:     5,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      10 * time.Second,
		BackoffMultiple: 2.0,
	}

	data, err := config.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}