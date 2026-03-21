package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Workspace Service Tests

func TestWorkspaceService_New(t *testing.T) {
	ws := service.NewWorkspaceService()
	if ws == nil {
		t.Fatal("Expected non-nil workspace service")
	}
}

func TestWorkspaceService_Create(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:          "workspace-1",
		Name:        "Test Workspace",
		Description: "A test workspace",
		TenantID:    "tenant-1",
		Path:        "/workspaces/test",
	}

	err := ws.Create(workspace)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
}

func TestWorkspaceService_Create_MissingName(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "no-name",
		TenantID: "tenant-1",
	}

	err := ws.Create(workspace)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestWorkspaceService_Get(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "get-workspace",
		Name:     "Get Workspace",
		TenantID: "tenant-get",
	}
	ws.Create(workspace)

	retrieved, err := ws.Get("get-workspace")
	if err != nil {
		t.Fatalf("Failed to get workspace: %v", err)
	}

	if retrieved.Name != "Get Workspace" {
		t.Error("Workspace name mismatch")
	}
}

func TestWorkspaceService_Get_NotFound(t *testing.T) {
	ws := service.NewWorkspaceService()

	_, err := ws.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent workspace")
	}
}

func TestWorkspaceService_List(t *testing.T) {
	ws := service.NewWorkspaceService()

	ws.Create(&service.Workspace{
		ID:       "list-1",
		Name:     "List 1",
		TenantID: "tenant-list",
	})

	ws.Create(&service.Workspace{
		ID:       "list-2",
		Name:     "List 2",
		TenantID: "tenant-list",
	})

	workspaces := ws.List("tenant-list")
	if len(workspaces) < 2 {
		t.Errorf("Expected at least 2 workspaces, got %d", len(workspaces))
	}
}

func TestWorkspaceService_Update(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:          "update-workspace",
		Name:        "Update Workspace",
		Description: "Original",
		TenantID:    "tenant-update",
	}
	ws.Create(workspace)

	err := ws.Update("update-workspace", map[string]interface{}{
		"description": "Updated",
		"name":        "Updated Name",
	})

	if err != nil {
		t.Fatalf("Failed to update workspace: %v", err)
	}

	retrieved, _ := ws.Get("update-workspace")
	if retrieved.Description != "Updated" {
		t.Error("Description not updated")
	}
}

func TestWorkspaceService_Delete(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "delete-workspace",
		Name:     "Delete Workspace",
		TenantID: "tenant-delete",
	}
	ws.Create(workspace)

	err := ws.Delete("delete-workspace")
	if err != nil {
		t.Fatalf("Failed to delete workspace: %v", err)
	}

	_, err = ws.Get("delete-workspace")
	if err == nil {
		t.Error("Expected error for deleted workspace")
	}
}

func TestWorkspaceService_Archive(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "archive-workspace",
		Name:     "Archive Workspace",
		TenantID: "tenant-archive",
	}
	ws.Create(workspace)

	err := ws.Archive("archive-workspace")
	if err != nil {
		t.Fatalf("Failed to archive workspace: %v", err)
	}

	retrieved, _ := ws.Get("archive-workspace")
	if retrieved.Status != "archived" {
		t.Error("Expected status to be archived")
	}
}

func TestWorkspaceService_Restore(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "restore-workspace",
		Name:     "Restore Workspace",
		TenantID: "tenant-restore",
		Status:   "archived",
	}
	ws.Create(workspace)

	err := ws.Restore("restore-workspace")
	if err != nil {
		t.Fatalf("Failed to restore workspace: %v", err)
	}

	retrieved, _ := ws.Get("restore-workspace")
	if retrieved.Status != "active" {
		t.Error("Expected status to be active")
	}
}

func TestWorkspaceService_AddMember(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "member-workspace",
		Name:     "Member Workspace",
		TenantID: "tenant-member",
	}
	ws.Create(workspace)

	err := ws.AddMember("member-workspace", "user-1", "admin")
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}
}

func TestWorkspaceService_RemoveMember(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "remove-member-workspace",
		Name:     "Remove Member Workspace",
		TenantID: "tenant-remove",
	}
	ws.Create(workspace)
	ws.AddMember("remove-member-workspace", "user-1", "admin")

	err := ws.RemoveMember("remove-member-workspace", "user-1")
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}
}

func TestWorkspaceService_GetMembers(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "members-workspace",
		Name:     "Members Workspace",
		TenantID: "tenant-members",
	}
	ws.Create(workspace)
	ws.AddMember("members-workspace", "user-1", "admin")
	ws.AddMember("members-workspace", "user-2", "member")

	members := ws.GetMembers("members-workspace")
	if len(members) < 2 {
		t.Errorf("Expected at least 2 members, got %d", len(members))
	}
}

func TestWorkspaceService_SetSettings(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "settings-workspace",
		Name:     "Settings Workspace",
		TenantID: "tenant-settings",
	}
	ws.Create(workspace)

	settings := map[string]interface{}{
		"auto_save": true,
		"theme":     "dark",
	}

	err := ws.SetSettings("settings-workspace", settings)
	if err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}
}

func TestWorkspaceService_GetSettings(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "get-settings-workspace",
		Name:     "Get Settings Workspace",
		TenantID: "tenant-getsettings",
	}
	ws.Create(workspace)
	ws.SetSettings("get-settings-workspace", map[string]interface{}{
		"key": "value",
	})

	settings := ws.GetSettings("get-settings-workspace")
	if settings == nil {
		t.Fatal("Expected settings")
	}

	if settings["key"] != "value" {
		t.Error("Settings value mismatch")
	}
}

func TestWorkspaceService_AddVariable(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "var-workspace",
		Name:     "Variable Workspace",
		TenantID: "tenant-var",
	}
	ws.Create(workspace)

	err := ws.AddVariable("var-workspace", "API_KEY", "secret-key", true)
	if err != nil {
		t.Fatalf("Failed to add variable: %v", err)
	}
}

func TestWorkspaceService_GetVariables(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "get-var-workspace",
		Name:     "Get Variable Workspace",
		TenantID: "tenant-getvar",
	}
	ws.Create(workspace)
	ws.AddVariable("get-var-workspace", "VAR1", "value1", false)
	ws.AddVariable("get-var-workspace", "VAR2", "value2", true)

	vars := ws.GetVariables("get-var-workspace")
	if len(vars) < 2 {
		t.Errorf("Expected at least 2 variables, got %d", len(vars))
	}
}

func TestWorkspaceService_Clone(t *testing.T) {
	ws := service.NewWorkspaceService()

	workspace := &service.Workspace{
		ID:       "clone-source",
		Name:     "Clone Source",
		TenantID: "tenant-clone",
	}
	ws.Create(workspace)

	cloned, err := ws.Clone("clone-source", "cloned-workspace", "tenant-clone")
	if err != nil {
		t.Fatalf("Failed to clone workspace: %v", err)
	}

	if cloned.ID == workspace.ID {
		t.Error("Cloned workspace should have different ID")
	}

	if cloned.Name != "cloned-workspace" {
		t.Error("Cloned workspace name mismatch")
	}
}

func TestWorkspaceService_GetStats(t *testing.T) {
	ws := service.NewWorkspaceService()

	ws.Create(&service.Workspace{
		ID:       "stats-ws-1",
		Name:     "Stats 1",
		TenantID: "tenant-wsstats",
	})

	ws.Create(&service.Workspace{
		ID:       "stats-ws-2",
		Name:     "Stats 2",
		TenantID: "tenant-wsstats",
	})

	stats := ws.GetStats("tenant-wsstats")

	if stats.TotalWorkspaces < 2 {
		t.Errorf("Expected at least 2 workspaces, got %d", stats.TotalWorkspaces)
	}
}

func TestWorkspaceService_WorkspaceToJSON(t *testing.T) {
	workspace := &service.Workspace{
		ID:          "json-workspace",
		Name:        "JSON Workspace",
		Description: "Test",
		TenantID:    "tenant-1",
		Path:        "/test",
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	data, err := workspace.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}