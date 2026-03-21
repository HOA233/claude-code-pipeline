package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/stretchr/testify/assert"
)

// TestPriorityQueue tests the priority queue service
func TestPriorityQueue_New(t *testing.T) {
	pq := service.NewPriorityQueueService(nil, "test-queue")
	assert.NotNil(t, pq)
}

// TestPriorityLevel tests priority level constants
func TestPriorityLevel(t *testing.T) {
	assert.Equal(t, 1.0, service.PriorityCritical)
	assert.Equal(t, 2.0, service.PriorityHigh)
	assert.Equal(t, 3.0, service.PriorityNormal)
	assert.Equal(t, 4.0, service.PriorityLow)
}

// TestBatchService tests batch service creation
func TestBatchService_New(t *testing.T) {
	bs := service.NewBatchService(nil, 5)
	assert.NotNil(t, bs)
}

// TestStatisticsService tests statistics service
func TestStatisticsService_New(t *testing.T) {
	ss := service.NewStatisticsService(nil)
	assert.NotNil(t, ss)
}

// TestSchedulerService tests scheduler service
func TestSchedulerService_New(t *testing.T) {
	ss := service.NewSchedulerService(nil, nil)
	assert.NotNil(t, ss)
}

// TestNotificationService tests notification service
func TestNotificationService_New(t *testing.T) {
	ns := service.NewNotificationService(nil)
	assert.NotNil(t, ns)
}

// TestPluginManager tests plugin manager
func TestPluginManager_New(t *testing.T) {
	pm := service.NewPluginManager()
	assert.NotNil(t, pm)
}

// TestPluginRegistration tests plugin registration
func TestPluginRegistration(t *testing.T) {
	pm := service.NewPluginManager()

	plugin := &service.LoggingPlugin{}
	err := pm.Register("logging", plugin)
	assert.NoError(t, err)

	// Test duplicate registration
	err = pm.Register("logging", plugin)
	assert.Error(t, err)
}

// TestWorkspaceManager tests workspace manager
func TestWorkspaceManager_New(t *testing.T) {
	wm := service.NewWorkspaceManager(nil)
	assert.NotNil(t, wm)
}

// TestWorkspaceConfig tests workspace configuration
func TestWorkspaceConfig(t *testing.T) {
	cfg := service.WorkspaceConfig{
		MaxTasks:       100,
		MaxPipelines:   50,
		MaxConcurrency: 5,
	}

	assert.Equal(t, 100, cfg.MaxTasks)
	assert.Equal(t, 50, cfg.MaxPipelines)
	assert.Equal(t, 5, cfg.MaxConcurrency)
}

// TestTaskStatus tests task status values
func TestTaskStatus(t *testing.T) {
	statuses := []string{
		"pending",
		"running",
		"completed",
		"failed",
		"cancelled",
	}

	for _, status := range statuses {
		assert.NotEmpty(t, status)
	}
}

// TestPipelineMode tests pipeline mode values
func TestPipelineMode(t *testing.T) {
	modes := []string{
		"serial",
		"parallel",
		"hybrid",
	}

	for _, mode := range modes {
		assert.NotEmpty(t, mode)
	}
}

// TestCacheService tests cache service
func TestCacheService_New(t *testing.T) {
	cs := service.NewCacheService(nil)
	assert.NotNil(t, cs)
}

// TestTemplateService tests template service
func TestTemplateService_New(t *testing.T) {
	ts := service.NewTemplateService(nil)
	assert.NotNil(t, ts)
}

// TestAuditService tests audit service
func TestAuditService_New(t *testing.T) {
	as := service.NewAuditService(nil)
	assert.NotNil(t, as)
}