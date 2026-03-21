package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// RateLimiter Service Tests

func TestRateLimiter_New(t *testing.T) {
	rl := service.NewRateLimiterService()
	if rl == nil {
		t.Fatal("Expected non-nil rate limiter service")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := service.NewRateLimiterService()

	// Create a rule allowing 10 requests per second
	rule := &service.RateLimitRule{
		ID:        "rule-1",
		Name:      "Test Rule",
		Rate:      10,
		Window:    time.Second,
		KeyPrefix: "test:",
	}

	rl.SetRule(rule)

	// Should allow the first request
	allowed := rl.Allow("rule-1", "user-1")
	if !allowed {
		t.Error("Expected request to be allowed")
	}
}

func TestRateLimiter_RateLimit(t *testing.T) {
	rl := service.NewRateLimiterService()

	// Rule allowing only 2 requests per second
	rule := &service.RateLimitRule{
		ID:        "limit-rule",
		Name:      "Limit Rule",
		Rate:      2,
		Window:    time.Second,
		KeyPrefix: "limit:",
	}
	rl.SetRule(rule)

	// First two should be allowed
	rl.Allow("limit-rule", "user-limit")
	rl.Allow("limit-rule", "user-limit")

	// Third should be denied
	allowed := rl.Allow("limit-rule", "user-limit")
	if allowed {
		t.Error("Expected third request to be denied")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "reset-rule",
		Name:      "Reset Rule",
		Rate:      1,
		Window:    time.Second,
		KeyPrefix: "reset:",
	}
	rl.SetRule(rule)

	rl.Allow("reset-rule", "user-reset")
	// Should be limited now

	rl.Reset("reset-rule", "user-reset")

	// Should be allowed after reset
	allowed := rl.Allow("reset-rule", "user-reset")
	if !allowed {
		t.Error("Expected request to be allowed after reset")
	}
}

func TestRateLimiter_GetCount(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "count-rule",
		Name:      "Count Rule",
		Rate:      10,
		Window:    time.Second,
		KeyPrefix: "count:",
	}
	rl.SetRule(rule)

	rl.Allow("count-rule", "user-count")
	rl.Allow("count-rule", "user-count")
	rl.Allow("count-rule", "user-count")

	count := rl.GetCount("count-rule", "user-count")
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "remain-rule",
		Name:      "Remain Rule",
		Rate:      5,
		Window:    time.Second,
		KeyPrefix: "remain:",
	}
	rl.SetRule(rule)

	rl.Allow("remain-rule", "user-remain")
	rl.Allow("remain-rule", "user-remain")

	remaining := rl.GetRemaining("remain-rule", "user-remain")
	if remaining != 3 {
		t.Errorf("Expected 3 remaining, got %d", remaining)
	}
}

func TestRateLimiter_GetResetTime(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "time-rule",
		Name:      "Time Rule",
		Rate:      10,
		Window:    time.Minute,
		KeyPrefix: "time:",
	}
	rl.SetRule(rule)

	rl.Allow("time-rule", "user-time")

	resetTime := rl.GetResetTime("time-rule", "user-time")
	if resetTime.IsZero() {
		t.Error("Expected non-zero reset time")
	}
}

func TestRateLimiter_SetRule(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "set-rule",
		Name:      "Set Rule",
		Rate:      100,
		Window:    time.Minute,
		KeyPrefix: "set:",
	}

	err := rl.SetRule(rule)
	if err != nil {
		t.Fatalf("Failed to set rule: %v", err)
	}
}

func TestRateLimiter_GetRule(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "get-rule",
		Name:      "Get Rule",
		Rate:      50,
		Window:    time.Hour,
		KeyPrefix: "get:",
	}
	rl.SetRule(rule)

	retrieved := rl.GetRule("get-rule")
	if retrieved == nil {
		t.Fatal("Expected rule")
	}

	if retrieved.Name != "Get Rule" {
		t.Error("Rule name mismatch")
	}
}

func TestRateLimiter_DeleteRule(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "delete-rule",
		Name:      "Delete Rule",
		Rate:      10,
		Window:    time.Second,
		KeyPrefix: "delete:",
	}
	rl.SetRule(rule)

	rl.DeleteRule("delete-rule")

	retrieved := rl.GetRule("delete-rule")
	if retrieved != nil {
		t.Error("Expected rule to be deleted")
	}
}

func TestRateLimiter_ListRules(t *testing.T) {
	rl := service.NewRateLimiterService()

	rl.SetRule(&service.RateLimitRule{
		ID:        "list-1",
		Name:      "List Rule 1",
		Rate:      10,
		Window:    time.Second,
		KeyPrefix: "list1:",
	})

	rl.SetRule(&service.RateLimitRule{
		ID:        "list-2",
		Name:      "List Rule 2",
		Rate:      100,
		Window:    time.Minute,
		KeyPrefix: "list2:",
	})

	rules := rl.ListRules()
	if len(rules) < 2 {
		t.Errorf("Expected at least 2 rules, got %d", len(rules))
	}
}

func TestRateLimiter_GetStats(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "stats-rule",
		Name:      "Stats Rule",
		Rate:      10,
		Window:    time.Second,
		KeyPrefix: "stats:",
	}
	rl.SetRule(rule)

	rl.Allow("stats-rule", "user-stats")
	rl.Allow("stats-rule", "user-stats")

	stats := rl.GetStats("stats-rule", "user-stats")

	if stats.Requests != 2 {
		t.Errorf("Expected 2 requests, got %d", stats.Requests)
	}

	if stats.Remaining != 8 {
		t.Errorf("Expected 8 remaining, got %d", stats.Remaining)
	}
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	rl := service.NewRateLimiterService()

	rule := &service.RateLimitRule{
		ID:        "multi-user",
		Name:      "Multi User Rule",
		Rate:      1,
		Window:    time.Second,
		KeyPrefix: "multi:",
	}
	rl.SetRule(rule)

	// User 1 uses their quota
	rl.Allow("multi-user", "user-1")

	// User 2 should still have quota
	allowed := rl.Allow("multi-user", "user-2")
	if !allowed {
		t.Error("User 2 should have their own quota")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := service.NewRateLimiterService()

	// Very short window
	rule := &service.RateLimitRule{
		ID:        "expiry-rule",
		Name:      "Expiry Rule",
		Rate:      1,
		Window:    100 * time.Millisecond,
		KeyPrefix: "expiry:",
	}
	rl.SetRule(rule)

	rl.Allow("expiry-rule", "user-expiry")

	// Should be denied
	if rl.Allow("expiry-rule", "user-expiry") {
		t.Error("Expected to be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("expiry-rule", "user-expiry") {
		t.Error("Expected to be allowed after window expiry")
	}
}

func TestRateLimiter_RateLimitRuleToJSON(t *testing.T) {
	rule := &service.RateLimitRule{
		ID:        "json-rule",
		Name:      "JSON Rule",
		Rate:      100,
		Window:    time.Minute,
		KeyPrefix: "json:",
		CreatedAt: time.Now(),
	}

	data, err := rule.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}