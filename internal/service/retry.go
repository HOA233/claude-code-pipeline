package service

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryPolicy defines how retries should be performed
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	Multiplier      float64       `json:"multiplier"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryPolicy returns a default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// ExponentialBackoffPolicy returns a policy with exponential backoff
func ExponentialBackoffPolicy(maxAttempts int) *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  maxAttempts,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Minute,
		Multiplier:   2.0,
	}
}

// LinearBackoffPolicy returns a policy with linear backoff
func LinearBackoffPolicy(maxAttempts int, delay time.Duration) *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  maxAttempts,
		InitialDelay: delay,
		MaxDelay:     delay * time.Duration(maxAttempts),
		Multiplier:   1.0,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithContext is a function with context that can be retried
type RetryableFuncWithContext func(ctx context.Context) error

// RetryExecutor handles retry logic
type RetryExecutor struct {
	policy *RetryPolicy
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(policy *RetryPolicy) *RetryExecutor {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}
	return &RetryExecutor{policy: policy}
}

// Execute executes a function with retry logic
func (e *RetryExecutor) Execute(fn RetryableFunc) error {
	return e.ExecuteWithContext(context.Background(), func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteWithContext executes a function with retry logic and context
func (e *RetryExecutor) ExecuteWithContext(ctx context.Context, fn RetryableFuncWithContext) error {
	var lastErr error

	for attempt := 1; attempt <= e.policy.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !e.isRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't delay after last attempt
		if attempt < e.policy.MaxAttempts {
			delay := e.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", e.policy.MaxAttempts, lastErr)
}

// ExecuteWithResult executes a function that returns a result with retry logic
func (e *RetryExecutor) ExecuteWithResult(fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := e.Execute(func() error {
		var err error
		result, err = fn()
		return err
	})
	return result, err
}

// calculateDelay calculates the delay for a given attempt
func (e *RetryExecutor) calculateDelay(attempt int) time.Duration {
	delay := float64(e.policy.InitialDelay)
	delay = delay * math.Pow(e.policy.Multiplier, float64(attempt-1))

	if delay > float64(e.policy.MaxDelay) {
		delay = float64(e.policy.MaxDelay)
	}

	return time.Duration(delay)
}

// isRetryable checks if an error is retryable
func (e *RetryExecutor) isRetryable(err error) bool {
	if len(e.policy.RetryableErrors) == 0 {
		return true // All errors are retryable by default
	}

	errStr := err.Error()
	for _, retryable := range e.policy.RetryableErrors {
		if retryable == errStr {
			return true
		}
	}
	return false
}

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Success      bool          `json:"success"`
	Attempts     int           `json:"attempts"`
	TotalTime    time.Duration `json:"total_time"`
	LastDelay    time.Duration `json:"last_delay"`
	LastError    string        `json:"last_error,omitempty"`
	FinalResult  interface{}   `json:"final_result,omitempty"`
}

// ExecuteWithDetails executes and returns detailed retry information
func (e *RetryExecutor) ExecuteWithDetails(fn func() (interface{}, error)) *RetryResult {
	start := time.Now()
	result := &RetryResult{}

	var lastErr error
	var lastDelay time.Duration

	for attempt := 1; attempt <= e.policy.MaxAttempts; attempt++ {
		result.Attempts = attempt

		res, err := fn()
		if err == nil {
			result.Success = true
			result.FinalResult = res
			result.TotalTime = time.Since(start)
			return result
		}

		lastErr = err
		result.LastError = err.Error()

		if !e.isRetryable(err) {
			break
		}

		if attempt < e.policy.MaxAttempts {
			lastDelay = e.calculateDelay(attempt)
			time.Sleep(lastDelay)
		}
	}

	result.TotalTime = time.Since(start)
	result.LastDelay = lastDelay
	return result
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures   int
	timeout       time.Duration
	state         string // "closed", "open", "half-open"
	failures      int
	lastFailTime  time.Time
	successCount  int
	successTarget int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:   maxFailures,
		timeout:       timeout,
		state:         "closed",
		successTarget: 3,
	}
}

// Execute executes a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) canExecute() bool {
	if cb.state == "closed" {
		return true
	}

	if cb.state == "open" {
		// Check if timeout has passed
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = "half-open"
			cb.successCount = 0
			return true
		}
		return false
	}

	// half-open state
	return true
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == "half-open" {
		cb.state = "open"
	} else if cb.failures >= cb.maxFailures {
		cb.state = "open"
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0

	if cb.state == "half-open" {
		cb.successCount++
		if cb.successCount >= cb.successTarget {
			cb.state = "closed"
		}
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() string {
	return cb.state
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.state = "closed"
	cb.failures = 0
	cb.successCount = 0
}