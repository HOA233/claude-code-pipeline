package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CostService tracks API usage and costs
type CostService struct {
	mu        sync.RWMutex
	records   map[string]*CostRecord
	budgets   map[string]*Budget
	alerts    []CostAlert
	providers map[string]CostProvider
}

// CostRecord represents a cost tracking record
type CostRecord struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Provider    string    `json:"provider"` // anthropic, openai, etc.
	Model       string    `json:"model"`
	RequestID   string    `json:"request_id"`
	InputTokens int64     `json:"input_tokens"`
	OutputTokens int64    `json:"output_tokens"`
	TotalTokens int64     `json:"total_tokens"`
	Cost        float64   `json:"cost"`
	Currency    string    `json:"currency"`
	Timestamp   time.Time `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Budget represents a cost budget
type Budget struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Limit        float64   `json:"limit"`
	Used         float64   `json:"used"`
	Period       BudgetPeriod `json:"period"`
	AlertThresholds []float64 `json:"alert_thresholds"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	ResetAt      time.Time `json:"reset_at"`
}

// BudgetPeriod represents budget period
type BudgetPeriod string

const (
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

// CostAlert represents a cost alert
type CostAlert struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	BudgetID  string    `json:"budget_id"`
	Type      string    `json:"type"` // threshold_exceeded, budget_exhausted
	Threshold float64   `json:"threshold"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// CostProvider defines cost calculation interface
type CostProvider interface {
	Name() string
	CalculateCost(model string, inputTokens, outputTokens int64) float64
}

// AnthropicCostProvider implements CostProvider for Anthropic
type AnthropicCostProvider struct {
	prices map[string]ModelPrice
}

// ModelPrice represents model pricing
type ModelPrice struct {
	InputPrice  float64 // per 1M tokens
	OutputPrice float64 // per 1M tokens
}

// NewCostService creates a new cost service
func NewCostService() *CostService {
	cs := &CostService{
		records:   make(map[string]*CostRecord),
		budgets:   make(map[string]*Budget),
		alerts:    make([]CostAlert, 0),
		providers: make(map[string]CostProvider),
	}

	// Register default providers
	cs.RegisterProvider(&AnthropicCostProvider{
		prices: map[string]ModelPrice{
			"claude-opus-4-6":   {InputPrice: 15, OutputPrice: 75},
			"claude-sonnet-4-6": {InputPrice: 3, OutputPrice: 15},
			"claude-haiku-4-5":  {InputPrice: 0.8, OutputPrice: 4},
		},
	})

	return cs
}

// RegisterProvider registers a cost provider
func (s *CostService) RegisterProvider(provider CostProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[provider.Name()] = provider
}

// RecordUsage records a usage event
func (s *CostService) RecordUsage(ctx context.Context, record *CostRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.ID == "" {
		record.ID = generateID()
	}
	record.Timestamp = time.Now()

	// Calculate cost if not provided
	if record.Cost == 0 {
		if provider, ok := s.providers[record.Provider]; ok {
			record.Cost = provider.CalculateCost(record.Model, record.InputTokens, record.OutputTokens)
		}
	}
	record.TotalTokens = record.InputTokens + record.OutputTokens
	record.Currency = "USD"

	s.records[record.ID] = record

	// Update budget usage
	if budget, ok := s.budgets[record.TenantID]; ok {
		budget.Used += record.Cost
		s.checkBudgetAlerts(budget)
	}

	return nil
}

// checkBudgetAlerts checks and generates budget alerts
func (s *CostService) checkBudgetAlerts(budget *Budget) {
	usagePercent := (budget.Used / budget.Limit) * 100

	for _, threshold := range budget.AlertThresholds {
		if usagePercent >= threshold && !s.hasAlertForThreshold(budget.ID, threshold) {
			s.alerts = append(s.alerts, CostAlert{
				ID:        generateID(),
				TenantID:  budget.TenantID,
				BudgetID:  budget.ID,
				Type:      "threshold_exceeded",
				Threshold: threshold,
				Message:   fmt.Sprintf("Budget usage at %.1f%% (%.2f/%.2f)", usagePercent, budget.Used, budget.Limit),
				Timestamp: time.Now(),
			})
		}
	}

	if budget.Used >= budget.Limit {
		s.alerts = append(s.alerts, CostAlert{
			ID:        generateID(),
			TenantID:  budget.TenantID,
			BudgetID:  budget.ID,
			Type:      "budget_exhausted",
			Threshold: 100,
			Message:   fmt.Sprintf("Budget exhausted: %.2f/%.2f", budget.Used, budget.Limit),
			Timestamp: time.Now(),
		})
	}
}

// hasAlertForThreshold checks if an alert exists for a threshold
func (s *CostService) hasAlertForThreshold(budgetID string, threshold float64) bool {
	for _, alert := range s.alerts {
		if alert.BudgetID == budgetID && alert.Threshold == threshold {
			return true
		}
	}
	return false
}

// SetBudget sets a budget for a tenant
func (s *CostService) SetBudget(budget *Budget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if budget.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	if budget.ID == "" {
		budget.ID = generateID()
	}
	budget.CreatedAt = time.Now()
	budget.Used = 0

	// Calculate reset time based on period
	now := time.Now()
	switch budget.Period {
	case BudgetPeriodDaily:
		budget.ResetAt = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case BudgetPeriodWeekly:
		budget.ResetAt = now.AddDate(0, 0, 7-int(now.Weekday()))
	case BudgetPeriodMonthly:
		budget.ResetAt = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	}

	s.budgets[budget.TenantID] = budget
	return nil
}

// GetBudget gets a budget for a tenant
func (s *CostService) GetBudget(tenantID string) (*Budget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budget, ok := s.budgets[tenantID]
	if !ok {
		return nil, fmt.Errorf("budget not found for tenant: %s", tenantID)
	}
	return budget, nil
}

// GetUsage gets usage records for a tenant
func (s *CostService) GetUsage(tenantID string, start, end time.Time) []*CostRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*CostRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.Timestamp.After(start) && record.Timestamp.Before(end) {
			records = append(records, record)
		}
	}
	return records
}

// GetCostSummary gets a cost summary for a tenant
func (s *CostService) GetCostSummary(tenantID string, start, end time.Time) *CostSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := &CostSummary{
		TenantID: tenantID,
		ByModel:  make(map[string]ModelSummary),
		ByDay:    make(map[string]DaySummary),
	}

	for _, record := range s.records {
		if record.TenantID == tenantID && record.Timestamp.After(start) && record.Timestamp.Before(end) {
			summary.TotalRequests++
			summary.TotalCost += record.Cost
			summary.TotalInputTokens += record.InputTokens
			summary.TotalOutputTokens += record.OutputTokens

			// By model
			ms := summary.ByModel[record.Model]
			ms.Requests++
			ms.Cost += record.Cost
			ms.InputTokens += record.InputTokens
			ms.OutputTokens += record.OutputTokens
			summary.ByModel[record.Model] = ms

			// By day
			day := record.Timestamp.Format("2006-01-02")
			ds := summary.ByDay[day]
			ds.Requests++
			ds.Cost += record.Cost
			summary.ByDay[day] = ds
		}
	}

	return summary
}

// CostSummary represents a cost summary
type CostSummary struct {
	TenantID        string                    `json:"tenant_id"`
	TotalRequests   int64                     `json:"total_requests"`
	TotalCost       float64                   `json:"total_cost"`
	TotalInputTokens  int64                   `json:"total_input_tokens"`
	TotalOutputTokens int64                   `json:"total_output_tokens"`
	ByModel         map[string]ModelSummary   `json:"by_model"`
	ByDay           map[string]DaySummary     `json:"by_day"`
}

// ModelSummary represents model-level summary
type ModelSummary struct {
	Requests     int64   `json:"requests"`
	Cost         float64 `json:"cost"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
}

// DaySummary represents day-level summary
type DaySummary struct {
	Requests int64   `json:"requests"`
	Cost     float64 `json:"cost"`
}

// GetAlerts gets alerts for a tenant
func (s *CostService) GetAlerts(tenantID string) []CostAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []CostAlert
	for _, alert := range s.alerts {
		if alert.TenantID == tenantID {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// MarkAlertRead marks an alert as read
func (s *CostService) MarkAlertRead(alertID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == alertID {
			s.alerts[i].Read = true
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// ResetBudget resets a budget for a new period
func (s *CostService) ResetBudget(tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	budget, ok := s.budgets[tenantID]
	if !ok {
		return fmt.Errorf("budget not found for tenant: %s", tenantID)
	}

	budget.Used = 0
	budget.ResetAt = time.Now().AddDate(0, 0, 1) // Simplified, should use period

	return nil
}

// CheckAndResetBudgets checks and resets budgets that have expired
func (s *CostService) CheckAndResetBudgets() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	reset := 0

	for _, budget := range s.budgets {
		if now.After(budget.ResetAt) {
			budget.Used = 0
			switch budget.Period {
			case BudgetPeriodDaily:
				budget.ResetAt = now.AddDate(0, 0, 1)
			case BudgetPeriodWeekly:
				budget.ResetAt = now.AddDate(0, 0, 7)
			case BudgetPeriodMonthly:
				budget.ResetAt = now.AddDate(0, 1, 0)
			}
			reset++
		}
	}

	return reset
}

// Name returns the provider name
func (p *AnthropicCostProvider) Name() string {
	return "anthropic"
}

// CalculateCost calculates the cost for Anthropic models
func (p *AnthropicCostProvider) CalculateCost(model string, inputTokens, outputTokens int64) float64 {
	price, ok := p.prices[model]
	if !ok {
		// Default pricing
		price = ModelPrice{InputPrice: 3, OutputPrice: 15}
	}

	inputCost := float64(inputTokens) / 1_000_000 * price.InputPrice
	outputCost := float64(outputTokens) / 1_000_000 * price.OutputPrice

	return inputCost + outputCost
}

// ToJSON serializes cost record to JSON
func (r *CostRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes budget to JSON
func (b *Budget) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}