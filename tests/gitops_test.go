package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// GitOps Service Tests

func TestGitOpsService_New(t *testing.T) {
	gs := service.NewGitOpsService()
	if gs == nil {
		t.Fatal("Expected non-nil gitops service")
	}
}

func TestGitOpsService_CreateRepository(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "repo-1",
		Name:     "test-repo",
		URL:      "https://github.com/test/repo.git",
		Branch:   "main",
		TenantID: "tenant-1",
	}

	err := gs.CreateRepository(repo)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
}

func TestGitOpsService_CreateRepository_MissingName(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "no-name",
		URL:      "https://github.com/test/repo.git",
		TenantID: "tenant-1",
	}

	err := gs.CreateRepository(repo)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestGitOpsService_GetRepository(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "get-repo",
		Name:     "Get Repo",
		URL:      "https://github.com/test/get.git",
		TenantID: "tenant-get",
	}
	gs.CreateRepository(repo)

	retrieved, err := gs.GetRepository("get-repo")
	if err != nil {
		t.Fatalf("Failed to get repository: %v", err)
	}

	if retrieved.Name != "Get Repo" {
		t.Error("Repository name mismatch")
	}
}

func TestGitOpsService_GetRepository_NotFound(t *testing.T) {
	gs := service.NewGitOpsService()

	_, err := gs.GetRepository("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent repository")
	}
}

func TestGitOpsService_ListRepositories(t *testing.T) {
	gs := service.NewGitOpsService()

	gs.CreateRepository(&service.GitRepository{
		ID:       "list-1",
		Name:     "List 1",
		URL:      "https://github.com/test/list1.git",
		TenantID: "tenant-list",
	})

	gs.CreateRepository(&service.GitRepository{
		ID:       "list-2",
		Name:     "List 2",
		URL:      "https://github.com/test/list2.git",
		TenantID: "tenant-list",
	})

	repos := gs.ListRepositories("tenant-list")
	if len(repos) < 2 {
		t.Errorf("Expected at least 2 repositories, got %d", len(repos))
	}
}

func TestGitOpsService_DeleteRepository(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "delete-repo",
		Name:     "Delete Repo",
		URL:      "https://github.com/test/delete.git",
		TenantID: "tenant-delete",
	}
	gs.CreateRepository(repo)

	err := gs.DeleteRepository("delete-repo")
	if err != nil {
		t.Fatalf("Failed to delete repository: %v", err)
	}

	_, err = gs.GetRepository("delete-repo")
	if err == nil {
		t.Error("Expected error for deleted repository")
	}
}

func TestGitOpsService_CreateSync(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "sync-repo",
		Name:     "Sync Repo",
		URL:      "https://github.com/test/sync.git",
		TenantID: "tenant-sync",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "sync-repo",
		Path:         "/manifests",
		Target:       "production",
	}

	err := gs.CreateSync(sync)
	if err != nil {
		t.Fatalf("Failed to create sync: %v", err)
	}

	if sync.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestGitOpsService_GetSync(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "getsync-repo",
		Name:     "Get Sync Repo",
		URL:      "https://github.com/test/getsync.git",
		TenantID: "tenant-getsync",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "getsync-repo",
		Path:         "/manifests",
	}
	gs.CreateSync(sync)

	retrieved, err := gs.GetSync(sync.ID)
	if err != nil {
		t.Fatalf("Failed to get sync: %v", err)
	}

	if retrieved.RepositoryID != "getsync-repo" {
		t.Error("Repository ID mismatch")
	}
}

func TestGitOpsService_TriggerSync(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "trigger-repo",
		Name:     "Trigger Repo",
		URL:      "https://github.com/test/trigger.git",
		TenantID: "tenant-trigger",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "trigger-repo",
		Path:         "/manifests",
		Enabled:      true,
	}
	gs.CreateSync(sync)

	err := gs.TriggerSync(sync.ID)
	if err != nil {
		t.Fatalf("Failed to trigger sync: %v", err)
	}
}

func TestGitOpsService_GetSyncStatus(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "status-repo",
		Name:     "Status Repo",
		URL:      "https://github.com/test/status.git",
		TenantID: "tenant-status",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "status-repo",
		Path:         "/manifests",
	}
	gs.CreateSync(sync)

	status := gs.GetSyncStatus(sync.ID)
	if status == nil {
		t.Error("Expected sync status")
	}
}

func TestGitOpsService_ListSyncs(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "listsync-repo",
		Name:     "List Sync Repo",
		URL:      "https://github.com/test/listsync.git",
		TenantID: "tenant-listsync",
	}
	gs.CreateRepository(repo)

	gs.CreateSync(&service.GitSync{RepositoryID: "listsync-repo", Path: "/a"})
	gs.CreateSync(&service.GitSync{RepositoryID: "listsync-repo", Path: "/b"})

	syncs := gs.ListSyncs("listsync-repo")
	if len(syncs) < 2 {
		t.Errorf("Expected at least 2 syncs, got %d", len(syncs))
	}
}

func TestGitOpsService_DeleteSync(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "delsync-repo",
		Name:     "Del Sync Repo",
		URL:      "https://github.com/test/delsync.git",
		TenantID: "tenant-delsync",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "delsync-repo",
		Path:         "/manifests",
	}
	gs.CreateSync(sync)

	err := gs.DeleteSync(sync.ID)
	if err != nil {
		t.Fatalf("Failed to delete sync: %v", err)
	}

	_, err = gs.GetSync(sync.ID)
	if err == nil {
		t.Error("Expected error for deleted sync")
	}
}

func TestGitOpsService_GetHistory(t *testing.T) {
	gs := service.NewGitOpsService()

	repo := &service.GitRepository{
		ID:       "history-repo",
		Name:     "History Repo",
		URL:      "https://github.com/test/history.git",
		TenantID: "tenant-history",
	}
	gs.CreateRepository(repo)

	sync := &service.GitSync{
		RepositoryID: "history-repo",
		Path:         "/manifests",
	}
	gs.CreateSync(sync)

	// Record some history
	gs.RecordHistory(sync.ID, "success", "Synced successfully")

	history := gs.GetHistory(sync.ID)
	if len(history) == 0 {
		t.Error("Expected sync history")
	}
}

func TestGitOpsService_GitRepositoryToJSON(t *testing.T) {
	repo := &service.GitRepository{
		ID:        "json-repo",
		Name:      "JSON Repo",
		URL:       "https://github.com/test/json.git",
		Branch:    "main",
		TenantID:  "tenant-1",
		CreatedAt: time.Now(),
	}

	data, err := repo.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}