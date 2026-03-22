package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// HealthService provides health check and status reporting
type HealthService struct {
	mu          sync.RWMutex
	checks      map[string]HealthCheck
	status      ServiceStatus
	version     string
	startTime   time.Time
	dependencies map[string]DependencyHealth
}

// HealthCheck represents a health check function
type HealthCheck struct {
	Name        string
	Check       func() error
	Interval    time.Duration
	LastRun     time.Time
	LastStatus  string
	LastError   string
	LatencyMs   int64
	Consecutive int
}

// ServiceStatus represents the overall service status
type ServiceStatus struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Uptime      int64             `json:"uptime_seconds"`
	StartTime   time.Time         `json:"start_time"`
	Checks      map[string]CheckResult `json:"checks"`
	System      SystemInfo        `json:"system"`
	Dependencies map[string]DependencyHealth `json:"dependencies"`
}

// CheckResult represents a health check result
type CheckResult struct {
	Status    string `json:"status"`
	LastRun   string `json:"last_run"`
	Error     string `json:"error,omitempty"`
	LatencyMs int64  `json:"latency_ms"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
}

// DependencyHealth represents dependency health status
type DependencyHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency int64  `json:"latency_ms"`
	Error   string `json:"error,omitempty"`
}

// HealthStatus constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDegraded  = "degraded"
)

// NewHealthService creates a new health service
func NewHealthService(version string) *HealthService {
	return &HealthService{
		checks:      make(map[string]HealthCheck),
		version:     version,
		startTime:   time.Now(),
		dependencies: make(map[string]DependencyHealth),
	}
}

// RegisterCheck registers a health check
func (s *HealthService) RegisterCheck(name string, check func() error, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.checks[name] = HealthCheck{
		Name:     name,
		Check:    check,
		Interval: interval,
	}
}

// RunChecks runs all health checks
func (s *HealthService) RunChecks(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, check := range s.checks {
		start := time.Now()
		err := check.Check()
		latency := time.Since(start).Milliseconds()

		check.LastRun = time.Now()
		check.LatencyMs = latency

		if err != nil {
			check.LastStatus = HealthStatusUnhealthy
			check.LastError = err.Error()
			check.Consecutive++
		} else {
			check.LastStatus = HealthStatusHealthy
			check.LastError = ""
			check.Consecutive = 0
		}

		s.checks[name] = check
	}
}

// GetStatus returns the current service status
func (s *HealthService) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Build check results
	checks := make(map[string]CheckResult)
	allHealthy := true
	for name, check := range s.checks {
		status := check.LastStatus
		if status == "" {
			status = HealthStatusHealthy
		}
		if status != HealthStatusHealthy {
			allHealthy = false
		}

		checks[name] = CheckResult{
			Status:    status,
			LastRun:   check.LastRun.Format(time.RFC3339),
			Error:     check.LastError,
			LatencyMs: check.LatencyMs,
		}
	}

	// Check dependencies
	for name, dep := range s.dependencies {
		if dep.Status != HealthStatusHealthy {
			allHealthy = false
		}
		checks[name+"_dep"] = CheckResult{
			Status: dep.Status,
			Error:  dep.Error,
			LatencyMs: dep.Latency,
		}
	}

	// Determine overall status
	overall := HealthStatusHealthy
	if !allHealthy {
		overall = HealthStatusDegraded
	}

	return &ServiceStatus{
		Status:      overall,
		Version:     s.version,
		Uptime:      int64(time.Since(s.startTime).Seconds()),
		StartTime:   s.startTime,
		Checks:      checks,
		Dependencies: s.dependencies,
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
			MemAllocMB:   m.Alloc / 1024 / 1024,
			MemTotalMB:   m.TotalAlloc / 1024 / 1024,
			MemSysMB:     m.Sys / 1024 / 1024,
		},
	}
}

// IsHealthy returns whether the service is healthy
func (s *HealthService) IsHealthy() bool {
	status := s.GetStatus()
	return status.Status == HealthStatusHealthy
}

// AddDependency adds a dependency health status
func (s *HealthService) AddDependency(name string, health DependencyHealth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dependencies[name] = health
}

// SetDependencyHealth updates a dependency's health
func (s *HealthService) SetDependencyHealth(name, status string, latency int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep := DependencyHealth{
		Name:    name,
		Status:  status,
		Latency: latency,
	}
	if err != nil {
		dep.Error = err.Error()
	}
	s.dependencies[name] = dep
}

// CheckDependency checks a dependency's health
func (s *HealthService) CheckDependency(ctx context.Context, name string, check func() error) error {
	start := time.Now()
	err := check()
	latency := time.Since(start).Milliseconds()

	status := HealthStatusHealthy
	if err != nil {
		status = HealthStatusUnhealthy
	}

	s.SetDependencyHealth(name, status, latency, err)
	return err
}

// GetLive returns live status (simple up/down)
func (s *HealthService) GetLive() map[string]interface{} {
	return map[string]interface{}{
		"status": "up",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}
}

// GetReady returns readiness status
func (s *HealthService) GetReady() (map[string]interface{}, error) {
	status := s.GetStatus()

	ready := true
	reasons := []string{}

	for name, check := range status.Checks {
		if check.Status == HealthStatusUnhealthy {
			ready = false
			reasons = append(reasons, fmt.Sprintf("%s is unhealthy", name))
		}
	}

	result := map[string]interface{}{
		"ready":   ready,
		"status":  status.Status,
		"version": status.Version,
		"uptime":  status.Uptime,
	}

	if !ready {
		result["reasons"] = reasons
		return result, errors.New("service not ready")
	}

	return result, nil
}

// StartBackgroundChecks starts periodic health checks
func (s *HealthService) StartBackgroundChecks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.RunChecks(ctx)
		}
	}
}

// ToJSON serializes status to JSON
func (s *ServiceStatus) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ReadinessProbe creates a readiness probe
type ReadinessProbe struct {
	name     string
	check    func() bool
	required bool
}

// NewReadinessProbe creates a new readiness probe
func NewReadinessProbe(name string, check func() bool, required bool) *ReadinessProbe {
	return &ReadinessProbe{
		name:     name,
		check:    check,
		required: required,
	}
}

// LivenessProbe creates a liveness probe
type LivenessProbe struct {
	name      string
	check     func() bool
	threshold int
	failures  int
}

// NewLivenessProbe creates a new liveness probe
func NewLivenessProbe(name string, check func() bool, threshold int) *LivenessProbe {
	return &LivenessProbe{
		name:      name,
		check:     check,
		threshold: threshold,
	}
}

// Check executes the liveness check
func (p *LivenessProbe) Check() bool {
	if p.check() {
		p.failures = 0
		return true
	}
	p.failures++
	return false
}

// IsAlive returns true if failures are below threshold
func (p *LivenessProbe) IsAlive() bool {
	return p.failures < p.threshold
}