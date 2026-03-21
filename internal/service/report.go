package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ReportService manages report generation
type ReportService struct {
	mu       sync.RWMutex
	reports  map[string]*Report
	templates map[string]*ReportTemplate
	schedules map[string]*ReportSchedule
}

// Report represents a generated report
type Report struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         ReportType    `json:"type"`
	Format       ReportFormat  `json:"format"`
	Status       ReportStatus  `json:"status"`
	TemplateID   string        `json:"template_id,omitempty"`
	TenantID     string        `json:"tenant_id"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	PeriodStart  time.Time     `json:"period_start"`
	PeriodEnd    time.Time     `json:"period_end"`
	Data         interface{}   `json:"data,omitempty"`
	Summary      *ReportSummary `json:"summary,omitempty"`
	StoragePath  string        `json:"storage_path,omitempty"`
	Size         int64         `json:"size"`
	CreatedAt    time.Time     `json:"created_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	Duration     int64         `json:"duration_ms"`
	Error        string        `json:"error,omitempty"`
	CreatedBy    string        `json:"created_by"`
	Tags         []string      `json:"tags,omitempty"`
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        ReportType `json:"type"`
	Description string    `json:"description"`
	Definition  string    `json:"definition"`
	Parameters  []ReportParameter `json:"parameters"`
	DefaultFormat ReportFormat `json:"default_format"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// ReportParameter represents a report parameter
type ReportParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
}

// ReportSchedule represents a scheduled report
type ReportSchedule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	TemplateID  string        `json:"template_id"`
	Cron        string        `json:"cron"`
	Format      ReportFormat  `json:"format"`
	Parameters  map[string]interface{} `json:"parameters"`
	Recipients  []string      `json:"recipients"`
	Enabled     bool          `json:"enabled"`
	LastRun     *time.Time    `json:"last_run,omitempty"`
	NextRun     *time.Time    `json:"next_run,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// ReportSummary represents a report summary
type ReportSummary struct {
	TotalRecords  int                    `json:"total_records"`
	KeyMetrics    map[string]interface{} `json:"key_metrics"`
	TopItems      []ReportItem           `json:"top_items,omitempty"`
	Changes       []ReportChange         `json:"changes,omitempty"`
	Comparisons   map[string]float64     `json:"comparisons,omitempty"`
}

// ReportItem represents an item in a report
type ReportItem struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	Count  int         `json:"count"`
	Change float64     `json:"change,omitempty"`
}

// ReportChange represents a change in report data
type ReportChange struct {
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// ReportType represents report type
type ReportType string

const (
	ReportTypeUsage       ReportType = "usage"
	ReportTypeCost        ReportType = "cost"
	ReportTypePerformance ReportType = "performance"
	ReportTypeSecurity    ReportType = "security"
	ReportTypeAudit       ReportType = "audit"
	ReportTypeSummary     ReportType = "summary"
	ReportTypeCustom      ReportType = "custom"
)

// ReportFormat represents report format
type ReportFormat string

const (
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatHTML  ReportFormat = "html"
	ReportFormatExcel ReportFormat = "excel"
)

// ReportStatus represents report status
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusRunning   ReportStatus = "running"
	ReportStatusCompleted ReportStatus = "completed"
	ReportStatusFailed    ReportStatus = "failed"
)

// NewReportService creates a new report service
func NewReportService() *ReportService {
	return &ReportService{
		reports:   make(map[string]*Report),
		templates: make(map[string]*ReportTemplate),
		schedules: make(map[string]*ReportSchedule),
	}
}

// CreateReport creates a new report
func (s *ReportService) CreateReport(ctx context.Context, report *Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if report.Name == "" {
		return fmt.Errorf("name is required")
	}
	if report.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	now := time.Now()
	if report.ID == "" {
		report.ID = generateID()
	}
	report.Status = ReportStatusPending
	report.CreatedAt = now

	if report.PeriodEnd.IsZero() {
		report.PeriodEnd = now
	}
	if report.PeriodStart.IsZero() {
		report.PeriodStart = now.AddDate(0, -1, 0) // Default 1 month
	}

	s.reports[report.ID] = report

	return nil
}

// GetReport gets a report by ID
func (s *ReportService) GetReport(id string) (*Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, ok := s.reports[id]
	if !ok {
		return nil, fmt.Errorf("report not found: %s", id)
	}
	return report, nil
}

// StartReport starts report generation
func (s *ReportService) StartReport(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, ok := s.reports[id]
	if !ok {
		return fmt.Errorf("report not found: %s", id)
	}

	report.Status = ReportStatusRunning
	return nil
}

// CompleteReport completes a report
func (s *ReportService) CompleteReport(id string, data interface{}, summary *ReportSummary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, ok := s.reports[id]
	if !ok {
		return fmt.Errorf("report not found: %s", id)
	}

	now := time.Now()
	report.Status = ReportStatusCompleted
	report.Data = data
	report.Summary = summary
	report.CompletedAt = &now
	report.Duration = now.Sub(report.CreatedAt).Milliseconds()

	return nil
}

// FailReport marks a report as failed
func (s *ReportService) FailReport(id string, err string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, ok := s.reports[id]
	if !ok {
		return fmt.Errorf("report not found: %s", id)
	}

	now := time.Now()
	report.Status = ReportStatusFailed
	report.Error = err
	report.CompletedAt = &now
	report.Duration = now.Sub(report.CreatedAt).Milliseconds()

	return nil
}

// ListReports lists reports
func (s *ReportService) ListReports(tenantID string, reportType ReportType) []*Report {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Report
	for _, report := range s.reports {
		if tenantID != "" && report.TenantID != tenantID {
			continue
		}
		if reportType != "" && report.Type != reportType {
			continue
		}
		results = append(results, report)
	}
	return results
}

// DeleteReport deletes a report
func (s *ReportService) DeleteReport(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.reports[id]; !ok {
		return fmt.Errorf("report not found: %s", id)
	}

	delete(s.reports, id)
	return nil
}

// CreateTemplate creates a report template
func (s *ReportService) CreateTemplate(template *ReportTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if template.Name == "" {
		return fmt.Errorf("name is required")
	}

	if template.ID == "" {
		template.ID = generateID()
	}
	template.CreatedAt = time.Now()

	s.templates[template.ID] = template

	return nil
}

// GetTemplate gets a report template
func (s *ReportService) GetTemplate(id string) (*ReportTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	template, ok := s.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates lists report templates
func (s *ReportService) ListTemplates(reportType ReportType) []*ReportTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ReportTemplate
	for _, template := range s.templates {
		if reportType == "" || template.Type == reportType {
			results = append(results, template)
		}
	}
	return results
}

// DeleteTemplate deletes a report template
func (s *ReportService) DeleteTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.templates[id]; !ok {
		return fmt.Errorf("template not found: %s", id)
	}

	delete(s.templates, id)
	return nil
}

// CreateSchedule creates a report schedule
func (s *ReportService) CreateSchedule(schedule *ReportSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if schedule.Name == "" {
		return fmt.Errorf("name is required")
	}
	if schedule.Cron == "" {
		return fmt.Errorf("cron is required")
	}

	if schedule.ID == "" {
		schedule.ID = generateID()
	}
	schedule.CreatedAt = time.Now()
	schedule.Enabled = true

	s.schedules[schedule.ID] = schedule

	return nil
}

// GetSchedule gets a report schedule
func (s *ReportService) GetSchedule(id string) (*ReportSchedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	return schedule, nil
}

// ListSchedules lists report schedules
func (s *ReportService) ListSchedules() []*ReportSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ReportSchedule
	for _, schedule := range s.schedules {
		results = append(results, schedule)
	}
	return results
}

// EnableSchedule enables a schedule
func (s *ReportService) EnableSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	schedule.Enabled = true
	return nil
}

// DisableSchedule disables a schedule
func (s *ReportService) DisableSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	schedule.Enabled = false
	return nil
}

// DeleteSchedule deletes a report schedule
func (s *ReportService) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.schedules[id]; !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	delete(s.schedules, id)
	return nil
}

// GenerateFromTemplate generates a report from a template
func (s *ReportService) GenerateFromTemplate(ctx context.Context, templateID, tenantID, createdBy string, params map[string]interface{}) (*Report, error) {
	s.mu.RLock()
	template, ok := s.templates[templateID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Merge params with defaults
	mergedParams := make(map[string]interface{})
	for _, p := range template.Parameters {
		if p.Default != nil {
			mergedParams[p.Name] = p.Default
		}
	}
	for k, v := range params {
		mergedParams[k] = v
	}

	report := &Report{
		Name:       template.Name,
		Type:       template.Type,
		Format:     template.DefaultFormat,
		TemplateID: templateID,
		TenantID:   tenantID,
		Parameters: mergedParams,
		CreatedBy:  createdBy,
	}

	if err := s.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// GetReportStats gets report statistics
func (s *ReportService) GetReportStats(tenantID string) *ReportStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &ReportStats{
		ByType:   make(map[ReportType]int),
		ByStatus: make(map[ReportStatus]int),
	}

	for _, report := range s.reports {
		if tenantID != "" && report.TenantID != tenantID {
			continue
		}

		stats.TotalReports++
		stats.ByType[report.Type]++
		stats.ByStatus[report.Status]++
	}

	stats.TotalTemplates = len(s.templates)
	stats.TotalSchedules = len(s.schedules)

	return stats
}

// ReportStats represents report statistics
type ReportStats struct {
	TotalReports   int                 `json:"total_reports"`
	TotalTemplates int                 `json:"total_templates"`
	TotalSchedules int                 `json:"total_schedules"`
	ByType         map[ReportType]int  `json:"by_type"`
	ByStatus       map[ReportStatus]int `json:"by_status"`
}

// ExportReport exports a report to a specific format
func (s *ReportService) ExportReport(id string, format ReportFormat) ([]byte, error) {
	s.mu.RLock()
	report, ok := s.reports[id]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("report not found: %s", id)
	}

	if report.Status != ReportStatusCompleted {
		return nil, fmt.Errorf("report is not completed")
	}

	// Export based on format
	switch format {
	case ReportFormatJSON:
		return json.MarshalIndent(report.Data, "", "  ")
	default:
		return json.Marshal(report.Data)
	}
}

// ToJSON serializes report to JSON
func (r *Report) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes template to JSON
func (t *ReportTemplate) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}