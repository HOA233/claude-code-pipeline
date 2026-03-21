package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Plugin Service Tests

func TestPluginService_New(t *testing.T) {
	ps := service.NewPluginService()
	if ps == nil {
		t.Fatal("Expected non-nil plugin service")
	}
}

func TestPluginService_Register(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:          "plugin-1",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "test-author",
		TenantID:    "tenant-1",
		Enabled:     true,
	}

	err := ps.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}
}

func TestPluginService_Register_MissingName(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:      "plugin-no-name",
		Version: "1.0.0",
	}

	err := ps.Register(plugin)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestPluginService_Get(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-get",
		Name:     "Get Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-get",
	}
	ps.Register(plugin)

	retrieved, err := ps.Get("plugin-get")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if retrieved.Name != "Get Plugin" {
		t.Error("Plugin name mismatch")
	}
}

func TestPluginService_Get_NotFound(t *testing.T) {
	ps := service.NewPluginService()

	_, err := ps.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent plugin")
	}
}

func TestPluginService_List(t *testing.T) {
	ps := service.NewPluginService()

	ps.Register(&service.Plugin{
		ID:       "list-1",
		Name:     "List Plugin 1",
		Version:  "1.0.0",
		TenantID: "tenant-list",
	})

	ps.Register(&service.Plugin{
		ID:       "list-2",
		Name:     "List Plugin 2",
		Version:  "1.0.0",
		TenantID: "tenant-list",
	})

	plugins := ps.List("tenant-list")
	if len(plugins) < 2 {
		t.Errorf("Expected at least 2 plugins, got %d", len(plugins))
	}
}

func TestPluginService_Enable(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-enable",
		Name:     "Enable Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-enable",
		Enabled:  false,
	}
	ps.Register(plugin)

	err := ps.Enable("plugin-enable")
	if err != nil {
		t.Fatalf("Failed to enable plugin: %v", err)
	}

	retrieved, _ := ps.Get("plugin-enable")
	if !retrieved.Enabled {
		t.Error("Expected plugin to be enabled")
	}
}

func TestPluginService_Disable(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-disable",
		Name:     "Disable Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-disable",
		Enabled:  true,
	}
	ps.Register(plugin)

	err := ps.Disable("plugin-disable")
	if err != nil {
		t.Fatalf("Failed to disable plugin: %v", err)
	}

	retrieved, _ := ps.Get("plugin-disable")
	if retrieved.Enabled {
		t.Error("Expected plugin to be disabled")
	}
}

func TestPluginService_Update(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:          "plugin-update",
		Name:        "Update Plugin",
		Version:     "1.0.0",
		Description: "Original description",
		TenantID:    "tenant-update",
	}
	ps.Register(plugin)

	err := ps.Update("plugin-update", map[string]interface{}{
		"description": "Updated description",
		"version":     "2.0.0",
	})

	if err != nil {
		t.Fatalf("Failed to update plugin: %v", err)
	}

	retrieved, _ := ps.Get("plugin-update")
	if retrieved.Description != "Updated description" {
		t.Error("Description not updated")
	}
}

func TestPluginService_Delete(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-delete",
		Name:     "Delete Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-delete",
	}
	ps.Register(plugin)

	err := ps.Delete("plugin-delete")
	if err != nil {
		t.Fatalf("Failed to delete plugin: %v", err)
	}

	_, err = ps.Get("plugin-delete")
	if err == nil {
		t.Error("Expected error for deleted plugin")
	}
}

func TestPluginService_Execute(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-exec",
		Name:     "Execute Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-exec",
		Enabled:  true,
		Handler:  "echo",
	}
	ps.Register(plugin)

	result, err := ps.Execute("plugin-exec", map[string]interface{}{
		"input": "test",
	})

	if err != nil {
		t.Fatalf("Failed to execute plugin: %v", err)
	}

	_ = result
}

func TestPluginService_GetConfig(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-config",
		Name:     "Config Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-config",
		Config: map[string]interface{}{
			"setting1": "value1",
			"setting2": 123,
		},
	}
	ps.Register(plugin)

	config := ps.GetConfig("plugin-config")
	if config == nil {
		t.Fatal("Expected config")
	}

	if config["setting1"] != "value1" {
		t.Error("Config value mismatch")
	}
}

func TestPluginService_SetConfig(t *testing.T) {
	ps := service.NewPluginService()

	plugin := &service.Plugin{
		ID:       "plugin-setconfig",
		Name:     "Set Config Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-setconfig",
	}
	ps.Register(plugin)

	err := ps.SetConfig("plugin-setconfig", map[string]interface{}{
		"new_setting": "new_value",
	})

	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}
}

func TestPluginService_GetStats(t *testing.T) {
	ps := service.NewPluginService()

	ps.Register(&service.Plugin{
		ID:       "stats-1",
		Name:     "Stats Plugin",
		Version:  "1.0.0",
		TenantID: "tenant-stats",
		Enabled:  true,
	})

	ps.Register(&service.Plugin{
		ID:       "stats-2",
		Name:     "Stats Plugin 2",
		Version:  "1.0.0",
		TenantID: "tenant-stats",
		Enabled:  false,
	})

	stats := ps.GetStats("tenant-stats")

	if stats.Total < 2 {
		t.Errorf("Expected at least 2 plugins, got %d", stats.Total)
	}
}

func TestPluginService_PluginToJSON(t *testing.T) {
	plugin := &service.Plugin{
		ID:          "json-1",
		Name:        "JSON Plugin",
		Version:     "1.0.0",
		Description: "Test",
		Author:      "author",
		TenantID:    "tenant-1",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	data, err := plugin.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}