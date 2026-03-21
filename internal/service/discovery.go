package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
)

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"` // "healthy", "unhealthy", "degraded"
	Timestamp time.Time              `json:"timestamp"`
	Latency   time.Duration          `json:"latency"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker checks the health of a service
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) (*HealthStatus, error)
}

// HTTPHealthChecker checks HTTP endpoints
type HTTPHealthChecker struct {
	name     string
	url      string
	timeout  time.Duration
	expected int
}

// NewHTTPHealthChecker creates a new HTTP health checker
func NewHTTPHealthChecker(name, url string, timeout time.Duration) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		name:     name,
		url:      url,
		timeout:  timeout,
		expected: http.StatusOK,
	}
}

// Name returns the checker name
func (c *HTTPHealthChecker) Name() string {
	return c.name
}

// Check performs the health check
func (c *HTTPHealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()

	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return &HealthStatus{
			Name:      c.name,
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Message:   err.Error(),
		}, err
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return &HealthStatus{
			Name:      c.name,
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Latency:   latency,
			Message:   err.Error(),
		}, err
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode != c.expected {
		status = "unhealthy"
	}

	return &HealthStatus{
		Name:      c.name,
		Status:    status,
		Timestamp: time.Now(),
		Latency:   latency,
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
	}, nil
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	name  string
	redis *repository.RedisClient
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(name string, redis *repository.RedisClient) *RedisHealthChecker {
	return &RedisHealthChecker{name: name, redis: redis}
}

// Name returns the checker name
func (c *RedisHealthChecker) Name() string {
	return c.name
}

// Check performs the health check
func (c *RedisHealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	err := c.redis.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return &HealthStatus{
			Name:      c.name,
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Latency:   latency,
			Message:   err.Error(),
		}, err
	}

	return &HealthStatus{
		Name:      c.name,
		Status:    "healthy",
		Timestamp: time.Now(),
		Latency:   latency,
	}, nil
}

// HealthAggregator aggregates health checks
type HealthAggregator struct {
	mu       sync.RWMutex
	checkers map[string]HealthChecker
	results  map[string]*HealthStatus
	interval time.Duration
}

// NewHealthAggregator creates a new health aggregator
func NewHealthAggregator(interval time.Duration) *HealthAggregator {
	return &HealthAggregator{
		checkers: make(map[string]HealthChecker),
		results:  make(map[string]*HealthStatus),
		interval: interval,
	}
}

// Register registers a health checker
func (a *HealthAggregator) Register(checker HealthChecker) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkers[checker.Name()] = checker
}

// Run starts periodic health checks
func (a *HealthAggregator) Run(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	// Initial check
	a.checkAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.checkAll(ctx)
		}
	}
}

func (a *HealthAggregator) checkAll(ctx context.Context) {
	var wg sync.WaitGroup

	a.mu.RLock()
	checkers := make([]HealthChecker, 0, len(a.checkers))
	for _, c := range a.checkers {
		checkers = append(checkers, c)
	}
	a.mu.RUnlock()

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			result, _ := c.Check(ctx)

			a.mu.Lock()
			a.results[c.Name()] = result
			a.mu.Unlock()
		}(checker)
	}

	wg.Wait()
}

// GetHealth returns health status for all services
func (a *HealthAggregator) GetHealth() map[string]*HealthStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make(map[string]*HealthStatus)
	for k, v := range a.results {
		results[k] = v
	}
	return results
}

// GetOverallStatus returns the overall system status
func (a *HealthAggregator) GetOverallStatus() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	healthy := 0
	unhealthy := 0

	for _, status := range a.results {
		switch status.Status {
		case "healthy":
			healthy++
		case "unhealthy":
			unhealthy++
		}
	}

	if unhealthy == 0 {
		return "healthy"
	}
	if healthy == 0 {
		return "unhealthy"
	}
	return "degraded"
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]*ServiceInfo
}

// ServiceInfo contains information about a service
type ServiceInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Metadata  map[string]string `json:"metadata"`
	RegisteredAt time.Time      `json:"registered_at"`
	LastSeen  time.Time         `json:"last_seen"`
	Healthy   bool              `json:"healthy"`
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
	}
}

// Register registers a service
func (r *ServiceRegistry) Register(info *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info.RegisteredAt = time.Now()
	info.LastSeen = time.Now()
	info.Healthy = true

	r.services[info.ID] = info
	return nil
}

// Deregister deregisters a service
func (r *ServiceRegistry) Deregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, id)
}

// GetService gets a service by ID
func (r *ServiceRegistry) GetService(id string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[id]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", id)
	}
	return svc, nil
}

// GetServices gets all services by name
func (r *ServiceRegistry) GetServices(name string) []*ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*ServiceInfo
	for _, svc := range r.services {
		if svc.Name == name {
			services = append(services, svc)
		}
	}
	return services
}

// GetAllServices gets all registered services
func (r *ServiceRegistry) GetAllServices() []*ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(r.services))
	for _, svc := range r.services {
		services = append(services, svc)
	}
	return services
}

// Heartbeat updates the last seen time for a service
func (r *ServiceRegistry) Heartbeat(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.services[id]
	if !exists {
		return fmt.Errorf("service not found: %s", id)
	}

	svc.LastSeen = time.Now()
	svc.Healthy = true
	return nil
}

// CleanupStale removes services that haven't sent a heartbeat
func (r *ServiceRegistry) CleanupStale(maxAge time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, svc := range r.services {
		if svc.LastSeen.Before(cutoff) {
			delete(r.services, id)
			removed++
		}
	}

	return removed
}

// ToJSON exports the registry as JSON
func (r *ServiceRegistry) ToJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(r.services))
	for _, svc := range r.services {
		services = append(services, svc)
	}

	return json.MarshalIndent(services, "", "  ")
}