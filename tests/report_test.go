package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Report Service Tests

func TestReportService_New(t *testing.T) {
	rs := service.NewReportService()
	if rs == nil {
		t.Fatal("Expected non-nil report service")
	}
}

func TestReportService_CreateReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Monthly Usage",
		Type:     service.ReportTypeUsage,
		Format:   service.ReportFormatJSON,
		TenantID: "tenant-1",
	}

	err := rs.CreateReport(context.Background(), report)
	if err != nil {
		t.Fatalf("Failed to create report: %v", err)
	}

	if report.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if report.Status != service.ReportStatusPending {
		t.Errorf("Expected status pending, got %s", report.Status)
	}
}

func TestReportService_CreateReport_MissingName(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Type:     service.ReportTypeUsage,
		TenantID: "tenant-1",
	}

	err := rs.CreateReport(context.Background(), report)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestReportService_CreateReport_MissingTenant(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name: "Report",
		Type: service.ReportTypeUsage,
	}

	err := rs.CreateReport(context.Background(), report)
	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestReportService_GetReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Get Test",
		Type:     service.ReportTypeCost,
		TenantID: "tenant-get",
	}
	rs.CreateReport(context.Background(), report)

	retrieved, err := rs.GetReport(report.ID)
	if err != nil {
		t.Fatalf("Failed to get report: %v", err)
	}

	if retrieved.Name != "Get Test" {
		t.Error("Report name mismatch")
	}
}

func TestReportService_GetReport_NotFound(t *testing.T) {
	rs := service.NewReportService()

	_, err := rs.GetReport("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent report")
	}
}

func TestReportService_StartReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Start Test",
		Type:     service.ReportTypePerformance,
		TenantID: "tenant-start",
	}
	rs.CreateReport(context.Background(), report)

	err := rs.StartReport(report.ID)
	if err != nil {
		t.Fatalf("Failed to start report: %v", err)
	}

	retrieved, _ := rs.GetReport(report.ID)
	if retrieved.Status != service.ReportStatusRunning {
		t.Error("Expected status to be running")
	}
}

func TestReportService_CompleteReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Complete Test",
		Type:     service.ReportTypeAudit,
		TenantID: "tenant-complete",
	}
	rs.CreateReport(context.Background(), report)
	rs.StartReport(report.ID)

	data := map[string]interface{}{"total": 100}
	summary := &service.ReportSummary{
		TotalRecords: 100,
	}

	err := rs.CompleteReport(report.ID, data, summary)
	if err != nil {
		t.Fatalf("Failed to complete report: %v", err)
	}

	retrieved, _ := rs.GetReport(report.ID)
	if retrieved.Status != service.ReportStatusCompleted {
		t.Error("Expected status to be completed")
	}

	if retrieved.Summary.TotalRecords != 100 {
		t.Error("Summary not set correctly")
	}
}

func TestReportService_FailReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Fail Test",
		Type:     service.ReportTypeSecurity,
		TenantID: "tenant-fail",
	}
	rs.CreateReport(context.Background(), report)
	rs.StartReport(report.ID)

	err := rs.FailReport(report.ID, "database connection failed")
	if err != nil {
		t.Fatalf("Failed to fail report: %v", err)
	}

	retrieved, _ := rs.GetReport(report.ID)
	if retrieved.Status != service.ReportStatusFailed {
		t.Error("Expected status to be failed")
	}

	if retrieved.Error != "database connection failed" {
		t.Errorf("Expected error message, got %s", retrieved.Error)
	}
}

func TestReportService_ListReports(t *testing.T) {
	rs := service.NewReportService()

	rs.CreateReport(context.Background(), &service.Report{
		Name:     "List 1",
		Type:     service.ReportTypeUsage,
		TenantID: "tenant-list",
	})

	rs.CreateReport(context.Background(), &service.Report{
		Name:     "List 2",
		Type:     service.ReportTypeCost,
		TenantID: "tenant-list",
	})

	rs.CreateReport(context.Background(), &service.Report{
		Name:     "Other",
		Type:     service.ReportTypeUsage,
		TenantID: "other-tenant",
	})

	reports := rs.ListReports("tenant-list", "")
	if len(reports) < 2 {
		t.Errorf("Expected at least 2 reports, got %d", len(reports))
	}

	usageReports := rs.ListReports("tenant-list", service.ReportTypeUsage)
	for _, r := range usageReports {
		if r.TenantID != "tenant-list" || r.Type != service.ReportTypeUsage {
			t.Error("Filter not working correctly")
		}
	}
}

func TestReportService_DeleteReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Delete Test",
		Type:     service.ReportTypeSummary,
		TenantID: "tenant-delete",
	}
	rs.CreateReport(context.Background(), report)

	err := rs.DeleteReport(report.ID)
	if err != nil {
		t.Fatalf("Failed to delete report: %v", err)
	}

	_, err = rs.GetReport(report.ID)
	if err == nil {
		t.Error("Expected error for deleted report")
	}
}

func TestReportService_CreateTemplate(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name:          "Usage Template",
		Type:          service.ReportTypeUsage,
		Description:   "Monthly usage report",
		Definition:    "{\"fields\": [\"api_calls\", \"tokens\"]}",
		DefaultFormat: service.ReportFormatJSON,
	}

	err := rs.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestReportService_CreateTemplate_MissingName(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Type: service.ReportTypeCost,
	}

	err := rs.CreateTemplate(template)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestReportService_GetTemplate(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name: "Get Template",
		Type: service.ReportTypePerformance,
	}
	rs.CreateTemplate(template)

	retrieved, err := rs.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.Name != "Get Template" {
		t.Error("Template name mismatch")
	}
}

func TestReportService_GetTemplate_NotFound(t *testing.T) {
	rs := service.NewReportService()

	_, err := rs.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
}

func TestReportService_ListTemplates(t *testing.T) {
	rs := service.NewReportService()

	rs.CreateTemplate(&service.ReportTemplate{
		Name: "Template 1",
		Type: service.ReportTypeUsage,
	})

	rs.CreateTemplate(&service.ReportTemplate{
		Name: "Template 2",
		Type: service.ReportTypeCost,
	})

	templates := rs.ListTemplates("")
	if len(templates) < 2 {
		t.Errorf("Expected at least 2 templates, got %d", len(templates))
	}
}

func TestReportService_DeleteTemplate(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name: "Delete Template",
		Type: service.ReportTypeAudit,
	}
	rs.CreateTemplate(template)

	err := rs.DeleteTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	_, err = rs.GetTemplate(template.ID)
	if err == nil {
		t.Error("Expected error for deleted template")
	}
}

func TestReportService_CreateSchedule(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name: "Scheduled Report",
		Type: service.ReportTypeUsage,
	}
	rs.CreateTemplate(template)

	schedule := &service.ReportSchedule{
		Name:       "Daily Usage Report",
		TemplateID: template.ID,
		Cron:       "0 6 * * *",
		Format:     service.ReportFormatJSON,
	}

	err := rs.CreateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	if schedule.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if !schedule.Enabled {
		t.Error("Expected schedule to be enabled by default")
	}
}

func TestReportService_CreateSchedule_MissingCron(t *testing.T) {
	rs := service.NewReportService()

	schedule := &service.ReportSchedule{
		Name: "No Cron",
	}

	err := rs.CreateSchedule(schedule)
	if err == nil {
		t.Error("Expected error for missing cron")
	}
}

func TestReportService_GetSchedule(t *testing.T) {
	rs := service.NewReportService()

	schedule := &service.ReportSchedule{
		Name: "Get Schedule",
		Cron: "0 6 * * *",
	}
	rs.CreateSchedule(schedule)

	retrieved, err := rs.GetSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to get schedule: %v", err)
	}

	if retrieved.Name != "Get Schedule" {
		t.Error("Schedule name mismatch")
	}
}

func TestReportService_GetSchedule_NotFound(t *testing.T) {
	rs := service.NewReportService()

	_, err := rs.GetSchedule("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent schedule")
	}
}

func TestReportService_ListSchedules(t *testing.T) {
	rs := service.NewReportService()

	rs.CreateSchedule(&service.ReportSchedule{
		Name: "Schedule 1",
		Cron: "0 6 * * *",
	})

	rs.CreateSchedule(&service.ReportSchedule{
		Name: "Schedule 2",
		Cron: "0 18 * * *",
	})

	schedules := rs.ListSchedules()
	if len(schedules) < 2 {
		t.Errorf("Expected at least 2 schedules, got %d", len(schedules))
	}
}

func TestReportService_EnableDisableSchedule(t *testing.T) {
	rs := service.NewReportService()

	schedule := &service.ReportSchedule{
		Name: "Toggle Schedule",
		Cron: "0 6 * * *",
	}
	rs.CreateSchedule(schedule)

	err := rs.DisableSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to disable schedule: %v", err)
	}

	retrieved, _ := rs.GetSchedule(schedule.ID)
	if retrieved.Enabled {
		t.Error("Expected schedule to be disabled")
	}

	err = rs.EnableSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to enable schedule: %v", err)
	}

	retrieved, _ = rs.GetSchedule(schedule.ID)
	if !retrieved.Enabled {
		t.Error("Expected schedule to be enabled")
	}
}

func TestReportService_DeleteSchedule(t *testing.T) {
	rs := service.NewReportService()

	schedule := &service.ReportSchedule{
		Name: "Delete Schedule",
		Cron: "0 6 * * *",
	}
	rs.CreateSchedule(schedule)

	err := rs.DeleteSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to delete schedule: %v", err)
	}

	_, err = rs.GetSchedule(schedule.ID)
	if err == nil {
		t.Error("Expected error for deleted schedule")
	}
}

func TestReportService_GenerateFromTemplate(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name: "Generated Report",
		Type: service.ReportTypeCost,
		Parameters: []service.ReportParameter{
			{Name: "period", Type: "string", Default: "monthly"},
		},
		DefaultFormat: service.ReportFormatCSV,
	}
	rs.CreateTemplate(template)

	report, err := rs.GenerateFromTemplate(context.Background(), template.ID, "tenant-1", "user-1", map[string]interface{}{
		"region": "us-west",
	})

	if err != nil {
		t.Fatalf("Failed to generate from template: %v", err)
	}

	if report.Name != "Generated Report" {
		t.Error("Report name should match template name")
	}

	if report.TemplateID != template.ID {
		t.Error("Template ID not set")
	}

	// Check parameter merging
	if report.Parameters["period"] != "monthly" {
		t.Error("Default parameter not set")
	}

	if report.Parameters["region"] != "us-west" {
		t.Error("Custom parameter not set")
	}
}

func TestReportService_GenerateFromTemplate_TemplateNotFound(t *testing.T) {
	rs := service.NewReportService()

	_, err := rs.GenerateFromTemplate(context.Background(), "nonexistent", "tenant-1", "user-1", nil)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
}

func TestReportService_GetReportStats(t *testing.T) {
	rs := service.NewReportService()

	rs.CreateReport(context.Background(), &service.Report{
		Name:     "Stats 1",
		Type:     service.ReportTypeUsage,
		TenantID: "tenant-stats",
	})

	rs.CreateReport(context.Background(), &service.Report{
		Name:     "Stats 2",
		Type:     service.ReportTypeCost,
		TenantID: "tenant-stats",
	})

	rs.CreateTemplate(&service.ReportTemplate{
		Name: "Stats Template",
		Type: service.ReportTypeUsage,
	})

	rs.CreateSchedule(&service.ReportSchedule{
		Name: "Stats Schedule",
		Cron: "0 6 * * *",
	})

	stats := rs.GetReportStats("tenant-stats")

	if stats.TotalReports < 2 {
		t.Errorf("Expected at least 2 reports, got %d", stats.TotalReports)
	}

	if stats.TotalTemplates < 1 {
		t.Errorf("Expected at least 1 template, got %d", stats.TotalTemplates)
	}

	if stats.TotalSchedules < 1 {
		t.Errorf("Expected at least 1 schedule, got %d", stats.TotalSchedules)
	}
}

func TestReportService_ExportReport(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Export Test",
		Type:     service.ReportTypeUsage,
		TenantID: "tenant-export",
	}
	rs.CreateReport(context.Background(), report)
	rs.StartReport(report.ID)
	rs.CompleteReport(report.ID, map[string]interface{}{"data": "test"}, nil)

	data, err := rs.ExportReport(report.ID, service.ReportFormatJSON)
	if err != nil {
		t.Fatalf("Failed to export report: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty export data")
	}
}

func TestReportService_ExportReport_NotCompleted(t *testing.T) {
	rs := service.NewReportService()

	report := &service.Report{
		Name:     "Not Completed",
		Type:     service.ReportTypeUsage,
		TenantID: "tenant-export",
	}
	rs.CreateReport(context.Background(), report)

	_, err := rs.ExportReport(report.ID, service.ReportFormatJSON)
	if err == nil {
		t.Error("Expected error for non-completed report")
	}
}

func TestReportService_ReportTypes(t *testing.T) {
	types := []service.ReportType{
		service.ReportTypeUsage,
		service.ReportTypeCost,
		service.ReportTypePerformance,
		service.ReportTypeSecurity,
		service.ReportTypeAudit,
		service.ReportTypeSummary,
		service.ReportTypeCustom,
	}

	rs := service.NewReportService()

	for _, reportType := range types {
		report := &service.Report{
			Name:     string(reportType),
			Type:     reportType,
			TenantID: "tenant-types",
		}

		err := rs.CreateReport(context.Background(), report)
		if err != nil {
			t.Errorf("Failed to create %s report: %v", reportType, err)
		}
	}
}

func TestReportService_ReportFormats(t *testing.T) {
	formats := []service.ReportFormat{
		service.ReportFormatJSON,
		service.ReportFormatCSV,
		service.ReportFormatPDF,
		service.ReportFormatHTML,
		service.ReportFormatExcel,
	}

	for _, format := range formats {
		if string(format) == "" {
			t.Errorf("Format %s is empty", format)
		}
	}
}

func TestReportService_ReportStatuses(t *testing.T) {
	statuses := []service.ReportStatus{
		service.ReportStatusPending,
		service.ReportStatusRunning,
		service.ReportStatusCompleted,
		service.ReportStatusFailed,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status %s is empty", status)
		}
	}
}

func TestReportService_ReportToJSON(t *testing.T) {
	report := &service.Report{
		ID:        "report-1",
		Name:      "Test Report",
		Type:      service.ReportTypeUsage,
		TenantID:  "tenant-1",
		Status:    service.ReportStatusCompleted,
		CreatedAt: time.Now(),
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestReportService_TemplateToJSON(t *testing.T) {
	template := &service.ReportTemplate{
		ID:        "template-1",
		Name:      "Test Template",
		Type:      service.ReportTypeCost,
		CreatedAt: time.Now(),
	}

	data, err := template.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestReportService_ReportSummary(t *testing.T) {
	summary := &service.ReportSummary{
		TotalRecords: 1000,
		KeyMetrics: map[string]interface{}{
			"total_cost":    1500.50,
			"total_calls":   50000,
			"avg_duration":  250,
		},
		TopItems: []service.ReportItem{
			{Name: "API Calls", Value: 50000, Count: 1000, Change: 15.5},
		},
		Changes: []service.ReportChange{
			{Field: "status", OldValue: "pending", NewValue: "completed", Timestamp: time.Now()},
		},
		Comparisons: map[string]float64{
			"month_over_month": 12.5,
			"year_over_year":   45.2,
		},
	}

	if summary.TotalRecords != 1000 {
		t.Error("Total records mismatch")
	}

	if len(summary.KeyMetrics) != 3 {
		t.Error("Key metrics count mismatch")
	}
}

func TestReportService_ReportParameters(t *testing.T) {
	rs := service.NewReportService()

	template := &service.ReportTemplate{
		Name: "Parameterized Report",
		Type: service.ReportTypeCustom,
		Parameters: []service.ReportParameter{
			{
				Name:        "start_date",
				Type:        "date",
				Required:    true,
				Description: "Report start date",
			},
			{
				Name:        "end_date",
				Type:        "date",
				Required:    true,
				Description: "Report end date",
			},
			{
				Name:        "format",
				Type:        "string",
				Required:    false,
				Default:     "detailed",
				Description: "Report format (summary or detailed)",
			},
		},
		DefaultFormat: service.ReportFormatJSON,
	}

	err := rs.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if len(template.Parameters) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(template.Parameters))
	}
}