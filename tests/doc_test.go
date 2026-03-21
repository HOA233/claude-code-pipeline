package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

func TestDocService_New(t *testing.T) {
	doc := service.NewDocService()
	if doc == nil {
		t.Fatal("Expected non-nil doc service")
	}
}

func TestDocService_RegisterAPI(t *testing.T) {
	doc := service.NewDocService()

	err := doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/test",
		Summary: "Test endpoint",
	})

	if err != nil {
		t.Fatalf("Failed to register API: %v", err)
	}
}

func TestDocService_RegisterAPI_MissingMethod(t *testing.T) {
	doc := service.NewDocService()

	err := doc.RegisterAPI(&service.APIEndpoint{
		Path: "/api/test",
	})

	if err == nil {
		t.Error("Expected error for missing method")
	}
}

func TestDocService_RegisterAPI_MissingPath(t *testing.T) {
	doc := service.NewDocService()

	err := doc.RegisterAPI(&service.APIEndpoint{
		Method: "GET",
	})

	if err == nil {
		t.Error("Expected error for missing path")
	}
}

func TestDocService_GetAPI(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/get-test",
		Summary: "Get test",
	})

	api, err := doc.GetAPI("GET", "/api/get-test")
	if err != nil {
		t.Fatalf("Failed to get API: %v", err)
	}

	if api.Summary != "Get test" {
		t.Error("API summary mismatch")
	}
}

func TestDocService_GetAPI_NotFound(t *testing.T) {
	doc := service.NewDocService()

	_, err := doc.GetAPI("GET", "/nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent API")
	}
}

func TestDocService_ListAPIs(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/list1",
		Summary: "List 1",
		Tags:    []string{"test"},
	})

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/list2",
		Summary: "List 2",
		Tags:    []string{"test"},
	})

	apis := doc.ListAPIs("test")
	if len(apis) < 2 {
		t.Errorf("Expected at least 2 APIs, got %d", len(apis))
	}
}

func TestDocService_DeleteAPI(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "DELETE",
		Path:    "/api/delete-test",
		Summary: "Delete test",
	})

	err := doc.DeleteAPI("DELETE", "/api/delete-test")
	if err != nil {
		t.Fatalf("Failed to delete API: %v", err)
	}

	_, err = doc.GetAPI("DELETE", "/api/delete-test")
	if err == nil {
		t.Error("Expected error for deleted API")
	}
}

func TestDocService_CreateGuide(t *testing.T) {
	doc := service.NewDocService()

	err := doc.CreateGuide(&service.Guide{
		ID:      "getting-started",
		Title:   "Getting Started",
		Content: "This is a guide",
	})

	if err != nil {
		t.Fatalf("Failed to create guide: %v", err)
	}
}

func TestDocService_CreateGuide_MissingID(t *testing.T) {
	doc := service.NewDocService()

	err := doc.CreateGuide(&service.Guide{
		Title: "No ID",
	})

	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestDocService_GetGuide(t *testing.T) {
	doc := service.NewDocService()

	doc.CreateGuide(&service.Guide{
		ID:      "get-guide-test",
		Title:   "Get Guide Test",
		Content: "Content",
	})

	guide, err := doc.GetGuide("get-guide-test")
	if err != nil {
		t.Fatalf("Failed to get guide: %v", err)
	}

	if guide.Title != "Get Guide Test" {
		t.Error("Guide title mismatch")
	}
}

func TestDocService_GetGuide_NotFound(t *testing.T) {
	doc := service.NewDocService()

	_, err := doc.GetGuide("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent guide")
	}
}

func TestDocService_ListGuides(t *testing.T) {
	doc := service.NewDocService()

	doc.CreateGuide(&service.Guide{
		ID:       "guide1",
		Title:    "Guide 1",
		Category: "test",
	})

	doc.CreateGuide(&service.Guide{
		ID:       "guide2",
		Title:    "Guide 2",
		Category: "test",
	})

	guides := doc.ListGuides("test")
	if len(guides) < 2 {
		t.Errorf("Expected at least 2 guides, got %d", len(guides))
	}
}

func TestDocService_DeleteGuide(t *testing.T) {
	doc := service.NewDocService()

	doc.CreateGuide(&service.Guide{
		ID:      "delete-guide",
		Title:   "Delete Guide",
	})

	err := doc.DeleteGuide("delete-guide")
	if err != nil {
		t.Fatalf("Failed to delete guide: %v", err)
	}

	_, err = doc.GetGuide("delete-guide")
	if err == nil {
		t.Error("Expected error for deleted guide")
	}
}

func TestDocService_GenerateOpenAPI(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/openapi-test",
		Summary: "OpenAPI Test",
		Tags:    []string{"test"},
		Parameters: []service.Parameter{
			{Name: "id", In: "query", Type: "string", Required: true},
		},
		Responses: []service.Response{
			{Code: 200, Description: "Success"},
		},
	})

	openapi, err := doc.GenerateOpenAPI(service.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("Failed to generate OpenAPI: %v", err)
	}

	if len(openapi) == 0 {
		t.Error("Expected non-empty OpenAPI spec")
	}
}

func TestDocService_GenerateMarkdown(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/md-test",
		Summary: "Markdown Test",
		Tags:    []string{"test"},
		Parameters: []service.Parameter{
			{Name: "id", In: "path", Type: "string", Required: true, Description: "ID"},
		},
		Responses: []service.Response{
			{Code: 200, Description: "OK"},
		},
	})

	md := doc.GenerateMarkdown()

	if md == "" {
		t.Error("Expected non-empty markdown")
	}

	if !contains(md, "Markdown Test") {
		t.Error("Expected markdown to contain summary")
	}
}

func TestDocService_GenerateHTML(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "POST",
		Path:    "/api/html-test",
		Summary: "HTML Test",
		Tags:    []string{"test"},
		AuthRequired: true,
	})

	html := doc.GenerateHTML()

	if html == "" {
		t.Error("Expected non-empty HTML")
	}

	if !contains(html, "HTML Test") {
		t.Error("Expected HTML to contain summary")
	}
}

func TestDocService_APIWithRequestBody(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "POST",
		Path:    "/api/with-body",
		Summary: "With Body",
		RequestBody: &service.RequestBody{
			Required:    true,
			ContentType: "application/json",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{"type": "string"},
				},
			},
		},
	})

	api, _ := doc.GetAPI("POST", "/api/with-body")
	if api.RequestBody == nil {
		t.Error("Expected request body")
	}
}

func TestDocService_APIWithExamples(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/with-examples",
		Summary: "With Examples",
		Examples: []service.Example{
			{
				Name:     "Example 1",
				Request:  "GET /api/test",
				Response: `{"status": "ok"}`,
			},
		},
	})

	api, _ := doc.GetAPI("GET", "/api/with-examples")
	if len(api.Examples) != 1 {
		t.Error("Expected 1 example")
	}
}

func TestDocService_DeprecatedAPI(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:     "GET",
		Path:       "/api/deprecated",
		Summary:    "Deprecated",
		Deprecated: true,
	})

	api, _ := doc.GetAPI("GET", "/api/deprecated")
	if !api.Deprecated {
		t.Error("Expected API to be deprecated")
	}
}

func TestDocService_AuthRequired(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:       "GET",
		Path:         "/api/protected",
		Summary:      "Protected",
		AuthRequired: true,
	})

	api, _ := doc.GetAPI("GET", "/api/protected")
	if !api.AuthRequired {
		t.Error("Expected API to require auth")
	}
}

func TestDocService_GetStats(t *testing.T) {
	doc := service.NewDocService()

	doc.RegisterAPI(&service.APIEndpoint{
		Method:  "GET",
		Path:    "/api/stats-test",
		Summary: "Stats",
		Tags:    []string{"test"},
	})

	doc.CreateGuide(&service.Guide{
		ID:      "stats-guide",
		Title:   "Stats Guide",
	})

	stats := doc.GetStats()

	if stats["total_apis"].(int) < 1 {
		t.Error("Expected at least 1 API")
	}
	if stats["total_guides"].(int) < 1 {
		t.Error("Expected at least 1 guide")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}