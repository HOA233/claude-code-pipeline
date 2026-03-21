package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Metrics Service Tests

func TestMetricsService_New(t *testing.T) {
	ms := service.NewMetricsService()
	if ms == nil {
		t.Fatal("Expected non-nil metrics service")
	}
}

func TestMetricsService_Record(t *testing.T) {
	ms := service.NewMetricsService()

	err := ms.Record("api.requests", 100, map[string]string{"endpoint": "/users"})
	if err != nil {
		t.Fatalf("Failed to record metric: %v", err)
	}
}

func TestMetricsService_Record_Multiple(t *testing.T) {
	ms := service.NewMetricsService()

	// Record same metric multiple times
	ms.Record("counter.test", 1, nil)
	ms.Record("counter.test", 1, nil)
	ms.Record("counter.test", 1, nil)

	value := ms.Get("counter.test")
	if value != 3 {
		t.Errorf("Expected value 3, got %f", value)
	}
}

func TestMetricsService_Get(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("get.test", 42, nil)

	value := ms.Get("get.test")
	if value != 42 {
		t.Errorf("Expected value 42, got %f", value)
	}
}

func TestMetricsService_Get_NotFound(t *testing.T) {
	ms := service.NewMetricsService()

	value := ms.Get("nonexistent")
	if value != 0 {
		t.Errorf("Expected 0 for nonexistent metric, got %f", value)
	}
}

func TestMetricsService_Increment(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Increment("increment.test")
	ms.Increment("increment.test")
	ms.Increment("increment.test")

	value := ms.Get("increment.test")
	if value != 3 {
		t.Errorf("Expected value 3, got %f", value)
	}
}

func TestMetricsService_Decrement(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("decrement.test", 10, nil)
	ms.Decrement("decrement.test")
	ms.Decrement("decrement.test")

	value := ms.Get("decrement.test")
	if value != 8 {
		t.Errorf("Expected value 8, got %f", value)
	}
}

func TestMetricsService_SetGauge(t *testing.T) {
	ms := service.NewMetricsService()

	ms.SetGauge("gauge.test", 75.5)

	value := ms.Get("gauge.test")
	if value != 75.5 {
		t.Errorf("Expected value 75.5, got %f", value)
	}

	// Set again should replace
	ms.SetGauge("gauge.test", 100.0)

	value = ms.Get("gauge.test")
	if value != 100.0 {
		t.Errorf("Expected value 100.0, got %f", value)
	}
}

func TestMetricsService_Observe(t *testing.T) {
	ms := service.NewMetricsService()

	// Observe values for histogram
	ms.Observe("latency.test", 100, nil)
	ms.Observe("latency.test", 200, nil)
	ms.Observe("latency.test", 300, nil)

	count := ms.GetCount("latency.test")
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestMetricsService_GetCount(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Increment("count.test")
	ms.Increment("count.test")

	count := ms.GetCount("count.test")
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestMetricsService_GetSum(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Observe("sum.test", 10, nil)
	ms.Observe("sum.test", 20, nil)
	ms.Observe("sum.test", 30, nil)

	sum := ms.GetSum("sum.test")
	if sum != 60 {
		t.Errorf("Expected sum 60, got %f", sum)
	}
}

func TestMetricsService_GetAverage(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Observe("avg.test", 10, nil)
	ms.Observe("avg.test", 20, nil)
	ms.Observe("avg.test", 30, nil)

	avg := ms.GetAverage("avg.test")
	if avg != 20 {
		t.Errorf("Expected average 20, got %f", avg)
	}
}

func TestMetricsService_GetMin(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Observe("min.test", 50, nil)
	ms.Observe("min.test", 10, nil)
	ms.Observe("min.test", 30, nil)

	min := ms.GetMin("min.test")
	if min != 10 {
		t.Errorf("Expected min 10, got %f", min)
	}
}

func TestMetricsService_GetMax(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Observe("max.test", 50, nil)
	ms.Observe("max.test", 10, nil)
	ms.Observe("max.test", 30, nil)

	max := ms.GetMax("max.test")
	if max != 50 {
		t.Errorf("Expected max 50, got %f", max)
	}
}

func TestMetricsService_GetPercentile(t *testing.T) {
	ms := service.NewMetricsService()

	// Add 100 values
	for i := 1; i <= 100; i++ {
		ms.Observe("percentile.test", float64(i), nil)
	}

	p50 := ms.GetPercentile("percentile.test", 50)
	p95 := ms.GetPercentile("percentile.test", 95)
	p99 := ms.GetPercentile("percentile.test", 99)

	// Approximately correct
	if p50 < 45 || p50 > 55 {
		t.Errorf("Expected p50 around 50, got %f", p50)
	}

	if p95 < 90 || p95 > 100 {
		t.Errorf("Expected p95 around 95, got %f", p95)
	}

	_ = p99
}

func TestMetricsService_ListMetrics(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("list.1", 1, nil)
	ms.Record("list.2", 2, nil)
	ms.Record("list.3", 3, nil)

	names := ms.ListMetrics()
	if len(names) < 3 {
		t.Errorf("Expected at least 3 metrics, got %d", len(names))
	}
}

func TestMetricsService_DeleteMetric(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("delete.test", 100, nil)

	err := ms.Delete("delete.test")
	if err != nil {
		t.Fatalf("Failed to delete metric: %v", err)
	}

	value := ms.Get("delete.test")
	if value != 0 {
		t.Error("Expected deleted metric to return 0")
	}
}

func TestMetricsService_Reset(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("reset.test", 100, nil)
	ms.Increment("reset.counter")

	ms.Reset("reset.test")

	value := ms.Get("reset.test")
	if value != 0 {
		t.Error("Expected reset metric to be 0")
	}
}

func TestMetricsService_ResetAll(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("resetall.1", 100, nil)
	ms.Record("resetall.2", 200, nil)

	ms.ResetAll()

	if ms.Get("resetall.1") != 0 || ms.Get("resetall.2") != 0 {
		t.Error("Expected all metrics to be reset")
	}
}

func TestMetricsService_Export(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("export.test", 100, nil)
	ms.Increment("export.counter")

	data := ms.Export()

	if len(data) == 0 {
		t.Error("Expected exported metrics")
	}
}

func TestMetricsService_TaggedMetrics(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("tagged.requests", 10, map[string]string{"env": "prod", "region": "us-east"})
	ms.Record("tagged.requests", 20, map[string]string{"env": "dev", "region": "us-west"})

	value := ms.GetWithTag("tagged.requests", map[string]string{"env": "prod"})
	if value != 10 {
		t.Errorf("Expected tagged value 10, got %f", value)
	}
}

func TestMetricsService_GetAllWithTags(t *testing.T) {
	ms := service.NewMetricsService()

	ms.Record("alltag.requests", 10, map[string]string{"env": "prod"})
	ms.Record("alltag.requests", 20, map[string]string{"env": "dev"})

	results := ms.GetAllWithTags("alltag.requests")
	if len(results) != 2 {
		t.Errorf("Expected 2 tagged results, got %d", len(results))
	}
}

func TestMetricsService_Timing(t *testing.T) {
	ms := service.NewMetricsService()

	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	ms.RecordTiming("timing.test", start, nil)

	avg := ms.GetAverage("timing.test")
	if avg < 40 { // Allow some margin
		t.Errorf("Expected timing around 50ms, got %f", avg)
	}
}