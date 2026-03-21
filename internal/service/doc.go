package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// DocService manages API documentation
type DocService struct {
	mu     sync.RWMutex
	apis   map[string]*APIEndpoint
	guides map[string]*Guide
}

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	ID           string          `json:"id"`
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	Summary      string          `json:"summary"`
	Description  string          `json:"description,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	Parameters   []Parameter     `json:"parameters,omitempty"`
	RequestBody  *RequestBody    `json:"request_body,omitempty"`
	Responses    []Response      `json:"responses"`
	AuthRequired bool            `json:"auth_required"`
	Deprecated   bool            `json:"deprecated,omitempty"`
	Examples     []Example       `json:"examples,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // path, query, header, cookie
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	ContentType string      `json:"content_type"`
	Schema      interface{} `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Response represents an API response
type Response struct {
	Code        int         `json:"code"`
	Description string      `json:"description"`
	ContentType string      `json:"content_type,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Example represents an API example
type Example struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Request     string      `json:"request,omitempty"`
	Response    string      `json:"response,omitempty"`
}

// Guide represents a documentation guide
type Guide struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags,omitempty"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewDocService creates a new documentation service
func NewDocService() *DocService {
	return &DocService{
		apis:   make(map[string]*APIEndpoint),
		guides: make(map[string]*Guide),
	}
}

// RegisterAPI registers an API endpoint
func (s *DocService) RegisterAPI(endpoint *APIEndpoint) error {
	if endpoint.Method == "" {
		return fmt.Errorf("method is required")
	}
	if endpoint.Path == "" {
		return fmt.Errorf("path is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateAPIID(endpoint.Method, endpoint.Path)
	now := time.Now()

	if existing, exists := s.apis[id]; exists {
		endpoint.CreatedAt = existing.CreatedAt
	} else {
		endpoint.CreatedAt = now
	}
	endpoint.UpdatedAt = now
	endpoint.ID = id

	s.apis[id] = endpoint
	return nil
}

// GetAPI gets an API endpoint
func (s *DocService) GetAPI(method, path string) (*APIEndpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id := s.generateAPIID(method, path)
	api, exists := s.apis[id]
	if !exists {
		return nil, fmt.Errorf("API not found: %s %s", method, path)
	}
	return api, nil
}

// ListAPIs lists all API endpoints
func (s *DocService) ListAPIs(tag string) []*APIEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apis := make([]*APIEndpoint, 0, len(s.apis))
	for _, api := range s.apis {
		if tag == "" || containsTag(api.Tags, tag) {
			apis = append(apis, api)
		}
	}

	// Sort by path
	sort.Slice(apis, func(i, j int) bool {
		if apis[i].Path != apis[j].Path {
			return apis[i].Path < apis[j].Path
		}
		return methodOrder(apis[i].Method) < methodOrder(apis[j].Method)
	})

	return apis
}

// DeleteAPI deletes an API endpoint
func (s *DocService) DeleteAPI(method, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateAPIID(method, path)
	if _, exists := s.apis[id]; !exists {
		return fmt.Errorf("API not found: %s %s", method, path)
	}

	delete(s.apis, id)
	return nil
}

// CreateGuide creates a documentation guide
func (s *DocService) CreateGuide(guide *Guide) error {
	if guide.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if guide.Title == "" {
		return fmt.Errorf("title is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	guide.CreatedAt = now
	guide.UpdatedAt = now

	s.guides[guide.ID] = guide
	return nil
}

// GetGuide gets a guide by ID
func (s *DocService) GetGuide(id string) (*Guide, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guide, exists := s.guides[id]
	if !exists {
		return nil, fmt.Errorf("guide not found: %s", id)
	}
	return guide, nil
}

// ListGuides lists all guides
func (s *DocService) ListGuides(category string) []*Guide {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guides := make([]*Guide, 0, len(s.guides))
	for _, guide := range s.guides {
		if category == "" || guide.Category == category {
			guides = append(guides, guide)
		}
	}

	// Sort by order
	sort.Slice(guides, func(i, j int) bool {
		return guides[i].Order < guides[j].Order
	})

	return guides
}

// DeleteGuide deletes a guide
func (s *DocService) DeleteGuide(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.guides[id]; !exists {
		return fmt.Errorf("guide not found: %s", id)
	}

	delete(s.guides, id)
	return nil
}

// GenerateOpenAPI generates OpenAPI 3.0 specification
func (s *DocService) GenerateOpenAPI(info OpenAPIInfo) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	openapi := map[string]interface{}{
		"openapi": "3.0.0",
		"info":    info,
		"paths":   s.buildPaths(),
		"tags":    s.buildTags(),
	}

	return json.MarshalIndent(openapi, "", "  ")
}

// OpenAPIInfo represents OpenAPI info section
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// GenerateMarkdown generates markdown documentation
func (s *DocService) GenerateMarkdown() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("# API Documentation\n\n")

	// Group by tags
	tagGroups := make(map[string][]*APIEndpoint)
	for _, api := range s.apis {
		for _, tag := range api.Tags {
			if len(api.Tags) > 0 {
				tagGroups[tag] = append(tagGroups[tag], api)
			}
		}
		if len(api.Tags) == 0 {
			tagGroups["General"] = append(tagGroups["General"], api)
		}
	}

	// Sort tags
	tags := make([]string, 0, len(tagGroups))
	for tag := range tagGroups {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		sb.WriteString(fmt.Sprintf("## %s\n\n", tag))

		for _, api := range tagGroups[tag] {
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", api.Method, api.Path))
			sb.WriteString(fmt.Sprintf("**%s**\n\n", api.Summary))

			if api.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", api.Description))
			}

			// Parameters
			if len(api.Parameters) > 0 {
				sb.WriteString("**Parameters:**\n\n")
				sb.WriteString("| Name | In | Type | Required | Description |\n")
				sb.WriteString("|------|-----|------|----------|-------------|\n")
				for _, p := range api.Parameters {
					required := ""
					if p.Required {
						required = "Yes"
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
						p.Name, p.In, p.Type, required, p.Description))
				}
				sb.WriteString("\n")
			}

			// Responses
			if len(api.Responses) > 0 {
				sb.WriteString("**Responses:**\n\n")
				for _, r := range api.Responses {
					sb.WriteString(fmt.Sprintf("- `%d` - %s\n", r.Code, r.Description))
				}
				sb.WriteString("\n")
			}

			if api.AuthRequired {
				sb.WriteString("🔒 **Authentication Required**\n\n")
			}

			if api.Deprecated {
				sb.WriteString("⚠️ **Deprecated**\n\n")
			}
		}
	}

	return sb.String()
}

// GenerateHTML generates HTML documentation
func (s *DocService) GenerateHTML() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .endpoint { border: 1px solid #ddd; border-radius: 8px; margin-bottom: 16px; padding: 16px; }
        .method { font-weight: bold; padding: 4px 8px; border-radius: 4px; margin-right: 8px; }
        .get { background: #61affe; color: white; }
        .post { background: #49cc90; color: white; }
        .put { background: #fca130; color: white; }
        .delete { background: #f93e3e; color: white; }
        .path { font-family: monospace; font-size: 1.1em; }
        table { width: 100%; border-collapse: collapse; margin-top: 8px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #f5f5f5; }
        .deprecated { opacity: 0.6; }
        .auth-required { color: #f93e3e; }
    </style>
</head>
<body>
    <h1>API Documentation</h1>
`)

	apis := s.ListAPIs("")
	for _, api := range apis {
		deprecatedClass := ""
		if api.Deprecated {
			deprecatedClass = " deprecated"
		}

		sb.WriteString(fmt.Sprintf(`    <div class="endpoint%s">
        <h3><span class="method %s">%s</span> <span class="path">%s</span></h3>
        <p>%s</p>
`, deprecatedClass, strings.ToLower(api.Method), api.Method, api.Path, api.Summary))

		if len(api.Parameters) > 0 {
			sb.WriteString(`        <h4>Parameters</h4>
        <table>
            <tr><th>Name</th><th>In</th><th>Type</th><th>Required</th><th>Description</th></tr>
`)
			for _, p := range api.Parameters {
				required := "No"
				if p.Required {
					required = "Yes"
				}
				sb.WriteString(fmt.Sprintf("            <tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
					p.Name, p.In, p.Type, required, p.Description))
			}
			sb.WriteString("        </table>\n")
		}

		if len(api.Responses) > 0 {
			sb.WriteString("        <h4>Responses</h4>\n        <ul>\n")
			for _, r := range api.Responses {
				sb.WriteString(fmt.Sprintf("            <li><code>%d</code> - %s</li>\n", r.Code, r.Description))
			}
			sb.WriteString("        </ul>\n")
		}

		if api.AuthRequired {
			sb.WriteString("        <p class=\"auth-required\">🔒 Authentication Required</p>\n")
		}

		sb.WriteString("    </div>\n")
	}

	sb.WriteString(`</body>
</html>`)

	return sb.String()
}

// GetStats returns documentation statistics
func (s *DocService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	methods := make(map[string]int)
	for _, api := range s.apis {
		methods[api.Method]++
	}

	tags := make(map[string]int)
	for _, api := range s.apis {
		for _, tag := range api.Tags {
			tags[tag]++
		}
	}

	return map[string]interface{}{
		"total_apis":   len(s.apis),
		"total_guides": len(s.guides),
		"by_method":    methods,
		"by_tag":       tags,
	}
}

func (s *DocService) generateAPIID(method, path string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(method), strings.ReplaceAll(path, "/", "-"))
}

func (s *DocService) buildPaths() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, api := range s.apis {
		if _, exists := paths[api.Path]; !exists {
			paths[api.Path] = make(map[string]interface{})
		}

		pathItem := paths[api.Path].(map[string]interface{})
		pathItem[strings.ToLower(api.Method)] = s.buildOperation(api)
	}

	return paths
}

func (s *DocService) buildOperation(api *APIEndpoint) map[string]interface{} {
	operation := map[string]interface{}{
		"summary":     api.Summary,
		"description": api.Description,
		"tags":        api.Tags,
		"deprecated":  api.Deprecated,
	}

	if len(api.Parameters) > 0 {
		params := make([]map[string]interface{}, len(api.Parameters))
		for i, p := range api.Parameters {
			params[i] = map[string]interface{}{
				"name":        p.Name,
				"in":          p.In,
				"required":    p.Required,
				"description": p.Description,
				"schema": map[string]interface{}{
					"type": p.Type,
				},
			}
		}
		operation["parameters"] = params
	}

	if api.RequestBody != nil {
		operation["requestBody"] = map[string]interface{}{
			"description": api.RequestBody.Description,
			"required":    api.RequestBody.Required,
			"content": map[string]interface{}{
				api.RequestBody.ContentType: map[string]interface{}{
					"schema":  api.RequestBody.Schema,
					"example": api.RequestBody.Example,
				},
			},
		}
	}

	responses := make(map[string]interface{})
	for _, r := range api.Responses {
		responses[fmt.Sprintf("%d", r.Code)] = map[string]interface{}{
			"description": r.Description,
		}
	}
	operation["responses"] = responses

	if api.AuthRequired {
		operation["security"] = []map[string]interface{}{
			{"bearerAuth": []string{}},
		}
	}

	return operation
}

func (s *DocService) buildTags() []map[string]interface{} {
	tagSet := make(map[string]bool)
	for _, api := range s.apis {
		for _, tag := range api.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]map[string]interface{}, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, map[string]interface{}{
			"name": tag,
		})
	}

	return tags
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func methodOrder(method string) int {
	switch strings.ToUpper(method) {
	case "GET":
		return 1
	case "POST":
		return 2
	case "PUT":
		return 3
	case "PATCH":
		return 4
	case "DELETE":
		return 5
	default:
		return 6
	}
}