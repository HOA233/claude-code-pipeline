package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// BackupService manages backups
type BackupService struct {
	mu         sync.RWMutex
	backups    map[string]*Backup
	schedules  map[string]*BackupSchedule
	storageDir string
}

// Backup represents a backup
type Backup struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         BackupType    `json:"type"`
	Status       BackupStatus  `json:"status"`
	Size         int64         `json:"size"`
	Compression  string        `json:"compression"`
	Encrypted    bool          `json:"encrypted"`
	Checksum     string        `json:"checksum"`
	StoragePath  string        `json:"storage_path"`
	SourceType   string        `json:"source_type"`
	SourceID     string        `json:"source_id"`
	TenantID     string        `json:"tenant_id"`
	RetentionDays int          `json:"retention_days"`
	ExpiresAt    *time.Time    `json:"expires_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	Duration     int64         `json:"duration_ms"`
	Error        string        `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
}

// BackupSchedule represents a backup schedule
type BackupSchedule struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Cron         string        `json:"cron"`
	Type         BackupType    `json:"type"`
	SourceType   string        `json:"source_type"`
	SourceID     string        `json:"source_id"`
	TenantID     string        `json:"tenant_id"`
	RetentionDays int          `json:"retention_days"`
	Enabled      bool          `json:"enabled"`
	LastRun      *time.Time    `json:"last_run,omitempty"`
	NextRun      *time.Time    `json:"next_run,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// BackupType represents backup type
type BackupType string

const (
	BackupTypeFull    BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
	BackupTypeSnapshot BackupType = "snapshot"
)

// BackupStatus represents backup status
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
	BackupStatusExpired   BackupStatus = "expired"
	BackupStatusRestoring BackupStatus = "restoring"
)

// RestoreRequest represents a restore request
type RestoreRequest struct {
	BackupID     string    `json:"backup_id"`
	TargetID     string    `json:"target_id"`
	Overwrite    bool      `json:"overwrite"`
	RestorePoint time.Time `json:"restore_point,omitempty"`
}

// NewBackupService creates a new backup service
func NewBackupService(storageDir string) *BackupService {
	os.MkdirAll(storageDir, 0755)
	return &BackupService{
		backups:    make(map[string]*Backup),
		schedules:  make(map[string]*BackupSchedule),
		storageDir: storageDir,
	}
}

// CreateBackup creates a new backup
func (s *BackupService) CreateBackup(ctx context.Context, backup *Backup) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if backup.Name == "" {
		return fmt.Errorf("name is required")
	}
	if backup.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	now := time.Now()
	if backup.ID == "" {
		backup.ID = generateID()
	}
	backup.Status = BackupStatusPending
	backup.CreatedAt = now

	if backup.RetentionDays > 0 {
		expiresAt := now.AddDate(0, 0, backup.RetentionDays)
		backup.ExpiresAt = &expiresAt
	}

	s.backups[backup.ID] = backup

	return nil
}

// StartBackup starts a backup
func (s *BackupService) StartBackup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[id]
	if !ok {
		return fmt.Errorf("backup not found: %s", id)
	}

	backup.Status = BackupStatusRunning
	return nil
}

// CompleteBackup completes a backup
func (s *BackupService) CompleteBackup(id string, size int64, checksum string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[id]
	if !ok {
		return fmt.Errorf("backup not found: %s", id)
	}

	now := time.Now()
	backup.Status = BackupStatusCompleted
	backup.CompletedAt = &now
	backup.Duration = now.Sub(backup.CreatedAt).Milliseconds()
	backup.Size = size
	backup.Checksum = checksum

	return nil
}

// FailBackup marks a backup as failed
func (s *BackupService) FailBackup(id string, err string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[id]
	if !ok {
		return fmt.Errorf("backup not found: %s", id)
	}

	now := time.Now()
	backup.Status = BackupStatusFailed
	backup.CompletedAt = &now
	backup.Error = err
	backup.Duration = now.Sub(backup.CreatedAt).Milliseconds()

	return nil
}

// GetBackup gets a backup by ID
func (s *BackupService) GetBackup(id string) (*Backup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backup, ok := s.backups[id]
	if !ok {
		return nil, fmt.Errorf("backup not found: %s", id)
	}
	return backup, nil
}

// ListBackups lists backups
func (s *BackupService) ListBackups(tenantID string, backupType BackupType) []*Backup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Backup
	for _, backup := range s.backups {
		if tenantID != "" && backup.TenantID != tenantID {
			continue
		}
		if backupType != "" && backup.Type != backupType {
			continue
		}
		results = append(results, backup)
	}
	return results
}

// DeleteBackup deletes a backup
func (s *BackupService) DeleteBackup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[id]
	if !ok {
		return fmt.Errorf("backup not found: %s", id)
	}

	// Delete file
	if backup.StoragePath != "" {
		os.Remove(backup.StoragePath)
	}

	delete(s.backups, id)
	return nil
}

// CleanupExpiredBackups removes expired backups
func (s *BackupService) CleanupExpiredBackups() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	deleted := 0

	for id, backup := range s.backups {
		if backup.ExpiresAt != nil && now.After(*backup.ExpiresAt) {
			if backup.StoragePath != "" {
				os.Remove(backup.StoragePath)
			}
			delete(s.backups, id)
			deleted++
		}
	}

	return deleted
}

// CreateSchedule creates a backup schedule
func (s *BackupService) CreateSchedule(schedule *BackupSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if schedule.Name == "" {
		return fmt.Errorf("name is required")
	}
	if schedule.Cron == "" {
		return fmt.Errorf("cron is required")
	}

	if schedule.ID == "" {
		schedule.ID = generateID()
	}
	schedule.CreatedAt = time.Now()
	schedule.Enabled = true

	s.schedules[schedule.ID] = schedule

	return nil
}

// GetSchedule gets a backup schedule
func (s *BackupService) GetSchedule(id string) (*BackupSchedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	return schedule, nil
}

// ListSchedules lists backup schedules
func (s *BackupService) ListSchedules(tenantID string) []*BackupSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*BackupSchedule
	for _, schedule := range s.schedules {
		if tenantID == "" || schedule.TenantID == tenantID {
			results = append(results, schedule)
		}
	}
	return results
}

// EnableSchedule enables a backup schedule
func (s *BackupService) EnableSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	schedule.Enabled = true
	return nil
}

// DisableSchedule disables a backup schedule
func (s *BackupService) DisableSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	schedule.Enabled = false
	return nil
}

// DeleteSchedule deletes a backup schedule
func (s *BackupService) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.schedules[id]; !ok {
		return fmt.Errorf("schedule not found: %s", id)
	}

	delete(s.schedules, id)
	return nil
}

// Restore restores from a backup
func (s *BackupService) Restore(ctx context.Context, req *RestoreRequest) (*Backup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[req.BackupID]
	if !ok {
		return nil, fmt.Errorf("backup not found: %s", req.BackupID)
	}

	if backup.Status != BackupStatusCompleted {
		return nil, fmt.Errorf("backup is not completed")
	}

	// Create restore record
	restoreBackup := &Backup{
		ID:          generateID(),
		Name:        "restore-" + backup.Name,
		Type:        backup.Type,
		Status:      BackupStatusRestoring,
		SourceType:  backup.SourceType,
		SourceID:    req.TargetID,
		TenantID:    backup.TenantID,
		CreatedAt:   time.Now(),
		Metadata: map[string]string{
			"restored_from": req.BackupID,
		},
	}

	s.backups[restoreBackup.ID] = restoreBackup

	return restoreBackup, nil
}

// GetStats gets backup statistics
func (s *BackupService) GetStats(tenantID string) *BackupStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &BackupStats{
		ByType:   make(map[BackupType]int),
		ByStatus: make(map[BackupStatus]int),
	}

	for _, backup := range s.backups {
		if tenantID != "" && backup.TenantID != tenantID {
			continue
		}

		stats.TotalBackups++
		stats.TotalSize += backup.Size
		stats.ByType[backup.Type]++
		stats.ByStatus[backup.Status]++
	}

	stats.TotalSchedules = len(s.schedules)

	return stats
}

// BackupStats represents backup statistics
type BackupStats struct {
	TotalBackups  int                 `json:"total_backups"`
	TotalSize     int64               `json:"total_size"`
	TotalSchedules int                `json:"total_schedules"`
	ByType        map[BackupType]int  `json:"by_type"`
	ByStatus      map[BackupStatus]int `json:"by_status"`
}

// GetLatestBackup gets the latest backup for a source
func (s *BackupService) GetLatestBackup(sourceType, sourceID string) *Backup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *Backup
	for _, backup := range s.backups {
		if backup.SourceType == sourceType && backup.SourceID == sourceID {
			if backup.Status == BackupStatusCompleted {
				if latest == nil || backup.CreatedAt.After(latest.CreatedAt) {
					latest = backup
				}
			}
		}
	}
	return latest
}

// SearchBackups searches backups by name or tags
func (s *BackupService) SearchBackups(query string) []*Backup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Backup
	query = strings.ToLower(query)

	for _, backup := range s.backups {
		// Search by name
		if strings.Contains(strings.ToLower(backup.Name), query) {
			results = append(results, backup)
			continue
		}

		// Search by tags
		for _, tag := range backup.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, backup)
				break
			}
		}
	}

	return results
}

// AddBackupTags adds tags to a backup
func (s *BackupService) AddBackupTags(id string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, ok := s.backups[id]
	if !ok {
		return fmt.Errorf("backup not found: %s", id)
	}

	tagSet := make(map[string]bool)
	for _, t := range backup.Tags {
		tagSet[t] = true
	}
	for _, t := range tags {
		tagSet[t] = true
	}

	backup.Tags = make([]string, 0, len(tagSet))
	for t := range tagSet {
		backup.Tags = append(backup.Tags, t)
	}

	return nil
}

// DownloadBackup downloads a backup
func (s *BackupService) DownloadBackup(id string) (*Backup, io.ReadCloser, error) {
	s.mu.RLock()
	backup, ok := s.backups[id]
	s.mu.RUnlock()

	if !ok {
		return nil, nil, fmt.Errorf("backup not found: %s", id)
	}

	if backup.StoragePath == "" {
		return nil, nil, fmt.Errorf("backup has no storage path")
	}

	file, err := os.Open(backup.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	return backup, file, nil
}

// VerifyBackup verifies backup integrity
func (s *BackupService) VerifyBackup(id string) (bool, error) {
	s.mu.RLock()
	backup, ok := s.backups[id]
	s.mu.RUnlock()

	if !ok {
		return false, fmt.Errorf("backup not found: %s", id)
	}

	if backup.StoragePath == "" {
		return false, fmt.Errorf("backup has no storage path")
	}

	// Check file exists
	info, err := os.Stat(backup.StoragePath)
	if err != nil {
		return false, nil
	}

	// Verify size
	if backup.Size > 0 && info.Size() != backup.Size {
		return false, nil
	}

	return true, nil
}

// ToJSON serializes backup to JSON
func (b *Backup) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}

// ToJSON serializes schedule to JSON
func (s *BackupSchedule) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// GetStoragePath returns the storage path for a backup
func (s *BackupService) GetStoragePath(backup *Backup) string {
	return filepath.Join(s.storageDir, backup.TenantID, backup.ID+".bak")
}