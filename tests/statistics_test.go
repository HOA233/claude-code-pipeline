package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Statistics Service Tests

func TestStatisticsService_New(t *testing.T) {
	ss := service.NewStatisticsService()
	if ss == nil {
		t.Fatal("Expected non-nil statistics service")
	}
}

func TestStatisticsService_RecordMetric(t *testing.T) {
	ss := service.NewStatisticsService()

	metric := &service.Metric{
		Name:      "api.latency",
		Value:     150.5,
		TenantID:  "tenant-1",
		Timestamp: time.Now(),
		Tags:      map[string]string{"endpoint": "/api/users"},
	}

	err := ss.RecordMetric(metric)
	if err != nil {
		t.Fatalf("Failed to record metric: %v", err)
	}
}

func TestStatisticsService_GetMetric(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.RecordMetric(&service.Metric{
		Name:     "test.metric",
		Value:    100,
		TenantID: "tenant-get",
	})

	metrics := ss.GetMetric("test.metric", "tenant-get")
	if len(metrics) == 0 {
		t.Error("Expected metrics")
	}
}

func TestStatisticsService_GetMetricStats(t *testing.T) {
	ss := service.NewStatisticsService()

	// Record multiple metrics
	for i := 0; i < 10; i++ {
		ss.RecordMetric(&service.Metric{
			Name:     "stats.metric",
			Value:    float64(i * 10),
			TenantID: "tenant-stats",
		})
	}

	stats := ss.GetMetricStats("stats.metric", "tenant-stats")

	if stats.Count != 10 {
		t.Errorf("Expected count 10, got %d", stats.Count)
	}

	if stats.Min != 0 {
		t.Errorf("Expected min 0, got %f", stats.Min)
	}

	if stats.Max != 90 {
		t.Errorf("Expected max 90, got %f", stats.Max)
	}

	if stats.Sum != 450 {
		t.Errorf("Expected sum 450, got %f", stats.Sum)
	}
}

func TestStatisticsService_AggregateMetrics(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.RecordMetric(&service.Metric{
		Name:     "agg.metric",
		Value:    10,
		TenantID: "tenant-agg",
		Tags:     map[string]string{"region": "us-east"},
	})

	ss.RecordMetric(&service.Metric{
		Name:     "agg.metric",
		Value:    20,
		TenantID: "tenant-agg",
		Tags:     map[string]string{"region": "us-west"},
	})

	aggregated := ss.AggregateMetrics("agg.metric", "tenant-agg", "region")

	if len(aggregated) != 2 {
		t.Errorf("Expected 2 aggregated values, got %d", len(aggregated))
	}
}

func TestStatisticsService_GetTimeSeries(t *testing.T) {
	ss := service.NewStatisticsService()

	now := time.Now()

	// Record metrics at different times
	for i := 0; i < 5; i++ {
		ss.RecordMetric(&service.Metric{
			Name:      "timeseries.metric",
			Value:     float64(i),
			TenantID:  "tenant-ts",
			Timestamp: now.Add(time.Duration(i) * time.Minute),
		})
	}

	start := now.Add(-time.Minute)
	end := now.Add(6 * time.Minute)

	series := ss.GetTimeSeries("timeseries.metric", "tenant-ts", start, end)

	if len(series) < 5 {
		t.Errorf("Expected at least 5 points, got %d", len(series))
	}
}

func TestStatisticsService_DeleteMetric(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.RecordMetric(&service.Metric{
		Name:     "delete.metric",
		Value:    100,
		TenantID: "tenant-delete",
	})

	err := ss.DeleteMetric("delete.metric", "tenant-delete")
	if err != nil {
		t.Fatalf("Failed to delete metric: %v", err)
	}

	metrics := ss.GetMetric("delete.metric", "tenant-delete")
	if len(metrics) != 0 {
		t.Error("Expected no metrics after deletion")
	}
}

func TestStatisticsService_CleanupOldMetrics(t *testing.T) {
	ss := service.NewStatisticsService()

	// Record old metric
	ss.RecordMetric(&service.Metric{
		Name:      "old.metric",
		Value:     100,
		TenantID:  "tenant-old",
		Timestamp: time.Now().Add(-48 * time.Hour),
	})

	// Record recent metric
	ss.RecordMetric(&service.Metric{
		Name:      "recent.metric",
		Value:     100,
		TenantID:  "tenant-old",
		Timestamp: time.Now(),
	})

	deleted := ss.CleanupOldMetrics(24 * time.Hour)
	if deleted < 1 {
		t.Error("Expected at least 1 metric to be cleaned up")
	}
}

func TestStatisticsService_Counter(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.IncrementCounter("counter.test", "tenant-counter", 1)
	ss.IncrementCounter("counter.test", "tenant-counter", 5)

	value := ss.GetCounter("counter.test", "tenant-counter")
	if value != 6 {
		t.Errorf("Expected counter value 6, got %d", value)
	}
}

func TestStatisticsService_Gauge(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.SetGauge("gauge.test", "tenant-gauge", 42.5)

	value := ss.GetGauge("gauge.test", "tenant-gauge")
	if value != 42.5 {
		t.Errorf("Expected gauge value 42.5, got %f", value)
	}

	// Update gauge
	ss.SetGauge("gauge.test", "tenant-gauge", 50.0)

	value = ss.GetGauge("gauge.test", "tenant-gauge")
	if value != 50.0 {
		t.Errorf("Expected updated gauge value 50.0, got %f", value)
	}
}

func TestStatisticsService_Histogram(t *testing.T) {
	ss := service.NewStatisticsService()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, v := range values {
		ss.RecordHistogram("histogram.test", "tenant-hist", v)
	}

	p50 := ss.GetPercentile("histogram.test", "tenant-hist", 50)
	p95 := ss.GetPercentile("histogram.test", "tenant-hist", 95)
	p99 := ss.GetPercentile("histogram.test", "tenant-hist", 99)

	if p50 < 4 || p50 > 6 {
		t.Errorf("Expected p50 around 5, got %f", p50)
	}

	if p95 < 9 || p95 > 10 {
		t.Errorf("Expected p95 around 9.5, got %f", p95)
	}

	_ = p99
}

func TestStatisticsService_ListMetrics(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.RecordMetric(&service.Metric{Name: "list.metric1", Value: 1, TenantID: "tenant-list"})
	ss.RecordMetric(&service.Metric{Name: "list.metric2", Value: 2, TenantID: "tenant-list"})

	names := ss.ListMetrics("tenant-list")
	if len(names) < 2 {
		t.Errorf("Expected at least 2 metrics, got %d", len(names))
	}
}

func TestStatisticsService_ExportMetrics(t *testing.T) {
	ss := service.NewStatisticsService()

	ss.RecordMetric(&service.Metric{
		Name:     "export.metric",
		Value:    100,
		TenantID: "tenant-export",
	})

	data, err := ss.ExportMetrics("tenant-export", "json")
	if err != nil {
		t.Fatalf("Failed to export metrics: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected exported data")
	}
}

func TestStatisticsService_MetricToJSON(t *testing.T) {
	metric := &service.Metric{
		Name:      "json.metric",
		Value:     123.45,
		TenantID:  "tenant-1",
		Timestamp: time.Now(),
		Tags:      map[string]string{"env": "prod"},
	}

	data, err := metric.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}