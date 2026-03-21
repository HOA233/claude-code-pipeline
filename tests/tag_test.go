package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Tag Service Tests

func TestTagService_New(t *testing.T) {
	ts := service.NewTagService()
	if ts == nil {
		t.Fatal("Expected non-nil tag service")
	}
}

func TestTagService_CreateTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "environment",
		Value: "production",
	}

	err := ts.CreateTag(tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if tag.UsageCount != 0 {
		t.Errorf("Expected usage count 0, got %d", tag.UsageCount)
	}
}

func TestTagService_CreateTag_MissingKey(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Value: "production",
	}

	err := ts.CreateTag(tag)
	if err == nil {
		t.Error("Expected error for missing key")
	}
}

func TestTagService_GetTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "team",
		Value: "backend",
	}
	ts.CreateTag(tag)

	retrieved, err := ts.GetTag(tag.ID)
	if err != nil {
		t.Fatalf("Failed to get tag: %v", err)
	}

	if retrieved.Value != "backend" {
		t.Error("Tag value mismatch")
	}
}

func TestTagService_GetTag_NotFound(t *testing.T) {
	ts := service.NewTagService()

	_, err := ts.GetTag("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tag")
	}
}

func TestTagService_GetTagByKey(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "region",
		Value: "us-west-1",
	}
	ts.CreateTag(tag)

	retrieved, err := ts.GetTagByKey("region", "us-west-1")
	if err != nil {
		t.Fatalf("Failed to get tag by key: %v", err)
	}

	if retrieved.Key != "region" {
		t.Error("Tag key mismatch")
	}
}

func TestTagService_ListTags(t *testing.T) {
	ts := service.NewTagService()

	ts.CreateTag(&service.Tag{
		Key:      "env",
		Value:    "dev",
		Category: "environment",
	})

	ts.CreateTag(&service.Tag{
		Key:      "env",
		Value:    "prod",
		Category: "environment",
	})

	ts.CreateTag(&service.Tag{
		Key:      "owner",
		Value:    "team-a",
		Category: "ownership",
	})

	tags := ts.ListTags("environment")
	if len(tags) < 2 {
		t.Errorf("Expected at least 2 environment tags, got %d", len(tags))
	}

	allTags := ts.ListTags("")
	if len(allTags) < 3 {
		t.Errorf("Expected at least 3 total tags, got %d", len(allTags))
	}
}

func TestTagService_UpdateTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "status",
		Value: "active",
	}
	ts.CreateTag(tag)

	updated, err := ts.UpdateTag(tag.ID, map[string]interface{}{
		"value":       "inactive",
		"description": "Status tag",
		"color":       "#ff0000",
	})

	if err != nil {
		t.Fatalf("Failed to update tag: %v", err)
	}

	if updated.Value != "inactive" {
		t.Error("Value not updated")
	}

	if updated.Description != "Status tag" {
		t.Error("Description not updated")
	}

	if updated.Color != "#ff0000" {
		t.Error("Color not updated")
	}
}

func TestTagService_UpdateTag_ReadOnly(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:      "system",
		Value:    "protected",
		ReadOnly: true,
	}
	ts.CreateTag(tag)

	_, err := ts.UpdateTag(tag.ID, map[string]interface{}{
		"value": "modified",
	})

	if err == nil {
		t.Error("Expected error for read-only tag")
	}
}

func TestTagService_DeleteTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "temp",
		Value: "delete-me",
	}
	ts.CreateTag(tag)

	err := ts.DeleteTag(tag.ID)
	if err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}

	_, err = ts.GetTag(tag.ID)
	if err == nil {
		t.Error("Expected error for deleted tag")
	}
}

func TestTagService_DeleteTag_ReadOnly(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:      "readonly",
		Value:    "protected",
		ReadOnly: true,
	}
	ts.CreateTag(tag)

	err := ts.DeleteTag(tag.ID)
	if err == nil {
		t.Error("Expected error for deleting read-only tag")
	}
}

func TestTagService_BindTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "app",
		Value: "api-server",
	}
	ts.CreateTag(tag)

	err := ts.BindTag(tag.ID, "resource-1", "pipeline", "tenant-1", "user-1")
	if err != nil {
		t.Fatalf("Failed to bind tag: %v", err)
	}

	updated, _ := ts.GetTag(tag.ID)
	if updated.UsageCount != 1 {
		t.Errorf("Expected usage count 1, got %d", updated.UsageCount)
	}
}

func TestTagService_BindTag_TagNotFound(t *testing.T) {
	ts := service.NewTagService()

	err := ts.BindTag("nonexistent", "resource-1", "pipeline", "tenant-1", "user-1")
	if err == nil {
		t.Error("Expected error for nonexistent tag")
	}
}

func TestTagService_UnbindTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{
		Key:   "service",
		Value: "auth",
	}
	ts.CreateTag(tag)
	ts.BindTag(tag.ID, "resource-1", "pipeline", "tenant-1", "user-1")

	err := ts.UnbindTag(tag.ID, "resource-1")
	if err != nil {
		t.Fatalf("Failed to unbind tag: %v", err)
	}

	updated, _ := ts.GetTag(tag.ID)
	if updated.UsageCount != 0 {
		t.Errorf("Expected usage count 0, got %d", updated.UsageCount)
	}
}

func TestTagService_GetResourceTags(t *testing.T) {
	ts := service.NewTagService()

	tag1 := &service.Tag{Key: "env", Value: "prod"}
	tag2 := &service.Tag{Key: "team", Value: "platform"}
	ts.CreateTag(tag1)
	ts.CreateTag(tag2)

	ts.BindTag(tag1.ID, "resource-1", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag2.ID, "resource-1", "pipeline", "tenant-1", "user-1")

	bindings := ts.GetResourceTags("resource-1")
	if len(bindings) < 2 {
		t.Errorf("Expected at least 2 bindings, got %d", len(bindings))
	}
}

func TestTagService_GetResourcesByTag(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{Key: "critical", Value: "true"}
	ts.CreateTag(tag)

	ts.BindTag(tag.ID, "resource-1", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag.ID, "resource-2", "pipeline", "tenant-1", "user-1")

	bindings := ts.GetResourcesByTag(tag.ID)
	if len(bindings) < 2 {
		t.Errorf("Expected at least 2 bindings, got %d", len(bindings))
	}
}

func TestTagService_GetResourcesByTagKey(t *testing.T) {
	ts := service.NewTagService()

	tag1 := &service.Tag{Key: "priority", Value: "high"}
	tag2 := &service.Tag{Key: "priority", Value: "low"}
	ts.CreateTag(tag1)
	ts.CreateTag(tag2)

	ts.BindTag(tag1.ID, "resource-1", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag2.ID, "resource-2", "pipeline", "tenant-1", "user-1")

	bindings := ts.GetResourcesByTagKey("priority")
	if len(bindings) < 2 {
		t.Errorf("Expected at least 2 bindings, got %d", len(bindings))
	}
}

func TestTagService_SearchTags(t *testing.T) {
	ts := service.NewTagService()

	ts.CreateTag(&service.Tag{Key: "environment", Value: "production"})
	ts.CreateTag(&service.Tag{Key: "env", Value: "staging"})
	ts.CreateTag(&service.Tag{Key: "team", Value: "backend"})

	results := ts.SearchTags("env")
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}
}

func TestTagService_GetTagStats(t *testing.T) {
	ts := service.NewTagService()

	tag1 := &service.Tag{Key: "env", Value: "prod", Category: "environment"}
	tag2 := &service.Tag{Key: "team", Value: "api", Category: "ownership"}
	ts.CreateTag(tag1)
	ts.CreateTag(tag2)

	ts.BindTag(tag1.ID, "r1", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag1.ID, "r2", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag2.ID, "r3", "pipeline", "tenant-1", "user-1")

	stats := ts.GetTagStats()

	if stats.TotalTags < 2 {
		t.Errorf("Expected at least 2 tags, got %d", stats.TotalTags)
	}

	if stats.TotalBindings < 3 {
		t.Errorf("Expected at least 3 bindings, got %d", stats.TotalBindings)
	}
}

func TestTagService_BulkTagResources(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{Key: "bulk", Value: "test"}
	ts.CreateTag(tag)

	resources := []struct {
		ID   string
		Type string
	}{
		{ID: "r1", Type: "pipeline"},
		{ID: "r2", Type: "pipeline"},
		{ID: "r3", Type: "pipeline"},
	}

	count, err := ts.BulkTagResources(tag.ID, resources, "tenant-1", "user-1")
	if err != nil {
		t.Fatalf("Failed to bulk tag: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 bindings, got %d", count)
	}
}

func TestTagService_CopyTags(t *testing.T) {
	ts := service.NewTagService()

	tag1 := &service.Tag{Key: "copy1", Value: "val1"}
	tag2 := &service.Tag{Key: "copy2", Value: "val2"}
	ts.CreateTag(tag1)
	ts.CreateTag(tag2)

	ts.BindTag(tag1.ID, "source", "pipeline", "tenant-1", "user-1")
	ts.BindTag(tag2.ID, "source", "pipeline", "tenant-1", "user-1")

	count, err := ts.CopyTags("source", "target", "tenant-1", "user-1")
	if err != nil {
		t.Fatalf("Failed to copy tags: %v", err)
	}

	if count < 2 {
		t.Errorf("Expected at least 2 tags copied, got %d", count)
	}
}

func TestTagService_ValidateTags(t *testing.T) {
	ts := service.NewTagService()

	tag := &service.Tag{Key: "required", Value: "value"}
	ts.CreateTag(tag)

	bindings := []service.TagBinding{
		{TagKey: "required", TagValue: "value"},
	}

	errors := ts.ValidateTags("pipeline", bindings)
	if len(errors) != 0 {
		t.Errorf("Unexpected validation errors: %v", errors)
	}
}

func TestTagService_TagToJSON(t *testing.T) {
	tag := &service.Tag{
		ID:        "tag-1",
		Key:       "env",
		Value:     "prod",
		CreatedAt: time.Now(),
	}

	data, err := tag.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestTagService_TagBindingToJSON(t *testing.T) {
	binding := service.TagBinding{
		ID:           "binding-1",
		TagID:        "tag-1",
		TagKey:       "env",
		TagValue:     "prod",
		ResourceID:   "res-1",
		ResourceType: "pipeline",
		TenantID:     "tenant-1",
		CreatedAt:    time.Now(),
	}

	data, err := binding.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}