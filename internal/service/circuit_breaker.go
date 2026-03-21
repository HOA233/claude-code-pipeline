package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern for resilience
type CircuitBreaker struct {
	mu              sync.RWMutex
	name            string
	state           CircuitState
	failureCount    int
	successCount    int
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	lastFailureTime time.Time
	halfOpenCalls   int
	maxHalfOpenCalls int
	onStateChange   func(from, to CircuitState)
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half-open"
)

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name             string
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	MaxHalfOpenCalls int
	OnStateChange    func(from, to CircuitState)
}

// CircuitBreakerError represents circuit breaker specific errors
var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(cfg *CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 3
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxHalfOpenCalls == 0 {
		cfg.MaxHalfOpenCalls = 1
	}

	return &CircuitBreaker{
		name:             cfg.Name,
		state:            StateClosed,
		failureThreshold: cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		timeout:          cfg.Timeout,
		maxHalfOpenCalls: cfg.MaxHalfOpenCalls,
		onStateChange:    cfg.OnStateChange,
	}
}

// Execute runs the given function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// ExecuteWithFallback runs the function with a fallback
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func(error) error) error {
	err := cb.Execute(ctx, fn)
	if err != nil {
		if errors.Is(err, ErrCircuitOpen) || errors.Is(err, ErrTooManyRequests) {
			return fallback(err)
		}
		return err
	}
	return nil
}

// canExecute checks if a request can be executed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenCalls = 0
			return true
		}
		return false
	case StateHalfOpen:
		if cb.halfOpenCalls >= cb.maxHalfOpenCalls {
			return false
		}
		cb.halfOpenCalls++
		return true
	default:
		return false
	}
}

// recordFailure records a failure and updates state
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.state == StateHalfOpen {
		cb.setState(StateOpen)
	} else if cb.failureCount >= cb.failureThreshold {
		cb.setState(StateOpen)
	}
}

// recordSuccess records a success and updates state
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.setState(StateClosed)
		}
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		if cb.onStateChange != nil {
			go cb.onStateChange(oldState, newState)
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":            cb.name,
		"state":           string(cb.state),
		"failure_count":   cb.failureCount,
		"success_count":   cb.successCount,
		"failure_threshold": cb.failureThreshold,
		"success_threshold": cb.successThreshold,
		"timeout_ms":      cb.timeout.Milliseconds(),
		"last_failure":    cb.lastFailureTime,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
	cb.failureCount = 0
	cb.successCount = 0
}

// CircuitBreakerPool manages multiple circuit breakers
type CircuitBreakerPool struct {
	mu        sync.RWMutex
	breakers  map[string]*CircuitBreaker
	defaultCfg CircuitBreakerConfig
}

// NewCircuitBreakerPool creates a new circuit breaker pool
func NewCircuitBreakerPool(defaultCfg CircuitBreakerConfig) *CircuitBreakerPool {
	return &CircuitBreakerPool{
		breakers:   make(map[string]*CircuitBreaker),
		defaultCfg: defaultCfg,
	}
}

// Get retrieves or creates a circuit breaker by name
func (p *CircuitBreakerPool) Get(name string) *CircuitBreaker {
	p.mu.RLock()
	breaker, exists := p.breakers[name]
	p.mu.RUnlock()

	if exists {
		return breaker
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check after acquiring write lock
	if breaker, exists = p.breakers[name]; exists {
		return breaker
	}

	cfg := p.defaultCfg
	cfg.Name = name
	breaker = NewCircuitBreaker(&cfg)
	p.breakers[name] = breaker

	return breaker
}

// Execute runs a function through the named circuit breaker
func (p *CircuitBreakerPool) Execute(ctx context.Context, name string, fn func() error) error {
	return p.Get(name).Execute(ctx, fn)
}

// GetAllStats returns statistics for all circuit breakers
func (p *CircuitBreakerPool) GetAllStats() map[string]map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, breaker := range p.breakers {
		stats[name] = breaker.GetStats()
	}
	return stats
}

// Remove removes a circuit breaker from the pool
func (p *CircuitBreakerPool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.breakers, name)
}