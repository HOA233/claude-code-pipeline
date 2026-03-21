package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Template Service Tests

func TestTemplateService_New(t *testing.T) {
	ts := service.NewTemplateService()
	if ts == nil {
		t.Fatal("Expected non-nil template service")
	}
}

func TestTemplateService_Create(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:          "template-1",
		Name:        "Test Template",
		Description: "A test template",
		Content:     "Hello {{.Name}}!",
		Version:     "1.0.0",
		TenantID:    "tenant-1",
	}

	err := ts.Create(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
}

func TestTemplateService_Create_MissingName(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "no-name",
		Content:  "test",
		TenantID: "tenant-1",
	}

	err := ts.Create(template)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestTemplateService_Get(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "get-template",
		Name:     "Get Template",
		Content:  "test content",
		TenantID: "tenant-get",
	}
	ts.Create(template)

	retrieved, err := ts.Get("get-template")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.Name != "Get Template" {
		t.Error("Template name mismatch")
	}
}

func TestTemplateService_Get_NotFound(t *testing.T) {
	ts := service.NewTemplateService()

	_, err := ts.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
}

func TestTemplateService_List(t *testing.T) {
	ts := service.NewTemplateService()

	ts.Create(&service.Template{
		ID:       "list-1",
		Name:     "List 1",
		Content:  "test",
		TenantID: "tenant-list",
	})

	ts.Create(&service.Template{
		ID:       "list-2",
		Name:     "List 2",
		Content:  "test",
		TenantID: "tenant-list",
	})

	templates := ts.List("tenant-list")
	if len(templates) < 2 {
		t.Errorf("Expected at least 2 templates, got %d", len(templates))
	}
}

func TestTemplateService_Update(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "update-template",
		Name:     "Update Template",
		Content:  "original content",
		Version:  "1.0.0",
		TenantID: "tenant-update",
	}
	ts.Create(template)

	err := ts.Update("update-template", &service.Template{
		Content: "updated content",
		Version: "2.0.0",
	})

	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	retrieved, _ := ts.Get("update-template")
	if retrieved.Content != "updated content" {
		t.Error("Content not updated")
	}
}

func TestTemplateService_Delete(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "delete-template",
		Name:     "Delete Template",
		Content:  "test",
		TenantID: "tenant-delete",
	}
	ts.Create(template)

	err := ts.Delete("delete-template")
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	_, err = ts.Get("delete-template")
	if err == nil {
		t.Error("Expected error for deleted template")
	}
}

func TestTemplateService_Render(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "render-template",
		Name:     "Render Template",
		Content:  "Hello {{.Name}}, welcome to {{.Place}}!",
		TenantID: "tenant-render",
	}
	ts.Create(template)

	data := map[string]interface{}{
		"Name":  "World",
		"Place": "Earth",
	}

	rendered, err := ts.Render("render-template", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Hello World, welcome to Earth!"
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

func TestTemplateService_Render_InvalidTemplate(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "invalid-template",
		Name:     "Invalid Template",
		Content:  "Hello {{.Name", // Invalid syntax
		TenantID: "tenant-invalid",
	}
	ts.Create(template)

	_, err := ts.Render("invalid-template", map[string]interface{}{"Name": "Test"})
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

func TestTemplateService_GetVersion(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "version-template",
		Name:     "Version Template",
		Content:  "test",
		Version:  "1.0.0",
		TenantID: "tenant-version",
	}
	ts.Create(template)

	version := ts.GetVersion("version-template")
	if version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", version)
	}
}

func TestTemplateService_SetVersion(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "setversion-template",
		Name:     "Set Version Template",
		Content:  "test",
		Version:  "1.0.0",
		TenantID: "tenant-setversion",
	}
	ts.Create(template)

	err := ts.SetVersion("setversion-template", "2.0.0")
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	retrieved, _ := ts.Get("setversion-template")
	if retrieved.Version != "2.0.0" {
		t.Error("Version not updated")
	}
}

func TestTemplateService_Clone(t *testing.T) {
	ts := service.NewTemplateService()

	template := &service.Template{
		ID:       "clone-source",
		Name:     "Clone Source",
		Content:  "clone content",
		TenantID: "tenant-clone",
	}
	ts.Create(template)

	cloned, err := ts.Clone("clone-source", "cloned-template", "tenant-clone")
	if err != nil {
		t.Fatalf("Failed to clone template: %v", err)
	}

	if cloned.ID == template.ID {
		t.Error("Cloned template should have different ID")
	}

	if cloned.Name != "cloned-template" {
		t.Error("Cloned template name mismatch")
	}
}

func TestTemplateService_GetStats(t *testing.T) {
	ts := service.NewTemplateService()

	ts.Create(&service.Template{
		ID:       "stats-1",
		Name:     "Stats 1",
		Content:  "test",
		TenantID: "tenant-stats",
	})

	ts.Create(&service.Template{
		ID:       "stats-2",
		Name:     "Stats 2",
		Content:  "test",
		TenantID: "tenant-stats",
	})

	stats := ts.GetStats("tenant-stats")

	if stats.TotalTemplates < 2 {
		t.Errorf("Expected at least 2 templates, got %d", stats.TotalTemplates)
	}
}

func TestTemplateService_TemplateToJSON(t *testing.T) {
	template := &service.Template{
		ID:          "json-template",
		Name:        "JSON Template",
		Description: "Test description",
		Content:     "Hello {{.Name}}",
		Version:     "1.0.0",
		TenantID:    "tenant-1",
		CreatedAt:   time.Now(),
	}

	data, err := template.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}