package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// SLA Service Tests

func TestSLAService_New(t *testing.T) {
	sla := service.NewSLAService()
	if sla == nil {
		t.Fatal("Expected non-nil SLA service")
	}
}

func TestSLAService_CreateAgreement(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Production Uptime",
		TenantID: "tenant-1",
		Type:     service.SLATypeAvailability,
		Target:   99.9,
	}

	err := sla.CreateAgreement(context.Background(), agreement)
	if err != nil {
		t.Fatalf("Failed to create agreement: %v", err)
	}

	if agreement.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if agreement.Status != service.SLAStatusHealthy {
		t.Errorf("Expected status healthy, got %s", agreement.Status)
	}
}

func TestSLAService_CreateAgreement_MissingName(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		TenantID: "tenant-1",
		Type:     service.SLATypeAvailability,
	}

	err := sla.CreateAgreement(context.Background(), agreement)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestSLAService_CreateAgreement_MissingTenant(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name: "SLA",
		Type: service.SLATypeAvailability,
	}

	err := sla.CreateAgreement(context.Background(), agreement)
	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestSLAService_GetAgreement(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Get Test",
		TenantID: "tenant-get",
		Type:     service.SLATypeLatency,
	}
	sla.CreateAgreement(context.Background(), agreement)

	retrieved, err := sla.GetAgreement(agreement.ID)
	if err != nil {
		t.Fatalf("Failed to get agreement: %v", err)
	}

	if retrieved.Name != "Get Test" {
		t.Error("Agreement name mismatch")
	}
}

func TestSLAService_GetAgreement_NotFound(t *testing.T) {
	sla := service.NewSLAService()

	_, err := sla.GetAgreement("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent agreement")
	}
}

func TestSLAService_ListAgreements(t *testing.T) {
	sla := service.NewSLAService()

	sla.CreateAgreement(context.Background(), &service.SLAAgreement{
		Name:     "SLA 1",
		TenantID: "tenant-list",
		Type:     service.SLATypeAvailability,
	})

	sla.CreateAgreement(context.Background(), &service.SLAAgreement{
		Name:     "SLA 2",
		TenantID: "tenant-list",
		Type:     service.SLATypeLatency,
	})

	agreements := sla.ListAgreements("tenant-list")
	if len(agreements) < 2 {
		t.Errorf("Expected at least 2 agreements, got %d", len(agreements))
	}
}

func TestSLAService_UpdateAgreement(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Original",
		TenantID: "tenant-update",
		Type:     service.SLATypeAvailability,
	}
	sla.CreateAgreement(context.Background(), agreement)

	updated, err := sla.UpdateAgreement(agreement.ID, map[string]interface{}{
		"name":              "Updated",
		"target":            99.99,
		"warning_threshold": 99.5,
	})

	if err != nil {
		t.Fatalf("Failed to update agreement: %v", err)
	}

	if updated.Name != "Updated" {
		t.Error("Name not updated")
	}
}

func TestSLAService_DeleteAgreement(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "To Delete",
		TenantID: "tenant-delete",
		Type:     service.SLATypeAvailability,
	}
	sla.CreateAgreement(context.Background(), agreement)

	err := sla.DeleteAgreement(agreement.ID)
	if err != nil {
		t.Fatalf("Failed to delete agreement: %v", err)
	}

	_, err = sla.GetAgreement(agreement.ID)
	if err == nil {
		t.Error("Expected error for deleted agreement")
	}
}

func TestSLAService_RecordMetric(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Metric Test",
		TenantID: "tenant-metric",
		Type:     service.SLATypeAvailability,
	}
	sla.CreateAgreement(context.Background(), agreement)

	metric := &service.SLAMetric{
		AgreementID: agreement.ID,
		Total:       1000,
		Success:     990,
		Failure:     10,
	}

	err := sla.RecordMetric(context.Background(), metric)
	if err != nil {
		t.Fatalf("Failed to record metric: %v", err)
	}

	if metric.Availability < 99.0 || metric.Availability > 99.1 {
		t.Errorf("Expected availability ~99%%, got %.2f%%", metric.Availability)
	}
}

func TestSLAService_RecordMetric_UpdatesCurrent(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Current Test",
		TenantID: "tenant-current",
		Type:     service.SLATypeAvailability,
	}
	sla.CreateAgreement(context.Background(), agreement)

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 95.0,
	})

	updated, _ := sla.GetAgreement(agreement.ID)
	if updated.Current != 95.0 {
		t.Errorf("Expected current 95.0, got %.2f", updated.Current)
	}
}

func TestSLAService_BreachDetection(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:            "Breach Test",
		TenantID:        "tenant-breach",
		Type:            service.SLATypeAvailability,
		Target:          99.9,
		BreachThreshold: 95.0,
	}
	sla.CreateAgreement(context.Background(), agreement)

	// Record metric that breaches threshold
	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 90.0,
	})

	updated, _ := sla.GetAgreement(agreement.ID)
	if updated.Status != service.SLAStatusBreached {
		t.Error("Expected status to be breached")
	}

	breaches := sla.GetBreaches(agreement.ID)
	if len(breaches) == 0 {
		t.Error("Expected breach record")
	}
}

func TestSLAService_WarningDetection(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:            "Warning Test",
		TenantID:        "tenant-warning",
		Type:            service.SLATypeAvailability,
		Target:          99.9,
		WarningThreshold: 98.0,
		BreachThreshold: 95.0,
	}
	sla.CreateAgreement(context.Background(), agreement)

	// Record metric that triggers warning
	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 97.0,
	})

	updated, _ := sla.GetAgreement(agreement.ID)
	if updated.Status != service.SLAStatusWarning {
		t.Error("Expected status to be warning")
	}
}

func TestSLAService_AcknowledgeBreach(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:            "Ack Test",
		TenantID:        "tenant-ack",
		Type:            service.SLATypeAvailability,
		BreachThreshold: 95.0,
	}
	sla.CreateAgreement(context.Background(), agreement)

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 90.0,
	})

	breaches := sla.GetBreaches(agreement.ID)
	if len(breaches) == 0 {
		t.Fatal("Expected breach")
	}

	err := sla.AcknowledgeBreach(breaches[0].ID, "admin")
	if err != nil {
		t.Fatalf("Failed to acknowledge breach: %v", err)
	}
}

func TestSLAService_ResolveBreach(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:            "Resolve Test",
		TenantID:        "tenant-resolve",
		Type:            service.SLATypeAvailability,
		BreachThreshold: 95.0,
	}
	sla.CreateAgreement(context.Background(), agreement)

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 90.0,
	})

	breaches := sla.GetBreaches(agreement.ID)

	err := sla.ResolveBreach(breaches[0].ID)
	if err != nil {
		t.Fatalf("Failed to resolve breach: %v", err)
	}
}

func TestSLAService_GetReport(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Report Test",
		TenantID: "tenant-report",
		Type:     service.SLATypeAvailability,
		Target:   99.9,
	}
	sla.CreateAgreement(context.Background(), agreement)

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 99.5,
	})

	report := sla.GetReport(agreement.ID, service.SLAPeriodDaily)
	if report == nil {
		t.Fatal("Expected report")
	}

	if report.Name != "Report Test" {
		t.Error("Report name mismatch")
	}
}

func TestSLAService_GetDashboard(t *testing.T) {
	sla := service.NewSLAService()

	sla.CreateAgreement(context.Background(), &service.SLAAgreement{
		Name:     "Dashboard 1",
		TenantID: "tenant-dashboard",
		Type:     service.SLATypeAvailability,
	})

	sla.CreateAgreement(context.Background(), &service.SLAAgreement{
		Name:     "Dashboard 2",
		TenantID: "tenant-dashboard",
		Type:     service.SLATypeLatency,
	})

	dashboard := sla.GetDashboard("tenant-dashboard")
	if dashboard == nil {
		t.Fatal("Expected dashboard")
	}

	if dashboard.TotalAgreements < 2 {
		t.Errorf("Expected at least 2 agreements, got %d", dashboard.TotalAgreements)
	}
}

func TestSLAService_GetMetrics(t *testing.T) {
	sla := service.NewSLAService()

	agreement := &service.SLAAgreement{
		Name:     "Metrics Test",
		TenantID: "tenant-metrics",
		Type:     service.SLATypeAvailability,
	}
	sla.CreateAgreement(context.Background(), agreement)

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 99.0,
	})

	sla.RecordMetric(context.Background(), &service.SLAMetric{
		AgreementID: agreement.ID,
		Availability: 99.5,
	})

	metrics := sla.GetMetrics(agreement.ID, time.Time{}, time.Time{})
	if len(metrics) < 2 {
		t.Errorf("Expected at least 2 metrics, got %d", len(metrics))
	}
}

func TestSLAService_SLATypes(t *testing.T) {
	types := []service.SLAType{
		service.SLATypeAvailability,
		service.SLATypeLatency,
		service.SLATypeThroughput,
		service.SLATypeErrorRate,
		service.SLATypeCustom,
	}

	sla := service.NewSLAService()

	for _, slaType := range types {
		agreement := &service.SLAAgreement{
			Name:     string(slaType),
			TenantID: "tenant-types",
			Type:     slaType,
		}

		err := sla.CreateAgreement(context.Background(), agreement)
		if err != nil {
			t.Errorf("Failed to create %s agreement: %v", slaType, err)
		}
	}
}

func TestSLAService_SLAPeriods(t *testing.T) {
	periods := []service.SLAPeriod{
		service.SLAPeriodHourly,
		service.SLAPeriodDaily,
		service.SLAPeriodWeekly,
		service.SLAPeriodMonthly,
		service.SLAPeriodYearly,
	}

	for _, period := range periods {
		if string(period) == "" {
			t.Errorf("Period %s is empty", period)
		}
	}
}

func TestSLAService_AgreementToJSON(t *testing.T) {
	agreement := &service.SLAAgreement{
		ID:       "sla-1",
		Name:     "Test SLA",
		TenantID: "tenant-1",
		Type:     service.SLATypeAvailability,
		Target:   99.9,
	}

	data, err := agreement.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}