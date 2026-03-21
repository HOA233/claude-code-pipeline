package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Cost Service Tests

func TestCostService_New(t *testing.T) {
	cs := service.NewCostService()
	if cs == nil {
		t.Fatal("Expected non-nil cost service")
	}
}

func TestCostService_RecordUsage(t *testing.T) {
	cs := service.NewCostService()

	record := &service.CostRecord{
		TenantID:     "tenant-1",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		RequestID:    "req-001",
		InputTokens:  1000,
		OutputTokens: 500,
	}

	err := cs.RecordUsage(context.Background(), record)
	if err != nil {
		t.Fatalf("Failed to record usage: %v", err)
	}

	if record.Cost == 0 {
		t.Error("Expected cost to be calculated")
	}

	if record.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestCostService_SetBudget(t *testing.T) {
	cs := service.NewCostService()

	budget := &service.Budget{
		TenantID: "tenant-1",
		Name:     "Monthly Budget",
		Limit:    100.0,
		Period:   service.BudgetPeriodMonthly,
		AlertThresholds: []float64{50, 80, 100},
	}

	err := cs.SetBudget(budget)
	if err != nil {
		t.Fatalf("Failed to set budget: %v", err)
	}

	if budget.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestCostService_GetBudget(t *testing.T) {
	cs := service.NewCostService()

	cs.SetBudget(&service.Budget{
		TenantID: "tenant-get",
		Name:     "Test Budget",
		Limit:    50.0,
		Period:   service.BudgetPeriodDaily,
	})

	budget, err := cs.GetBudget("tenant-get")
	if err != nil {
		t.Fatalf("Failed to get budget: %v", err)
	}

	if budget.Name != "Test Budget" {
		t.Error("Budget name mismatch")
	}
}

func TestCostService_GetBudget_NotFound(t *testing.T) {
	cs := service.NewCostService()

	_, err := cs.GetBudget("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent budget")
	}
}

func TestCostService_GetUsage(t *testing.T) {
	cs := service.NewCostService()

	// Record some usage
	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-usage",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  100,
		OutputTokens: 50,
	})

	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-usage",
		Provider:     "anthropic",
		Model:        "claude-opus-4-6",
		InputTokens:  200,
		OutputTokens: 100,
	})

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)

	records := cs.GetUsage("tenant-usage", start, end)
	if len(records) < 2 {
		t.Errorf("Expected at least 2 records, got %d", len(records))
	}
}

func TestCostService_GetCostSummary(t *testing.T) {
	cs := service.NewCostService()

	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-summary",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  1000,
		OutputTokens: 500,
	})

	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-summary",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  2000,
		OutputTokens: 1000,
	})

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)

	summary := cs.GetCostSummary("tenant-summary", start, end)

	if summary.TotalRequests < 2 {
		t.Errorf("Expected at least 2 requests, got %d", summary.TotalRequests)
	}

	if summary.TotalCost <= 0 {
		t.Error("Expected positive total cost")
	}

	if _, ok := summary.ByModel["claude-sonnet-4-6"]; !ok {
		t.Error("Expected model summary")
	}
}

func TestCostService_BudgetAlert(t *testing.T) {
	cs := service.NewCostService()

	cs.SetBudget(&service.Budget{
		TenantID:        "tenant-alert",
		Name:            "Alert Budget",
		Limit:           1.0, // Very low limit
		Period:          service.BudgetPeriodDaily,
		AlertThresholds: []float64{50, 100},
	})

	// Record usage that exceeds threshold
	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-alert",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  100000,
		OutputTokens: 50000,
	})

	alerts := cs.GetAlerts("tenant-alert")
	if len(alerts) == 0 {
		t.Error("Expected budget alerts")
	}
}

func TestCostService_MarkAlertRead(t *testing.T) {
	cs := service.NewCostService()

	cs.SetBudget(&service.Budget{
		TenantID:        "tenant-read",
		Name:            "Test Budget",
		Limit:           0.01,
		Period:          service.BudgetPeriodDaily,
		AlertThresholds: []float64{100},
	})

	cs.RecordUsage(context.Background(), &service.CostRecord{
		TenantID:     "tenant-read",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  10000,
		OutputTokens: 5000,
	})

	alerts := cs.GetAlerts("tenant-read")
	if len(alerts) == 0 {
		t.Fatal("Expected at least one alert")
	}

	err := cs.MarkAlertRead(alerts[0].ID)
	if err != nil {
		t.Fatalf("Failed to mark alert as read: %v", err)
	}
}

func TestCostService_AnthropicCostProvider(t *testing.T) {
	provider := &service.AnthropicCostProvider{
		Prices: map[string]service.ModelPrice{
			"claude-sonnet-4-6": {InputPrice: 3, OutputPrice: 15},
		},
	}

	if provider.Name() != "anthropic" {
		t.Error("Expected provider name to be anthropic")
	}

	cost := provider.CalculateCost("claude-sonnet-4-6", 1_000_000, 1_000_000)
	expected := 3.0 + 15.0 // $3 for input, $15 for output per 1M tokens

	if cost != expected {
		t.Errorf("Expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCostService_RecordToJSON(t *testing.T) {
	record := &service.CostRecord{
		ID:           "rec-1",
		TenantID:     "tenant-1",
		Provider:     "anthropic",
		Model:        "claude-sonnet-4-6",
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         0.015,
	}

	data, err := record.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestCostService_BudgetToJSON(t *testing.T) {
	budget := &service.Budget{
		ID:        "budget-1",
		TenantID:  "tenant-1",
		Name:      "Test",
		Limit:     100.0,
		Used:      25.0,
		Period:    service.BudgetPeriodMonthly,
	}

	data, err := budget.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestCostService_BudgetPeriods(t *testing.T) {
	cs := service.NewCostService()

	periods := []service.BudgetPeriod{
		service.BudgetPeriodDaily,
		service.BudgetPeriodWeekly,
		service.BudgetPeriodMonthly,
	}

	for _, period := range periods {
		err := cs.SetBudget(&service.Budget{
			TenantID: "tenant-" + string(period),
			Name:     string(period) + " Budget",
			Limit:    100.0,
			Period:   period,
		})

		if err != nil {
			t.Errorf("Failed to set %s budget: %v", period, err)
		}
	}
}