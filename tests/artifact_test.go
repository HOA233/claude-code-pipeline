package tests

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Artifact Service Tests

func TestArtifactService_New(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "artifacts")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)
	if as == nil {
		t.Fatal("Expected non-nil artifact service")
	}
}

func TestArtifactService_Upload(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	content := []byte("test artifact content")
	reader := bytes.NewReader(content)

	artifact := &service.Artifact{
		Name:        "test.txt",
		TenantID:    "tenant-1",
		Type:        service.ArtifactTypeLog,
		ContentType: "text/plain",
	}

	err := as.Upload(context.Background(), reader, artifact)
	if err != nil {
		t.Fatalf("Failed to upload artifact: %v", err)
	}

	if artifact.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if artifact.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), artifact.Size)
	}

	if artifact.Checksum == "" {
		t.Error("Expected checksum to be calculated")
	}
}

func TestArtifactService_Upload_MissingName(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		TenantID: "tenant-1",
		Type:     service.ArtifactTypeLog,
	}

	err := as.Upload(context.Background(), bytes.NewReader([]byte("test")), artifact)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestArtifactService_Upload_MissingTenant(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name: "test.txt",
		Type: service.ArtifactTypeLog,
	}

	err := as.Upload(context.Background(), bytes.NewReader([]byte("test")), artifact)
	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestArtifactService_GetArtifact(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "get-test.txt",
		TenantID: "tenant-get",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	retrieved, err := as.GetArtifact(artifact.ID)
	if err != nil {
		t.Fatalf("Failed to get artifact: %v", err)
	}

	if retrieved.Name != "get-test.txt" {
		t.Error("Artifact name mismatch")
	}
}

func TestArtifactService_GetArtifact_NotFound(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	_, err := as.GetArtifact("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent artifact")
	}
}

func TestArtifactService_Download(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	content := []byte("download test content")
	artifact := &service.Artifact{
		Name:     "download.txt",
		TenantID: "tenant-download",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader(content), artifact)

	meta, reader, err := as.Download(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("Failed to download artifact: %v", err)
	}
	defer reader.Close()

	if meta.Name != "download.txt" {
		t.Error("Artifact name mismatch")
	}

	// Read content
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	if !bytes.Equal(buf.Bytes(), content) {
		t.Error("Content mismatch")
	}

	// Check download count
	if meta.DownloadCount != 1 {
		t.Errorf("Expected download count 1, got %d", meta.DownloadCount)
	}
}

func TestArtifactService_ListArtifacts(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	for i := 0; i < 3; i++ {
		as.Upload(context.Background(), bytes.NewReader([]byte("content")), &service.Artifact{
			Name:     "list-test-" + string(rune('a'+i)),
			TenantID: "tenant-list",
			Type:     service.ArtifactTypeLog,
		})
	}

	// Add artifact for different tenant
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), &service.Artifact{
		Name:     "other.txt",
		TenantID: "other-tenant",
		Type:     service.ArtifactTypeLog,
	})

	artifacts := as.ListArtifacts(service.ArtifactSearch{
		TenantID: "tenant-list",
	})

	if len(artifacts) < 3 {
		t.Errorf("Expected at least 3 artifacts, got %d", len(artifacts))
	}
}

func TestArtifactService_ListArtifacts_FilterByType(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	as.Upload(context.Background(), bytes.NewReader([]byte("log")), &service.Artifact{
		Name:     "log.txt",
		TenantID: "tenant-filter",
		Type:     service.ArtifactTypeLog,
	})

	as.Upload(context.Background(), bytes.NewReader([]byte("report")), &service.Artifact{
		Name:     "report.html",
		TenantID: "tenant-filter",
		Type:     service.ArtifactTypeReport,
	})

	artifacts := as.ListArtifacts(service.ArtifactSearch{
		TenantID: "tenant-filter",
		Type:     service.ArtifactTypeLog,
	})

	for _, a := range artifacts {
		if a.Type != service.ArtifactTypeLog {
			t.Errorf("Expected only log artifacts, got %s", a.Type)
		}
	}
}

func TestArtifactService_DeleteArtifact(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "delete.txt",
		TenantID: "tenant-delete",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	err := as.DeleteArtifact(artifact.ID)
	if err != nil {
		t.Fatalf("Failed to delete artifact: %v", err)
	}

	_, err = as.GetArtifact(artifact.ID)
	if err == nil {
		t.Error("Expected error for deleted artifact")
	}
}

func TestArtifactService_GetStorageStats(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	for i := 0; i < 3; i++ {
		as.Upload(context.Background(), bytes.NewReader([]byte("content")), &service.Artifact{
			Name:     "stats-" + string(rune('a'+i)),
			TenantID: "tenant-stats",
			Type:     service.ArtifactTypeLog,
		})
	}

	stats := as.GetStorageStats("tenant-stats")

	if stats.TotalArtifacts < 3 {
		t.Errorf("Expected at least 3 artifacts, got %d", stats.TotalArtifacts)
	}

	if stats.TotalSize <= 0 {
		t.Error("Expected positive total size")
	}
}

func TestArtifactService_UpdateArtifactMetadata(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "metadata.txt",
		TenantID: "tenant-meta",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	err := as.UpdateArtifactMetadata(artifact.ID, map[string]string{
		"author":  "test-user",
		"version": "1.0",
	})

	if err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	retrieved, _ := as.GetArtifact(artifact.ID)
	if retrieved.Metadata["author"] != "test-user" {
		t.Error("Metadata not updated")
	}
}

func TestArtifactService_AddArtifactTags(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "tags.txt",
		TenantID: "tenant-tags",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	err := as.AddArtifactTags(artifact.ID, []string{"prod", "release"})
	if err != nil {
		t.Fatalf("Failed to add tags: %v", err)
	}

	retrieved, _ := as.GetArtifact(artifact.ID)

	tagSet := make(map[string]bool)
	for _, t := range retrieved.Tags {
		tagSet[t] = true
	}

	if !tagSet["prod"] || !tagSet["release"] {
		t.Error("Tags not added correctly")
	}
}

func TestArtifactService_RemoveArtifactTags(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "remove-tags.txt",
		TenantID: "tenant-removetags",
		Type:     service.ArtifactTypeLog,
		Tags:     []string{"keep", "remove"},
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	err := as.RemoveArtifactTags(artifact.ID, []string{"remove"})
	if err != nil {
		t.Fatalf("Failed to remove tags: %v", err)
	}

	retrieved, _ := as.GetArtifact(artifact.ID)

	for _, t := range retrieved.Tags {
		if t == "remove" {
			t.Error("Tag should have been removed")
		}
	}
}

func TestArtifactService_DeleteExpiredArtifacts(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	// Create expired artifact
	past := time.Now().Add(-24 * time.Hour)
	artifact := &service.Artifact{
		Name:      "expired.txt",
		TenantID:  "tenant-expired",
		Type:      service.ArtifactTypeLog,
		ExpiresAt: &past,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("content")), artifact)

	deleted := as.DeleteExpiredArtifacts()
	if deleted < 1 {
		t.Errorf("Expected at least 1 deleted artifact, got %d", deleted)
	}

	_, err := as.GetArtifact(artifact.ID)
	if err == nil {
		t.Error("Expected error for expired artifact")
	}
}

func TestArtifactService_GetArtifactVersions(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	as.Upload(context.Background(), bytes.NewReader([]byte("v1")), &service.Artifact{
		Name:    "app.jar",
		Version: "1.0.0",
		TenantID: "tenant-versions",
		Type:    service.ArtifactTypeBinary,
	})

	as.Upload(context.Background(), bytes.NewReader([]byte("v2")), &service.Artifact{
		Name:    "app.jar",
		Version: "2.0.0",
		TenantID: "tenant-versions",
		Type:    service.ArtifactTypeBinary,
	})

	versions := as.GetArtifactVersions("tenant-versions", "app.jar")
	if len(versions) < 2 {
		t.Errorf("Expected at least 2 versions, got %d", len(versions))
	}
}

func TestArtifactService_VerifyChecksum(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	artifact := &service.Artifact{
		Name:     "verify.txt",
		TenantID: "tenant-verify",
		Type:     service.ArtifactTypeLog,
	}
	as.Upload(context.Background(), bytes.NewReader([]byte("original content")), artifact)

	valid, err := as.VerifyChecksum(artifact.ID)
	if err != nil {
		t.Fatalf("Failed to verify checksum: %v", err)
	}

	if !valid {
		t.Error("Expected checksum to be valid")
	}
}

func TestArtifactService_ArtifactTypes(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	as := service.NewArtifactService(dir, 1.0)

	types := []service.ArtifactType{
		service.ArtifactTypeBinary,
		service.ArtifactTypeArchive,
		service.ArtifactTypeDocker,
		service.ArtifactTypePackage,
		service.ArtifactTypeReport,
		service.ArtifactTypeLog,
		service.ArtifactTypeCoverage,
	}

	for i, artifactType := range types {
		err := as.Upload(context.Background(), bytes.NewReader([]byte("content")), &service.Artifact{
			Name:     string(artifactType) + ".file",
			TenantID: "tenant-types",
			Type:     artifactType,
		})

		if err != nil {
			t.Errorf("Failed to upload %s artifact: %v", artifactType, err)
		}

		_ = i
	}
}

func TestArtifactService_ArtifactToJSON(t *testing.T) {
	artifact := &service.Artifact{
		ID:        "art-1",
		Name:      "test.txt",
		TenantID:  "tenant-1",
		Type:      service.ArtifactTypeLog,
		Size:      100,
		Checksum:  "abc123",
		CreatedAt: time.Now(),
	}

	data, err := artifact.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestArtifactService_StorageLimit(t *testing.T) {
	dir, _ := os.MkdirTemp("", "artifacts")
	defer os.RemoveAll(dir)

	// Create service with very small limit
	as := service.NewArtifactService(dir, 0.000001) // ~1KB

	// First upload should succeed
	err := as.Upload(context.Background(), bytes.NewReader(make([]byte, 500)), &service.Artifact{
		Name:     "small.txt",
		TenantID: "tenant-limit",
		Type:     service.ArtifactTypeLog,
	})

	if err != nil {
		t.Fatalf("First upload should succeed: %v", err)
	}

	// Second large upload should fail
	err = as.Upload(context.Background(), bytes.NewReader(make([]byte, 50000)), &service.Artifact{
		Name:     "large.txt",
		TenantID: "tenant-limit",
		Type:     service.ArtifactTypeLog,
	})

	if err == nil {
		t.Error("Expected error for exceeding storage limit")
	}
}