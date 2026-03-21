package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled  bool
	APIKeys  map[string]bool // set of valid API keys
	JWTSecret string
}

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	if config == nil || !config.Enabled {
		// No auth required
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Skip auth for health and status endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/status" {
			c.Next()
			return
		}

		// Check API key in header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if config.APIKeys[apiKey] {
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		// Check Authorization header (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token := parts[1]
				// Simple token validation (extend for JWT)
				if config.APIKeys[token] {
					c.Set("auth_method", "bearer")
					c.Next()
					return
				}
			}
		}

		// Unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Valid API key or Bearer token required",
		})
		c.Abort()
	}
}

// OptionalAuthMiddleware creates an optional auth middleware
// It will set auth info if present but won't reject requests
func OptionalAuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" && config != nil && config.APIKeys[apiKey] {
			c.Set("authenticated", true)
			c.Set("auth_method", "api_key")
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && config != nil {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if config.APIKeys[parts[1]] {
					c.Set("authenticated", true)
					c.Set("auth_method", "bearer")
				}
			}
		}

		c.Next()
	}
}

// AdminMiddleware requires admin privileges
func AdminMiddleware(adminKeys map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 {
					apiKey = parts[1]
				}
			}
		}

		if !adminKeys[apiKey] {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}