package tests

import (
	"context"
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestConfigService_New(t *testing.T) {
	config := service.NewConfigService()
	if config == nil {
		t.Fatal("Expected non-nil config service")
	}
}

func TestConfigService_Set(t *testing.T) {
	cs := service.NewConfigService()

	config, err := cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:         "app.name",
		Value:       "Claude Pipeline",
		Type:        "string",
		Description: "Application name",
		User:        "admin",
	})

	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	if config.Key != "app.name" {
		t.Errorf("Expected key 'app.name', got '%s'", config.Key)
	}
	if config.Value != "Claude Pipeline" {
		t.Errorf("Expected value 'Claude Pipeline', got '%v'", config.Value)
	}
	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}
}

func TestConfigService_Set_MissingKey(t *testing.T) {
	cs := service.NewConfigService()

	_, err := cs.Set(context.Background(), &service.ConfigSetRequest{
		Value: "test",
	})

	if err == nil {
		t.Error("Expected error for missing key")
	}
}

func TestConfigService_Get(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "test.key",
		Value: "test-value",
	})

	config, err := cs.Get(context.Background(), "test.key", "", "")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if config.Value != "test-value" {
		t.Errorf("Expected 'test-value', got '%v'", config.Value)
	}
}

func TestConfigService_Get_NotFound(t *testing.T) {
	cs := service.NewConfigService()

	_, err := cs.Get(context.Background(), "nonexistent", "", "")
	if err == nil {
		t.Error("Expected error for nonexistent config")
	}
}

func TestConfigService_GetString(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "string.key",
		Value: "string-value",
	})

	value, err := cs.GetString(context.Background(), "string.key", "", "")
	if err != nil {
		t.Fatalf("Failed to get string: %v", err)
	}

	if value != "string-value" {
		t.Errorf("Expected 'string-value', got '%s'", value)
	}
}

func TestConfigService_GetInt(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "int.key",
		Value: 42,
	})

	value, err := cs.GetInt(context.Background(), "int.key", "", "")
	if err != nil {
		t.Fatalf("Failed to get int: %v", err)
	}

	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

func TestConfigService_GetBool(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "bool.key",
		Value: true,
	})

	value, err := cs.GetBool(context.Background(), "bool.key", "", "")
	if err != nil {
		t.Fatalf("Failed to get bool: %v", err)
	}

	if !value {
		t.Error("Expected true")
	}
}

func TestConfigService_Profiles(t *testing.T) {
	cs := service.NewConfigService()

	// Set config in different profiles
	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:     "db.host",
		Value:   "localhost",
		Profile: "development",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:     "db.host",
		Value:   "prod-db.example.com",
		Profile: "production",
	})

	devConfig, _ := cs.Get(context.Background(), "db.host", "development", "")
	prodConfig, _ := cs.Get(context.Background(), "db.host", "production", "")

	if devConfig.Value != "localhost" {
		t.Errorf("Expected 'localhost' for dev, got '%v'", devConfig.Value)
	}
	if prodConfig.Value != "prod-db.example.com" {
		t.Errorf("Expected 'prod-db.example.com' for prod, got '%v'", prodConfig.Value)
	}
}

func TestConfigService_ListProfiles(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:     "test",
		Value:   "dev",
		Profile: "development",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:     "test",
		Value:   "prod",
		Profile: "production",
	})

	profiles := cs.ListProfiles()
	if len(profiles) < 2 {
		t.Errorf("Expected at least 2 profiles, got %d", len(profiles))
	}
}

func TestConfigService_Delete(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "delete.test",
		Value: "test",
	})

	err := cs.Delete(context.Background(), "delete.test", "", "")
	if err != nil {
		t.Fatalf("Failed to delete config: %v", err)
	}

	_, err = cs.Get(context.Background(), "delete.test", "", "")
	if err == nil {
		t.Error("Expected error for deleted config")
	}
}

func TestConfigService_Update(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "update.test",
		Value: "initial",
		User:  "user1",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "update.test",
		Value: "updated",
		User:  "user2",
	})

	config, _ := cs.Get(context.Background(), "update.test", "", "")

	if config.Value != "updated" {
		t.Errorf("Expected 'updated', got '%v'", config.Value)
	}
	if config.Version != 2 {
		t.Errorf("Expected version 2, got %d", config.Version)
	}
	if config.UpdatedBy != "user2" {
		t.Errorf("Expected updated_by 'user2', got '%s'", config.UpdatedBy)
	}
}

func TestConfigService_Watch(t *testing.T) {
	cs := service.NewConfigService()

	changes := 0
	cs.Watch("watch.test", func(key string, oldValue, newValue interface{}) {
		changes++
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "watch.test",
		Value: "first",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "watch.test",
		Value: "second",
	})

	// Wait for goroutines
	for i := 0; i < 10; i++ {
		if changes >= 2 {
			break
		}
	}

	if changes < 2 {
		t.Errorf("Expected at least 2 change notifications, got %d", changes)
	}
}

func TestConfigService_GetHistory(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:    "history.test",
		Value:  "v1",
		Reason: "Initial",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:    "history.test",
		Value:  "v2",
		Reason: "Update",
	})

	history := cs.GetHistory("history.test", "", "")
	if len(history) < 2 {
		t.Errorf("Expected at least 2 history entries, got %d", len(history))
	}
}

func TestConfigService_Rollback(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "rollback.test",
		Value: "original",
		User:  "user1",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "rollback.test",
		Value: "changed",
		User:  "user2",
	})

	// Rollback to version 1
	config, err := cs.Rollback(context.Background(), "rollback.test", "", "", 1, "admin")
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	if config.Value != "original" {
		t.Errorf("Expected 'original' after rollback, got '%v'", config.Value)
	}
	if config.Version != 3 {
		t.Errorf("Expected version 3, got %d", config.Version)
	}
}

func TestConfigService_Export(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "export.test1",
		Value: "value1",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "export.test2",
		Value: "value2",
	})

	data, err := cs.Export("default", "")
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty export data")
	}
}

func TestConfigService_Import(t *testing.T) {
	cs := service.NewConfigService()

	data := `[{"key":"import.test1","value":"val1","type":"string"},{"key":"import.test2","value":"val2","type":"string"}]`

	err := cs.Import(context.Background(), []byte(data), "default", "admin", false)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	config1, err := cs.Get(context.Background(), "import.test1", "", "")
	if err != nil {
		t.Fatalf("Failed to get imported config: %v", err)
	}

	if config1.Value != "val1" {
		t.Errorf("Expected 'val1', got '%v'", config1.Value)
	}
}

func TestConfigService_GetStats(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:   "stats.test",
		Value: "test",
	})

	stats := cs.GetStats()

	if stats["total_configs"].(int) < 1 {
		t.Error("Expected at least 1 config")
	}
}

func TestConfigService_ProjectIsolation(t *testing.T) {
	cs := service.NewConfigService()

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:       "project.key",
		Value:     "project-a-value",
		ProjectID: "project-a",
	})

	cs.Set(context.Background(), &service.ConfigSetRequest{
		Key:       "project.key",
		Value:     "project-b-value",
		ProjectID: "project-b",
	})

	configA, _ := cs.Get(context.Background(), "project.key", "", "project-a")
	configB, _ := cs.Get(context.Background(), "project.key", "", "project-b")

	if configA.Value != "project-a-value" {
		t.Error("Project A config mismatch")
	}
	if configB.Value != "project-b-value" {
		t.Error("Project B config mismatch")
	}
}