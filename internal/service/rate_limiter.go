package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if N tokens can be consumed
func (rl *RateLimiter) AllowN(n float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= n {
		rl.tokens -= n
		return true
	}
	return false
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.WaitN(ctx, 1)
}

// WaitN blocks until N tokens are available
func (rl *RateLimiter) WaitN(ctx context.Context, n float64) error {
	for {
		if rl.AllowN(n) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// Tokens returns the current number of tokens
func (rl *RateLimiter) Tokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	return rl.tokens
}

// SlidingWindowLimiter implements a sliding window rate limiter
type SlidingWindowLimiter struct {
	mu        sync.Mutex
	requests  []time.Time
	window    time.Duration
	maxReqs   int
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(window time.Duration, maxReqs int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		requests: make([]time.Time, 0),
		window:   window,
		maxReqs:  maxReqs,
	}
}

// Allow checks if a request is allowed
func (sw *SlidingWindowLimiter) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// Remove old requests
	valid := sw.requests[:0]
	for _, t := range sw.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	sw.requests = valid

	if len(sw.requests) >= sw.maxReqs {
		return false
	}

	sw.requests = append(sw.requests, now)
	return true
}

// Count returns the current request count in the window
func (sw *SlidingWindowLimiter) Count() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	valid := sw.requests[:0]
	for _, t := range sw.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	sw.requests = valid

	return len(sw.requests)
}

// ResetTime returns when the oldest request will expire
func (sw *SlidingWindowLimiter) ResetTime() time.Time {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if len(sw.requests) == 0 {
		return time.Now()
	}

	return sw.requests[0].Add(sw.window)
}

// RateLimiterRegistry manages multiple rate limiters
type RateLimiterRegistry struct {
	mu       sync.RWMutex
	limiters map[string]*SlidingWindowLimiter
	defaults struct {
		window  time.Duration
		maxReqs int
	}
}

// NewRateLimiterRegistry creates a new rate limiter registry
func NewRateLimiterRegistry(defaultWindow time.Duration, defaultMaxReqs int) *RateLimiterRegistry {
	return &RateLimiterRegistry{
		limiters: make(map[string]*SlidingWindowLimiter),
		defaults: struct {
			window  time.Duration
			maxReqs int
		}{window: defaultWindow, maxReqs: defaultMaxReqs},
	}
}

// GetOrCreate gets or creates a rate limiter for a key
func (r *RateLimiterRegistry) GetOrCreate(key string) *SlidingWindowLimiter {
	r.mu.RLock()
	limiter, exists := r.limiters[key]
	r.mu.RUnlock()

	if exists {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check
	if limiter, exists := r.limiters[key]; exists {
		return limiter
	}

	limiter = NewSlidingWindowLimiter(r.defaults.window, r.defaults.maxReqs)
	r.limiters[key] = limiter
	return limiter
}

// Allow checks if a request is allowed for a key
func (r *RateLimiterRegistry) Allow(key string) bool {
	return r.GetOrCreate(key).Allow()
}

// GetStats returns stats for all rate limiters
func (r *RateLimiterRegistry) GetStats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]int)
	for key, limiter := range r.limiters {
		stats[key] = limiter.Count()
	}
	return stats
}

// LeakyBucket implements a leaky bucket rate limiter
type LeakyBucket struct {
	mu        sync.Mutex
	queue     chan struct{}
	interval  time.Duration
	lastLeak  time.Time
}

// NewLeakyBucket creates a new leaky bucket rate limiter
func NewLeakyBucket(capacity int, interval time.Duration) *LeakyBucket {
	return &LeakyBucket{
		queue:    make(chan struct{}, capacity),
		interval: interval,
	}
}

// Allow tries to add a request to the bucket
func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	select {
	case lb.queue <- struct{}{}:
		return true
	default:
		return false
	}
}

// leak removes requests based on elapsed time
func (lb *LeakyBucket) leak() {
	now := time.Now()
	if lb.lastLeak.IsZero() {
		lb.lastLeak = now
		return
	}

	elapsed := now.Sub(lb.lastLeak)
	tokensToLeak := int(elapsed / lb.interval)

	for i := 0; i < tokensToLeak; i++ {
		select {
		case <-lb.queue:
		default:
		}
	}

	lb.lastLeak = now
}

// QueueLength returns the current queue length
func (lb *LeakyBucket) QueueLength() int {
	return len(lb.queue)
}

// RequestValidator validates request parameters
type RequestValidator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Required  bool
	Type      string
	Min       interface{}
	Max       interface{}
	Pattern   string
	Enum      []interface{}
	Custom    func(interface{}) error
}

// NewRequestValidator creates a new request validator
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule adds a validation rule for a field
func (v *RequestValidator) AddRule(field string, rule ValidationRule) {
	v.rules[field] = rule
}

// Validate validates a map of values
func (v *RequestValidator) Validate(values map[string]interface{}) error {
	for field, rule := range v.rules {
		value, exists := values[field]

		if !exists || value == nil {
			if rule.Required {
				return fmt.Errorf("field '%s' is required", field)
			}
			continue
		}

		if rule.Custom != nil {
			if err := rule.Custom(value); err != nil {
				return fmt.Errorf("field '%s': %w", field, err)
			}
		}
	}

	return nil
}