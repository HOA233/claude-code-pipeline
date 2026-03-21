package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Discovery Service Tests

func TestDiscoveryService_New(t *testing.T) {
	ds := service.NewDiscoveryService()
	if ds == nil {
		t.Fatal("Expected non-nil discovery service")
	}
}

func TestDiscoveryService_Register(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		ID:          "service-1",
		Name:        "test-service",
		Address:     "localhost",
		Port:        8080,
		ServiceType: "api",
	}

	err := ds.Register(instance)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
}

func TestDiscoveryService_Register_MissingID(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}

	err := ds.Register(instance)
	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestDiscoveryService_Deregister(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		ID:      "deregister-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}
	ds.Register(instance)

	err := ds.Deregister("deregister-1")
	if err != nil {
		t.Fatalf("Failed to deregister: %v", err)
	}

	_, err = ds.GetInstance("deregister-1")
	if err == nil {
		t.Error("Expected error for deregistered instance")
	}
}

func TestDiscoveryService_GetInstance(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		ID:      "get-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}
	ds.Register(instance)

	retrieved, err := ds.GetInstance("get-1")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.Name != "test-service" {
		t.Error("Service name mismatch")
	}
}

func TestDiscoveryService_GetInstance_NotFound(t *testing.T) {
	ds := service.NewDiscoveryService()

	_, err := ds.GetInstance("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent instance")
	}
}

func TestDiscoveryService_GetInstances(t *testing.T) {
	ds := service.NewDiscoveryService()

	ds.Register(&service.ServiceInstance{
		ID:   "list-1",
		Name: "list-service",
		Address: "localhost",
		Port: 8080,
	})

	ds.Register(&service.ServiceInstance{
		ID:   "list-2",
		Name: "list-service",
		Address: "localhost",
		Port: 8081,
	})

	ds.Register(&service.ServiceInstance{
		ID:   "other-1",
		Name: "other-service",
		Address: "localhost",
		Port: 8082,
	})

	instances := ds.GetInstances("list-service")
	if len(instances) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(instances))
	}
}

func TestDiscoveryService_GetAllInstances(t *testing.T) {
	ds := service.NewDiscoveryService()

	ds.Register(&service.ServiceInstance{
		ID:   "all-1",
		Name: "service-1",
		Address: "localhost",
		Port: 8080,
	})

	ds.Register(&service.ServiceInstance{
		ID:   "all-2",
		Name: "service-2",
		Address: "localhost",
		Port: 8081,
	})

	instances := ds.GetAllInstances()
	if len(instances) < 2 {
		t.Errorf("Expected at least 2 instances, got %d", len(instances))
	}
}

func TestDiscoveryService_Heartbeat(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		ID:      "heartbeat-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}
	ds.Register(instance)

	// Initial heartbeat
	time.Sleep(100 * time.Millisecond)

	err := ds.Heartbeat("heartbeat-1")
	if err != nil {
		t.Fatalf("Failed to send heartbeat: %v", err)
	}
}

func TestDiscoveryService_Heartbeat_NotFound(t *testing.T) {
	ds := service.NewDiscoveryService()

	err := ds.Heartbeat("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent instance")
	}
}

func TestDiscoveryService_SetMetadata(t *testing.T) {
	ds := service.NewDiscoveryService()

	instance := &service.ServiceInstance{
		ID:      "meta-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}
	ds.Register(instance)

	err := ds.SetMetadata("meta-1", map[string]interface{}{
		"version": "1.0.0",
		"env":     "production",
	})

	if err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}

	retrieved, _ := ds.GetInstance("meta-1")
	if retrieved.Metadata["version"] != "1.0.0" {
		t.Error("Metadata not set correctly")
	}
}

func TestDiscoveryService_CleanupExpired(t *testing.T) {
	ds := service.NewDiscoveryService()

	// Register with very short TTL
	instance := &service.ServiceInstance{
		ID:      "expired-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}
	ds.Register(instance)

	// Manually set last heartbeat to past
	past := time.Now().Add(-10 * time.Minute)
	ds.SetLastHeartbeat("expired-1", past)

	// Cleanup instances with no heartbeat for 5 minutes
	removed := ds.CleanupExpired(5 * time.Minute)
	if removed < 1 {
		t.Error("Expected at least 1 instance to be cleaned up")
	}
}

func TestDiscoveryService_GetServiceNames(t *testing.T) {
	ds := service.NewDiscoveryService()

	ds.Register(&service.ServiceInstance{
		ID:   "name-1",
		Name: "service-alpha",
		Address: "localhost",
		Port: 8080,
	})

	ds.Register(&service.ServiceInstance{
		ID:   "name-2",
		Name: "service-beta",
		Address: "localhost",
		Port: 8081,
	})

	names := ds.GetServiceNames()
	if len(names) < 2 {
		t.Errorf("Expected at least 2 service names, got %d", len(names))
	}
}

func TestDiscoveryService_ServiceInstanceToJSON(t *testing.T) {
	instance := &service.ServiceInstance{
		ID:        "json-1",
		Name:      "test-service",
		Address:   "localhost",
		Port:      8080,
		Status:    "healthy",
		CreatedAt: time.Now(),
	}

	data, err := instance.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}