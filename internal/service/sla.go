package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SLAService manages Service Level Agreements
type SLAService struct {
	mu         sync.RWMutex
	agreements map[string]*SLAAgreement
	metrics    map[string]*SLAMetric
	breaches   []SLABreach
}

// SLAAgreement represents an SLA agreement
type SLAAgreement struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	TenantID     string        `json:"tenant_id"`
	Description  string        `json:"description"`
	Type         SLAType       `json:"type"`
	Target       float64       `json:"target"` // e.g., 99.9 for 99.9% uptime
	Unit         string        `json:"unit"`   // percentage, seconds, count
	Measurement  string        `json:"measurement"`
	Period       SLAPeriod     `json:"period"`
	Current      float64       `json:"current"`
	Status       SLAStatus     `json:"status"`
	WarningThreshold float64   `json:"warning_threshold"`
	BreachThreshold  float64   `json:"breach_threshold"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// SLAMetric represents an SLA metric
type SLAMetric struct {
	ID           string    `json:"id"`
	AgreementID  string    `json:"agreement_id"`
	Timestamp    time.Time `json:"timestamp"`
	Value        float64   `json:"value"`
	Total        float64   `json:"total"`
	Success      float64   `json:"success"`
	Failure      float64   `json:"failure"`
	Latency      float64   `json:"latency_ms"`
	Availability float64   `json:"availability"`
}

// SLABreach represents an SLA breach
type SLABreach struct {
	ID           string    `json:"id"`
	AgreementID  string    `json:"agreement_id"`
	Type         string    `json:"type"` // warning, breach
	CurrentValue float64   `json:"current_value"`
	Threshold    float64   `json:"threshold"`
	Timestamp    time.Time `json:"timestamp"`
	Resolved     bool      `json:"resolved"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	Acknowledged bool      `json:"acknowledged"`
	AcknowledgedBy string  `json:"acknowledged_by,omitempty"`
	Message      string    `json:"message"`
}

// SLAType represents SLA type
type SLAType string

const (
	SLATypeAvailability SLAType = "availability"
	SLATypeLatency      SLAType = "latency"
	SLATypeThroughput   SLAType = "throughput"
	SLATypeErrorRate    SLAType = "error_rate"
	SLATypeCustom       SLAType = "custom"
)

// SLAPeriod represents SLA measurement period
type SLAPeriod string

const (
	SLAPeriodHourly  SLAPeriod = "hourly"
	SLAPeriodDaily   SLAPeriod = "daily"
	SLAPeriodWeekly  SLAPeriod = "weekly"
	SLAPeriodMonthly SLAPeriod = "monthly"
	SLAPeriodYearly  SLAPeriod = "yearly"
)

// SLAStatus represents SLA status
type SLAStatus string

const (
	SLAStatusHealthy  SLAStatus = "healthy"
	SLAStatusWarning  SLAStatus = "warning"
	SLAStatusBreached SLAStatus = "breached"
)

// NewSLAService creates a new SLA service
func NewSLAService() *SLAService {
	return &SLAService{
		agreements: make(map[string]*SLAAgreement),
		metrics:    make(map[string]*SLAMetric),
		breaches:   make([]SLABreach, 0),
	}
}

// CreateAgreement creates an SLA agreement
func (s *SLAService) CreateAgreement(ctx context.Context, agreement *SLAAgreement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agreement.Name == "" {
		return fmt.Errorf("name is required")
	}
	if agreement.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	now := time.Now()
	if agreement.ID == "" {
		agreement.ID = generateID()
	}
	agreement.CreatedAt = now
	agreement.UpdatedAt = now
	agreement.Status = SLAStatusHealthy
	agreement.Current = 100.0

	if agreement.StartTime.IsZero() {
		agreement.StartTime = now
	}

	s.agreements[agreement.ID] = agreement

	return nil
}

// GetAgreement gets an SLA agreement
func (s *SLAService) GetAgreement(id string) (*SLAAgreement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agreement, ok := s.agreements[id]
	if !ok {
		return nil, fmt.Errorf("SLA agreement not found: %s", id)
	}
	return agreement, nil
}

// ListAgreements lists SLA agreements for a tenant
func (s *SLAService) ListAgreements(tenantID string) []*SLAAgreement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*SLAAgreement
	for _, agreement := range s.agreements {
		if tenantID == "" || agreement.TenantID == tenantID {
			results = append(results, agreement)
		}
	}
	return results
}

// UpdateAgreement updates an SLA agreement
func (s *SLAService) UpdateAgreement(id string, updates map[string]interface{}) (*SLAAgreement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agreement, ok := s.agreements[id]
	if !ok {
		return nil, fmt.Errorf("SLA agreement not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		agreement.Name = name
	}
	if target, ok := updates["target"].(float64); ok {
		agreement.Target = target
	}
	if warning, ok := updates["warning_threshold"].(float64); ok {
		agreement.WarningThreshold = warning
	}
	if breach, ok := updates["breach_threshold"].(float64); ok {
		agreement.BreachThreshold = breach
	}

	agreement.UpdatedAt = time.Now()

	return agreement, nil
}

// DeleteAgreement deletes an SLA agreement
func (s *SLAService) DeleteAgreement(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.agreements[id]; !ok {
		return fmt.Errorf("SLA agreement not found: %s", id)
	}

	delete(s.agreements, id)

	return nil
}

// RecordMetric records an SLA metric
func (s *SLAService) RecordMetric(ctx context.Context, metric *SLAMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agreement, ok := s.agreements[metric.AgreementID]
	if !ok {
		return fmt.Errorf("SLA agreement not found: %s", metric.AgreementID)
	}

	if metric.ID == "" {
		metric.ID = generateID()
	}
	metric.Timestamp = time.Now()

	// Calculate availability if not provided
	if metric.Availability == 0 && metric.Total > 0 {
		metric.Availability = (metric.Success / metric.Total) * 100
	}

	s.metrics[metric.ID] = metric

	// Update agreement current value
	agreement.Current = metric.Availability
	agreement.UpdatedAt = time.Now()

	// Check thresholds
	s.checkThresholds(agreement)

	return nil
}

// checkThresholds checks SLA thresholds and creates breaches
func (s *SLAService) checkThresholds(agreement *SLAAgreement) {
	now := time.Now()

	// Check breach threshold
	if agreement.BreachThreshold > 0 && agreement.Current < agreement.BreachThreshold {
		agreement.Status = SLAStatusBreached
		s.breaches = append(s.breaches, SLABreach{
			ID:           generateID(),
			AgreementID:  agreement.ID,
			Type:         "breach",
			CurrentValue: agreement.Current,
			Threshold:    agreement.BreachThreshold,
			Timestamp:    now,
			Message:      fmt.Sprintf("SLA breached: %.2f%% < %.2f%%", agreement.Current, agreement.BreachThreshold),
		})
		return
	}

	// Check warning threshold
	if agreement.WarningThreshold > 0 && agreement.Current < agreement.WarningThreshold {
		agreement.Status = SLAStatusWarning
		s.breaches = append(s.breaches, SLABreach{
			ID:           generateID(),
			AgreementID:  agreement.ID,
			Type:         "warning",
			CurrentValue: agreement.Current,
			Threshold:    agreement.WarningThreshold,
			Timestamp:    now,
			Message:      fmt.Sprintf("SLA warning: %.2f%% < %.2f%%", agreement.Current, agreement.WarningThreshold),
		})
		return
	}

	agreement.Status = SLAStatusHealthy
}

// GetMetrics gets metrics for an agreement
func (s *SLAService) GetMetrics(agreementID string, start, end time.Time) []*SLAMetric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*SLAMetric
	for _, metric := range s.metrics {
		if metric.AgreementID == agreementID {
			if start.IsZero() || metric.Timestamp.After(start) {
				if end.IsZero() || metric.Timestamp.Before(end) {
					results = append(results, metric)
				}
			}
		}
	}
	return results
}

// GetBreaches gets SLA breaches
func (s *SLAService) GetBreaches(agreementID string) []SLABreach {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []SLABreach
	for _, breach := range s.breaches {
		if agreementID == "" || breach.AgreementID == agreementID {
			results = append(results, breach)
		}
	}
	return results
}

// AcknowledgeBreach acknowledges a breach
func (s *SLAService) AcknowledgeBreach(breachID, acknowledgedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.breaches {
		if s.breaches[i].ID == breachID {
			s.breaches[i].Acknowledged = true
			s.breaches[i].AcknowledgedBy = acknowledgedBy
			return nil
		}
	}
	return fmt.Errorf("breach not found: %s", breachID)
}

// ResolveBreach resolves a breach
func (s *SLAService) ResolveBreach(breachID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := range s.breaches {
		if s.breaches[i].ID == breachID {
			s.breaches[i].Resolved = true
			s.breaches[i].ResolvedAt = &now
			return nil
		}
	}
	return fmt.Errorf("breach not found: %s", breachID)
}

// GetReport generates an SLA report
func (s *SLAService) GetReport(agreementID string, period SLAPeriod) *SLAReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agreement, ok := s.agreements[agreementID]
	if !ok {
		return nil
	}

	report := &SLAReport{
		AgreementID: agreementID,
		Name:        agreement.Name,
		Target:      agreement.Target,
		Current:     agreement.Current,
		Status:      string(agreement.Status),
		Period:      string(period),
		GeneratedAt: time.Now(),
	}

	// Count breaches
	for _, breach := range s.breaches {
		if breach.AgreementID == agreementID {
			if breach.Type == "warning" {
				report.WarningCount++
			} else {
				report.BreachCount++
			}
		}
	}

	// Calculate uptime from metrics
	var totalUptime, totalChecks float64
	for _, metric := range s.metrics {
		if metric.AgreementID == agreementID {
			totalUptime += metric.Availability
			totalChecks++
		}
	}
	if totalChecks > 0 {
		report.AverageUptime = totalUptime / totalChecks
	}

	report.MeetsTarget = report.Current >= agreement.Target

	return report
}

// SLAReport represents an SLA report
type SLAReport struct {
	AgreementID  string    `json:"agreement_id"`
	Name         string    `json:"name"`
	Target       float64   `json:"target"`
	Current      float64   `json:"current"`
	AverageUptime float64  `json:"average_uptime"`
	Status       string    `json:"status"`
	Period       string    `json:"period"`
	MeetsTarget  bool      `json:"meets_target"`
	WarningCount int       `json:"warning_count"`
	BreachCount  int       `json:"breach_count"`
	GeneratedAt  time.Time `json:"generated_at"`
}

// GetDashboard gets SLA dashboard data
func (s *SLAService) GetDashboard(tenantID string) *SLADashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dashboard := &SLADashboard{
		TotalAgreements: len(s.agreements),
		Healthy:        0,
		Warning:        0,
		Breached:       0,
	}

	for _, agreement := range s.agreements {
		if tenantID != "" && agreement.TenantID != tenantID {
			continue
		}

		switch agreement.Status {
		case SLAStatusHealthy:
			dashboard.Healthy++
		case SLAStatusWarning:
			dashboard.Warning++
		case SLAStatusBreached:
			dashboard.Breached++
		}
	}

	// Get recent breaches
	var recentBreaches []SLABreach
	for i := len(s.breaches) - 1; i >= 0 && len(recentBreaches) < 10; i-- {
		if tenantID == "" || s.agreements[s.breaches[i].AgreementID].TenantID == tenantID {
			recentBreaches = append(recentBreaches, s.breaches[i])
		}
	}
	dashboard.RecentBreaches = recentBreaches

	return dashboard
}

// SLADashboard represents SLA dashboard data
type SLADashboard struct {
	TotalAgreements int         `json:"total_agreements"`
	Healthy         int         `json:"healthy"`
	Warning         int         `json:"warning"`
	Breached        int         `json:"breached"`
	RecentBreaches  []SLABreach `json:"recent_breaches"`
}

// ToJSON serializes agreement to JSON
func (a *SLAAgreement) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// ToJSON serializes report to JSON
func (r *SLAReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}