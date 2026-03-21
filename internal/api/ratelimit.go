package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type clientInfo struct {
	count     int
	expiresAt time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		rate:     rate,
		window:   window,
	}

	// Cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	info, exists := rl.requests[clientID]
	if !exists || now.After(info.expiresAt) {
		rl.requests[clientID] = &clientInfo{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true
	}

	if info.count >= rl.rate {
		return false
	}

	info.count++
	return true
}

// cleanup removes expired entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, info := range rl.requests {
			if now.After(info.expiresAt) {
				delete(rl.requests, id)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		clientID := c.ClientIP()

		// Use API key if available for more granular limiting
		if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
			clientID = apiKey
		}

		if !limiter.Allow(clientID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByPath creates a rate limiter for specific paths
func RateLimitByPath(pathRates map[string]struct {
	Rate   int
	Window time.Duration
}) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)

	for path, config := range pathRates {
		limiters[path] = NewRateLimiter(config.Rate, config.Window)
	}

	return func(c *gin.Context) {
		path := c.FullPath()

		if limiter, exists := limiters[path]; exists {
			clientID := c.ClientIP()

			if !limiter.Allow(clientID) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "rate limit exceeded",
					"message": "Too many requests for this endpoint",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}