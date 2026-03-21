package service

import (
	"context"
	"sync"
	"time"
)

// MetricsCollector collects and aggregates metrics
type MetricsCollector struct {
	mu         sync.RWMutex
	counters   map[string]int64
	gauges     map[string]float64
	histograms map[string][]float64
	startTime  time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
		startTime:  time.Now(),
	}
}

// IncrementCounter increments a counter by 1
func (m *MetricsCollector) IncrementCounter(name string) {
	m.IncrementCounterBy(name, 1)
}

// IncrementCounterBy increments a counter by a specific value
func (m *MetricsCollector) IncrementCounterBy(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

// DecrementCounter decrements a counter by 1
func (m *MetricsCollector) DecrementCounter(name string) {
	m.IncrementCounterBy(name, -1)
}

// SetGauge sets a gauge value
func (m *MetricsCollector) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// RecordHistogram records a value in a histogram
func (m *MetricsCollector) RecordHistogram(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histograms[name] = append(m.histograms[name], value)
}

// GetCounter returns a counter value
func (m *MetricsCollector) GetCounter(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

// GetGauge returns a gauge value
func (m *MetricsCollector) GetGauge(name string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gauges[name]
}

// GetHistogramStats returns statistics for a histogram
func (m *MetricsCollector) GetHistogramStats(name string) HistogramStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := m.histograms[name]
	if len(values) == 0 {
		return HistogramStats{}
	}

	stats := HistogramStats{Count: int64(len(values))}
	var sum float64
	min := values[0]
	max := values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	stats.Min = min
	stats.Max = max
	stats.Sum = sum
	stats.Mean = sum / float64(len(values))

	return stats
}

// HistogramStats represents histogram statistics
type HistogramStats struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
}

// AllMetrics returns all collected metrics
func (m *MetricsCollector) AllMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy counters
	counters := make(map[string]int64)
	for k, v := range m.counters {
		counters[k] = v
	}

	// Copy gauges
	gauges := make(map[string]float64)
	for k, v := range m.gauges {
		gauges[k] = v
	}

	// Get histogram stats
	histogramStats := make(map[string]HistogramStats)
	for k := range m.histograms {
		histogramStats[k] = m.GetHistogramStats(k)
	}

	return map[string]interface{}{
		"uptime_seconds": time.Since(m.startTime).Seconds(),
		"counters":       counters,
		"gauges":         gauges,
		"histograms":     histogramStats,
	}
}

// Reset resets all metrics
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters = make(map[string]int64)
	m.gauges = make(map[string]float64)
	m.histograms = make(map[string][]float64)
	m.startTime = time.Now()
}

// Timing helps measure execution time
type Timing struct {
	start time.Time
	name  string
	m     *MetricsCollector
}

// StartTiming starts a timing measurement
func (m *MetricsCollector) StartTiming(name string) *Timing {
	return &Timing{
		start: time.Now(),
		name:  name,
		m:     m,
	}
}

// Stop stops the timing and records the duration in milliseconds
func (t *Timing) Stop() {
	duration := time.Since(t.start).Seconds() * 1000
	t.m.RecordHistogram(t.name, duration)
}

// MetricsMiddleware tracks request metrics
type MetricsMiddleware struct {
	collector *MetricsCollector
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(collector *MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{collector: collector}
}

// RecordRequest records an API request
func (m *MetricsMiddleware) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	m.collector.IncrementCounter("requests_total")
	m.collector.IncrementCounter("requests_" + method)
	m.collector.RecordHistogram("request_duration_ms", float64(duration.Milliseconds()))

	if statusCode >= 400 {
		m.collector.IncrementCounter("requests_errors")
	}
}

// RecordTaskCreated records task creation
func (m *MetricsMiddleware) RecordTaskCreated(skillID string) {
	m.collector.IncrementCounter("tasks_created")
	m.collector.IncrementCounter("tasks_skill_" + skillID)
}

// RecordTaskCompleted records task completion
func (m *MetricsMiddleware) RecordTaskCompleted(skillID string, duration time.Duration) {
	m.collector.IncrementCounter("tasks_completed")
	m.collector.RecordHistogram("task_duration_seconds", duration.Seconds())
}

// RecordTaskFailed records task failure
func (m *MetricsMiddleware) RecordTaskFailed(skillID string) {
	m.collector.IncrementCounter("tasks_failed")
}

// UpdateQueueLength updates the queue length gauge
func (m *MetricsMiddleware) UpdateQueueLength(length int) {
	m.collector.SetGauge("queue_length", float64(length))
}

// UpdateActiveTasks updates the active tasks gauge
func (m *MetricsMiddleware) UpdateActiveTasks(count int) {
	m.collector.SetGauge("active_tasks", float64(count))
}

// MetricsRecorder records metrics to Redis
func (m *MetricsCollector) RecordToRedis(ctx context.Context, redis *repository.RedisClient) error {
	metrics := m.AllMetrics()

	for name, value := range metrics["counters"].(map[string]int64) {
		redis.IncrementCounter(ctx, name)
	}

	return nil
}