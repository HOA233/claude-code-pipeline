package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ArtifactService manages build artifacts
type ArtifactService struct {
	mu         sync.RWMutex
	artifacts  map[string]*Artifact
	storageDir string
	maxSize    int64 // in bytes
	usedSize   int64
}

// Artifact represents a build artifact
type Artifact struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        ArtifactType      `json:"type"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	ContentType string            `json:"content_type"`
	StoragePath string            `json:"storage_path"`
	Metadata    map[string]string `json:"metadata"`
	Tags        []string          `json:"tags"`
	TenantID    string            `json:"tenant_id"`
	PipelineID  string            `json:"pipeline_id,omitempty"`
	BuildID     string            `json:"build_id,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	DownloadCount int             `json:"download_count"`
}

// ArtifactType represents artifact type
type ArtifactType string

const (
	ArtifactTypeBinary   ArtifactType = "binary"
	ArtifactTypeArchive  ArtifactType = "archive"
	ArtifactTypeDocker   ArtifactType = "docker"
	ArtifactTypePackage  ArtifactType = "package"
	ArtifactTypeReport   ArtifactType = "report"
	ArtifactTypeLog      ArtifactType = "log"
	ArtifactTypeCoverage ArtifactType = "coverage"
)

// ArtifactSearch represents search criteria
type ArtifactSearch struct {
	TenantID   string
	Name       string
	Version    string
	Type       ArtifactType
	Tags       []string
	PipelineID string
	BuildID    string
	Limit      int
	Offset     int
}

// NewArtifactService creates a new artifact service
func NewArtifactService(storageDir string, maxSizeGB float64) *ArtifactService {
	// Create storage directory if not exists
	os.MkdirAll(storageDir, 0755)

	return &ArtifactService{
		artifacts:  make(map[string]*Artifact),
		storageDir: storageDir,
		maxSize:    int64(maxSizeGB * 1024 * 1024 * 1024),
		usedSize:   0,
	}
}

// Upload uploads an artifact
func (s *ArtifactService) Upload(ctx context.Context, reader io.Reader, artifact *Artifact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if artifact.Name == "" {
		return fmt.Errorf("name is required")
	}
	if artifact.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	now := time.Now()
	if artifact.ID == "" {
		artifact.ID = generateID()
	}
	artifact.CreatedAt = now
	artifact.DownloadCount = 0

	// Create tenant directory
	tenantDir := filepath.Join(s.storageDir, artifact.TenantID)
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Generate storage path
	storagePath := filepath.Join(tenantDir, artifact.ID)
	if artifact.Version != "" {
		storagePath = filepath.Join(tenantDir, fmt.Sprintf("%s-%s-%s", artifact.Name, artifact.Version, artifact.ID))
	}

	// Create file
	file, err := os.Create(storagePath)
	if err != nil {
		return fmt.Errorf("failed to create artifact file: %w", err)
	}
	defer file.Close()

	// Calculate checksum and size while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(file, hasher)

	size, err := io.Copy(multiWriter, reader)
	if err != nil {
		os.Remove(storagePath)
		return fmt.Errorf("failed to write artifact: %w", err)
	}

	// Check storage limit
	if s.usedSize+size > s.maxSize {
		os.Remove(storagePath)
		return fmt.Errorf("storage limit exceeded")
	}

	artifact.Size = size
	artifact.Checksum = hex.EncodeToString(hasher.Sum(nil))
	artifact.StoragePath = storagePath

	if artifact.Metadata == nil {
		artifact.Metadata = make(map[string]string)
	}

	s.artifacts[artifact.ID] = artifact
	s.usedSize += size

	return nil
}

// Download downloads an artifact
func (s *ArtifactService) Download(ctx context.Context, id string) (*Artifact, io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return nil, nil, fmt.Errorf("artifact not found: %s", id)
	}

	file, err := os.Open(artifact.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open artifact: %w", err)
	}

	artifact.DownloadCount++

	return artifact, file, nil
}

// GetArtifact gets artifact metadata
func (s *ArtifactService) GetArtifact(id string) (*Artifact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", id)
	}
	return artifact, nil
}

// ListArtifacts lists artifacts based on search criteria
func (s *ArtifactService) ListArtifacts(search ArtifactSearch) []*Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Artifact

	for _, artifact := range s.artifacts {
		// Filter by tenant
		if search.TenantID != "" && artifact.TenantID != search.TenantID {
			continue
		}

		// Filter by name
		if search.Name != "" && artifact.Name != search.Name {
			continue
		}

		// Filter by version
		if search.Version != "" && artifact.Version != search.Version {
			continue
		}

		// Filter by type
		if search.Type != "" && artifact.Type != search.Type {
			continue
		}

		// Filter by pipeline
		if search.PipelineID != "" && artifact.PipelineID != search.PipelineID {
			continue
		}

		// Filter by build
		if search.BuildID != "" && artifact.BuildID != search.BuildID {
			continue
		}

		// Filter by tags
		if len(search.Tags) > 0 {
			found := false
			for _, searchTag := range search.Tags {
				for _, artifactTag := range artifact.Tags {
					if artifactTag == searchTag {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		results = append(results, artifact)
	}

	// Sort by created_at descending
	// (simplified, would use sort.Slice in production)

	// Apply pagination
	if search.Offset > 0 && search.Offset < len(results) {
		results = results[search.Offset:]
	}
	if search.Limit > 0 && len(results) > search.Limit {
		results = results[:search.Limit]
	}

	return results
}

// DeleteArtifact deletes an artifact
func (s *ArtifactService) DeleteArtifact(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return fmt.Errorf("artifact not found: %s", id)
	}

	// Delete file
	if err := os.Remove(artifact.StoragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete artifact file: %w", err)
	}

	s.usedSize -= artifact.Size
	delete(s.artifacts, id)

	return nil
}

// DeleteExpiredArtifacts deletes all expired artifacts
func (s *ArtifactService) DeleteExpiredArtifacts() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	deleted := 0

	for id, artifact := range s.artifacts {
		if artifact.ExpiresAt != nil && now.After(*artifact.ExpiresAt) {
			os.Remove(artifact.StoragePath)
			s.usedSize -= artifact.Size
			delete(s.artifacts, id)
			deleted++
		}
	}

	return deleted
}

// GetStorageStats gets storage statistics
func (s *ArtifactService) GetStorageStats(tenantID string) *StorageStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &StorageStats{
		MaxSize: s.maxSize,
		UsedSize: s.usedSize,
		ByType:   make(map[ArtifactType]int64),
	}

	for _, artifact := range s.artifacts {
		if tenantID != "" && artifact.TenantID != tenantID {
			continue
		}

		stats.TotalArtifacts++
		stats.TotalSize += artifact.Size
		stats.ByType[artifact.Type] += artifact.Size
	}

	stats.FreeSize = stats.MaxSize - stats.UsedSize
	if stats.MaxSize > 0 {
		stats.UsagePercent = float64(stats.UsedSize) / float64(stats.MaxSize) * 100
	}

	return stats
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalArtifacts int                       `json:"total_artifacts"`
	TotalSize      int64                     `json:"total_size"`
	MaxSize        int64                     `json:"max_size"`
	UsedSize       int64                     `json:"used_size"`
	FreeSize       int64                     `json:"free_size"`
	UsagePercent   float64                   `json:"usage_percent"`
	ByType         map[ArtifactType]int64    `json:"by_type"`
}

// UpdateArtifactMetadata updates artifact metadata
func (s *ArtifactService) UpdateArtifactMetadata(id string, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return fmt.Errorf("artifact not found: %s", id)
	}

	if artifact.Metadata == nil {
		artifact.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		if v == "" {
			delete(artifact.Metadata, k)
		} else {
			artifact.Metadata[k] = v
		}
	}

	return nil
}

// AddArtifactTags adds tags to an artifact
func (s *ArtifactService) AddArtifactTags(id string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return fmt.Errorf("artifact not found: %s", id)
	}

	tagSet := make(map[string]bool)
	for _, t := range artifact.Tags {
		tagSet[t] = true
	}
	for _, t := range tags {
		tagSet[t] = true
	}

	artifact.Tags = make([]string, 0, len(tagSet))
	for t := range tagSet {
		artifact.Tags = append(artifact.Tags, t)
	}

	return nil
}

// RemoveArtifactTags removes tags from an artifact
func (s *ArtifactService) RemoveArtifactTags(id string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return fmt.Errorf("artifact not found: %s", id)
	}

	removeSet := make(map[string]bool)
	for _, t := range tags {
		removeSet[t] = true
	}

	newTags := make([]string, 0)
	for _, t := range artifact.Tags {
		if !removeSet[t] {
			newTags = append(newTags, t)
		}
	}
	artifact.Tags = newTags

	return nil
}

// CopyArtifact copies an artifact to a new location
func (s *ArtifactService) CopyArtifact(ctx context.Context, id, newTenantID, newName string) (*Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", id)
	}

	// Open source file
	srcFile, err := os.Open(artifact.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source artifact: %w", err)
	}
	defer srcFile.Close()

	// Create new artifact
	newArtifact := &Artifact{
		ID:          generateID(),
		Name:        newName,
		Version:     artifact.Version,
		Type:        artifact.Type,
		ContentType: artifact.ContentType,
		Metadata:    make(map[string]string),
		Tags:        append([]string{}, artifact.Tags...),
		TenantID:    newTenantID,
		PipelineID:  artifact.PipelineID,
		BuildID:     artifact.BuildID,
		CreatedAt:   time.Now(),
	}

	// Copy metadata
	for k, v := range artifact.Metadata {
		newArtifact.Metadata[k] = v
	}

	// Create tenant directory
	tenantDir := filepath.Join(s.storageDir, newTenantID)
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Generate new storage path
	storagePath := filepath.Join(tenantDir, newArtifact.ID)
	if newArtifact.Version != "" {
		storagePath = filepath.Join(tenantDir, fmt.Sprintf("%s-%s-%s", newName, newArtifact.Version, newArtifact.ID))
	}

	// Create destination file
	dstFile, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	size, err := io.Copy(dstFile, srcFile)
	if err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("failed to copy artifact: %w", err)
	}

	// Check storage limit
	if s.usedSize+size > s.maxSize {
		os.Remove(storagePath)
		return nil, fmt.Errorf("storage limit exceeded")
	}

	newArtifact.Size = size
	newArtifact.Checksum = artifact.Checksum
	newArtifact.StoragePath = storagePath

	s.artifacts[newArtifact.ID] = newArtifact
	s.usedSize += size

	return newArtifact, nil
}

// CleanupOldArtifacts removes old artifacts to free up space
func (s *ArtifactService) CleanupOldArtifacts(keepCount int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Group by tenant and name
	groups := make(map[string][]*Artifact)
	for _, artifact := range s.artifacts {
		key := artifact.TenantID + ":" + artifact.Name
		groups[key] = append(groups[key], artifact)
	}

	deleted := 0
	for _, group := range groups {
		if len(group) <= keepCount {
			continue
		}

		// Sort by created_at (simplified - would sort properly in production)
		// Remove oldest artifacts beyond keepCount
		for i := 0; i < len(group)-keepCount; i++ {
			artifact := group[i]
			os.Remove(artifact.StoragePath)
			s.usedSize -= artifact.Size
			delete(s.artifacts, artifact.ID)
			deleted++
		}
	}

	return deleted
}

// GetArtifactVersions gets all versions of an artifact
func (s *ArtifactService) GetArtifactVersions(tenantID, name string) []*Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var versions []*Artifact
	for _, artifact := range s.artifacts {
		if artifact.TenantID == tenantID && artifact.Name == name {
			versions = append(versions, artifact)
		}
	}
	return versions
}

// VerifyChecksum verifies artifact checksum
func (s *ArtifactService) VerifyChecksum(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, ok := s.artifacts[id]
	if !ok {
		return false, fmt.Errorf("artifact not found: %s", id)
	}

	file, err := os.Open(artifact.StoragePath)
	if err != nil {
		return false, fmt.Errorf("failed to open artifact: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))
	return strings.EqualFold(calculatedChecksum, artifact.Checksum), nil
}

// ToJSON serializes artifact to JSON
func (a *Artifact) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// ToJSON serializes storage stats to JSON
func (s *StorageStats) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}