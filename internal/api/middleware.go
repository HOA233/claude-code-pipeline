package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware logs request details
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		zap.L().Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		)
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		zap.L().Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// RateLimitMiddleware implements simple rate limiting
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	type clientInfo struct {
		count     int
		resetTime time.Time
	}

	clients := make(map[string]*clientInfo)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		info, exists := clients[clientIP]
		if !exists || now.After(info.resetTime) {
			clients[clientIP] = &clientInfo{
				count:     1,
				resetTime: now.Add(time.Minute),
			}
		} else {
			info.count++
			if info.count > requestsPerMinute {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded",
				})
				return
			}
		}

		c.Next()
	}
}

// AuthMiddleware validates API key
func AuthMiddleware(apiKeys []string) gin.HandlerFunc {
	validKeys := make(map[string]bool)
	for _, key := range apiKeys {
		validKeys[key] = true
	}

	return func(c *gin.Context) {
		// Skip auth for health and status endpoints
		path := c.Request.URL.Path
		if path == "/health" || path == "/api/status" {
			c.Next()
			return
		}

		// Check Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			// Allow requests without auth if no keys configured
			if len(validKeys) == 0 {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization required",
			})
			return
		}

		// Bearer token
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token := auth[7:]
			if validKeys[token] {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
		})
	}
}

// RequestIDMiddleware adds a unique request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}