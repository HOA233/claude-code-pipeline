package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
)

// Plugin represents a service plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Initialize is called when the plugin is loaded
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Shutdown is called when the plugin is unloaded
	Shutdown(ctx context.Context) error

	// Hooks that the plugin can implement
	OnTaskCreated(ctx context.Context, taskID string, skillID string) error
	OnTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) error
	OnTaskFailed(ctx context.Context, taskID string, err error) error
	OnPipelineRun(ctx context.Context, pipelineID string, runID string) error
	OnPipelineComplete(ctx context.Context, pipelineID string, runID string, success bool) error
}

// BasePlugin provides default implementations for plugin hooks
type BasePlugin struct {
	name    string
	version string
}

func (p *BasePlugin) Name() string    { return p.name }
func (p *BasePlugin) Version() string { return p.version }
func (p *BasePlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}
func (p *BasePlugin) Shutdown(ctx context.Context) error {
	return nil
}
func (p *BasePlugin) OnTaskCreated(ctx context.Context, taskID string, skillID string) error {
	return nil
}
func (p *BasePlugin) OnTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) error {
	return nil
}
func (p *BasePlugin) OnTaskFailed(ctx context.Context, taskID string, err error) error {
	return nil
}
func (p *BasePlugin) OnPipelineRun(ctx context.Context, pipelineID string, runID string) error {
	return nil
}
func (p *BasePlugin) OnPipelineComplete(ctx context.Context, pipelineID string, runID string, success bool) error {
	return nil
}

// PluginManager manages loaded plugins
type PluginManager struct {
	redis   *repository.RedisClient
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(redis *repository.RedisClient) *PluginManager {
	return &PluginManager{
		redis:   redis,
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	pm.plugins[name] = plugin
	logger.Info(fmt.Sprintf("Plugin registered: %s v%s", name, plugin.Version()))
	return nil
}

// Unregister removes a plugin
func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Call shutdown
	if err := plugin.Shutdown(context.Background()); err != nil {
		logger.Error(fmt.Sprintf("Plugin %s shutdown error: %v", name, err))
	}

	delete(pm.plugins, name)
	logger.Info(fmt.Sprintf("Plugin unregistered: %s", name))
	return nil
}

// Get retrieves a plugin by name
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// List returns all registered plugins
func (pm *PluginManager) List() []PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]PluginInfo, 0, len(pm.plugins))
	for name, plugin := range pm.plugins {
		plugins = append(plugins, PluginInfo{
			Name:    name,
			Version: plugin.Version(),
			Status:  "active",
		})
	}
	return plugins
}

// PluginInfo contains plugin information
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// InitializeAll initializes all plugins
func (pm *PluginManager) InitializeAll(ctx context.Context) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		config := make(map[string]interface{})
		if err := plugin.Initialize(ctx, config); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
		}
		logger.Info(fmt.Sprintf("Plugin initialized: %s", name))
	}
	return nil
}

// ShutdownAll shuts down all plugins
func (pm *PluginManager) ShutdownAll(ctx context.Context) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var errors []error
	for name, plugin := range pm.plugins {
		if err := plugin.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("plugin %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	return nil
}

// Hook methods - call all plugins

// OnTaskCreated notifies all plugins of task creation
func (pm *PluginManager) OnTaskCreated(ctx context.Context, taskID string, skillID string) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		if err := plugin.OnTaskCreated(ctx, taskID, skillID); err != nil {
			logger.Error(fmt.Sprintf("Plugin %s OnTaskCreated error: %v", name, err))
		}
	}
}

// OnTaskCompleted notifies all plugins of task completion
func (pm *PluginManager) OnTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		if err := plugin.OnTaskCompleted(ctx, taskID, result); err != nil {
			logger.Error(fmt.Sprintf("Plugin %s OnTaskCompleted error: %v", name, err))
		}
	}
}

// OnTaskFailed notifies all plugins of task failure
func (pm *PluginManager) OnTaskFailed(ctx context.Context, taskID string, err error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		if err := plugin.OnTaskFailed(ctx, taskID, err); err != nil {
			logger.Error(fmt.Sprintf("Plugin %s OnTaskFailed error: %v", name, err))
		}
	}
}

// OnPipelineRun notifies all plugins of pipeline run
func (pm *PluginManager) OnPipelineRun(ctx context.Context, pipelineID string, runID string) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		if err := plugin.OnPipelineRun(ctx, pipelineID, runID); err != nil {
			logger.Error(fmt.Sprintf("Plugin %s OnPipelineRun error: %v", name, err))
		}
	}
}

// OnPipelineComplete notifies all plugins of pipeline completion
func (pm *PluginManager) OnPipelineComplete(ctx context.Context, pipelineID string, runID string, success bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, plugin := range pm.plugins {
		if err := plugin.OnPipelineComplete(ctx, pipelineID, runID, success); err != nil {
			logger.Error(fmt.Sprintf("Plugin %s OnPipelineComplete error: %v", name, err))
		}
	}
}

// Example plugin implementations

// LoggingPlugin logs all events
type LoggingPlugin struct {
	BasePlugin
}

func NewLoggingPlugin() *LoggingPlugin {
	return &LoggingPlugin{
		BasePlugin: BasePlugin{name: "logging", version: "1.0.0"},
	}
}

func (p *LoggingPlugin) OnTaskCreated(ctx context.Context, taskID string, skillID string) error {
	logger.Info(fmt.Sprintf("[LoggingPlugin] Task created: %s (skill: %s)", taskID, skillID))
	return nil
}

func (p *LoggingPlugin) OnTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) error {
	logger.Info(fmt.Sprintf("[LoggingPlugin] Task completed: %s", taskID))
	return nil
}

func (p *LoggingPlugin) OnTaskFailed(ctx context.Context, taskID string, err error) error {
	logger.Info(fmt.Sprintf("[LoggingPlugin] Task failed: %s (error: %v)", taskID, err))
	return nil
}

// MetricsPlugin collects metrics
type MetricsPlugin struct {
	BasePlugin
	taskCount     int64
	failureCount  int64
	successCount  int64
}

func NewMetricsPlugin() *MetricsPlugin {
	return &MetricsPlugin{
		BasePlugin: BasePlugin{name: "metrics", version: "1.0.0"},
	}
}

func (p *MetricsPlugin) OnTaskCreated(ctx context.Context, taskID string, skillID string) error {
	p.taskCount++
	return nil
}

func (p *MetricsPlugin) OnTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) error {
	p.successCount++
	return nil
}

func (p *MetricsPlugin) OnTaskFailed(ctx context.Context, taskID string, err error) error {
	p.failureCount++
	return nil
}

func (p *MetricsPlugin) GetMetrics() map[string]int64 {
	return map[string]int64{
		"tasks_total":    p.taskCount,
		"tasks_success":  p.successCount,
		"tasks_failed":   p.failureCount,
	}
}