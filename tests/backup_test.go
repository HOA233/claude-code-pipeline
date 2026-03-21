package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Backup Service Tests

func TestBackupService_New(t *testing.T) {
	dir, err := os.MkdirTemp("", "backups")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)
	if bs == nil {
		t.Fatal("Expected non-nil backup service")
	}
}

func TestBackupService_CreateBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "test-backup",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-1",
	}

	err := bs.CreateBackup(context.Background(), backup)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if backup.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if backup.Status != service.BackupStatusPending {
		t.Errorf("Expected status pending, got %s", backup.Status)
	}
}

func TestBackupService_CreateBackup_MissingName(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Type:     service.BackupTypeFull,
		TenantID: "tenant-1",
	}

	err := bs.CreateBackup(context.Background(), backup)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestBackupService_CreateBackup_MissingTenant(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name: "backup",
		Type: service.BackupTypeFull,
	}

	err := bs.CreateBackup(context.Background(), backup)
	if err == nil {
		t.Error("Expected error for missing tenant_id")
	}
}

func TestBackupService_GetBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "get-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-get",
	}
	bs.CreateBackup(context.Background(), backup)

	retrieved, err := bs.GetBackup(backup.ID)
	if err != nil {
		t.Fatalf("Failed to get backup: %v", err)
	}

	if retrieved.Name != "get-test" {
		t.Error("Backup name mismatch")
	}
}

func TestBackupService_GetBackup_NotFound(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	_, err := bs.GetBackup("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent backup")
	}
}

func TestBackupService_StartBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "start-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-start",
	}
	bs.CreateBackup(context.Background(), backup)

	err := bs.StartBackup(backup.ID)
	if err != nil {
		t.Fatalf("Failed to start backup: %v", err)
	}

	retrieved, _ := bs.GetBackup(backup.ID)
	if retrieved.Status != service.BackupStatusRunning {
		t.Error("Expected status to be running")
	}
}

func TestBackupService_CompleteBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "complete-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-complete",
	}
	bs.CreateBackup(context.Background(), backup)
	bs.StartBackup(backup.ID)

	err := bs.CompleteBackup(backup.ID, 1024, "abc123")
	if err != nil {
		t.Fatalf("Failed to complete backup: %v", err)
	}

	retrieved, _ := bs.GetBackup(backup.ID)
	if retrieved.Status != service.BackupStatusCompleted {
		t.Error("Expected status to be completed")
	}

	if retrieved.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", retrieved.Size)
	}

	if retrieved.Checksum != "abc123" {
		t.Errorf("Expected checksum abc123, got %s", retrieved.Checksum)
	}
}

func TestBackupService_FailBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "fail-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-fail",
	}
	bs.CreateBackup(context.Background(), backup)
	bs.StartBackup(backup.ID)

	err := bs.FailBackup(backup.ID, "disk full")
	if err != nil {
		t.Fatalf("Failed to fail backup: %v", err)
	}

	retrieved, _ := bs.GetBackup(backup.ID)
	if retrieved.Status != service.BackupStatusFailed {
		t.Error("Expected status to be failed")
	}

	if retrieved.Error != "disk full" {
		t.Errorf("Expected error 'disk full', got %s", retrieved.Error)
	}
}

func TestBackupService_ListBackups(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "list-1",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-list",
	})

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "list-2",
		Type:     service.BackupTypeIncremental,
		TenantID: "tenant-list",
	})

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "other",
		Type:     service.BackupTypeFull,
		TenantID: "other-tenant",
	})

	backups := bs.ListBackups("tenant-list", "")
	if len(backups) < 2 {
		t.Errorf("Expected at least 2 backups, got %d", len(backups))
	}
}

func TestBackupService_ListBackups_FilterByType(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "full",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-filter",
	})

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "incremental",
		Type:     service.BackupTypeIncremental,
		TenantID: "tenant-filter",
	})

	backups := bs.ListBackups("tenant-filter", service.BackupTypeFull)
	for _, b := range backups {
		if b.Type != service.BackupTypeFull {
			t.Errorf("Expected only full backups, got %s", b.Type)
		}
	}
}

func TestBackupService_DeleteBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "delete-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-delete",
	}
	bs.CreateBackup(context.Background(), backup)

	err := bs.DeleteBackup(backup.ID)
	if err != nil {
		t.Fatalf("Failed to delete backup: %v", err)
	}

	_, err = bs.GetBackup(backup.ID)
	if err == nil {
		t.Error("Expected error for deleted backup")
	}
}

func TestBackupService_RetentionDays(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:          "retention-test",
		Type:          service.BackupTypeFull,
		TenantID:      "tenant-retention",
		RetentionDays: 7,
	}
	bs.CreateBackup(context.Background(), backup)

	if backup.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}
}

func TestBackupService_CleanupExpiredBackups(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	// Create expired backup
	backup := &service.Backup{
		Name:          "expired",
		Type:          service.BackupTypeFull,
		TenantID:      "tenant-expired",
		RetentionDays: 0,
	}
	bs.CreateBackup(context.Background(), backup)

	// Manually set expired time
	past := time.Now().Add(-24 * time.Hour)
	backup.ExpiresAt = &past

	deleted := bs.CleanupExpiredBackups()
	if deleted < 1 {
		t.Errorf("Expected at least 1 deleted backup, got %d", deleted)
	}
}

func TestBackupService_CreateSchedule(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	schedule := &service.BackupSchedule{
		Name:       "Daily Backup",
		Cron:       "0 2 * * *",
		Type:       service.BackupTypeFull,
		TenantID:   "tenant-schedule",
	}

	err := bs.CreateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	if schedule.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestBackupService_GetSchedule(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	schedule := &service.BackupSchedule{
		Name:       "Get Schedule",
		Cron:       "0 2 * * *",
		Type:       service.BackupTypeFull,
		TenantID:   "tenant-getschedule",
	}
	bs.CreateSchedule(schedule)

	retrieved, err := bs.GetSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to get schedule: %v", err)
	}

	if retrieved.Name != "Get Schedule" {
		t.Error("Schedule name mismatch")
	}
}

func TestBackupService_ListSchedules(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	bs.CreateSchedule(&service.BackupSchedule{
		Name:     "Schedule 1",
		Cron:     "0 2 * * *",
		TenantID: "tenant-listschedule",
	})

	bs.CreateSchedule(&service.BackupSchedule{
		Name:     "Schedule 2",
		Cron:     "0 3 * * *",
		TenantID: "tenant-listschedule",
	})

	schedules := bs.ListSchedules("tenant-listschedule")
	if len(schedules) < 2 {
		t.Errorf("Expected at least 2 schedules, got %d", len(schedules))
	}
}

func TestBackupService_EnableDisableSchedule(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	schedule := &service.BackupSchedule{
		Name:     "Toggle Schedule",
		Cron:     "0 2 * * *",
		TenantID: "tenant-toggle",
	}
	bs.CreateSchedule(schedule)

	err := bs.DisableSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to disable schedule: %v", err)
	}

	retrieved, _ := bs.GetSchedule(schedule.ID)
	if retrieved.Enabled {
		t.Error("Expected schedule to be disabled")
	}

	err = bs.EnableSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to enable schedule: %v", err)
	}

	retrieved, _ = bs.GetSchedule(schedule.ID)
	if !retrieved.Enabled {
		t.Error("Expected schedule to be enabled")
	}
}

func TestBackupService_DeleteSchedule(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	schedule := &service.BackupSchedule{
		Name:     "Delete Schedule",
		Cron:     "0 2 * * *",
		TenantID: "tenant-delschedule",
	}
	bs.CreateSchedule(schedule)

	err := bs.DeleteSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to delete schedule: %v", err)
	}

	_, err = bs.GetSchedule(schedule.ID)
	if err == nil {
		t.Error("Expected error for deleted schedule")
	}
}

func TestBackupService_Restore(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:        "restore-source",
		Type:        service.BackupTypeFull,
		TenantID:    "tenant-restore",
		SourceType:  "database",
		SourceID:    "db-1",
	}
	bs.CreateBackup(context.Background(), backup)
	bs.StartBackup(backup.ID)
	bs.CompleteBackup(backup.ID, 1024, "abc")

	restoreBackup, err := bs.Restore(context.Background(), &service.RestoreRequest{
		BackupID:  backup.ID,
		TargetID:  "db-2",
		Overwrite: false,
	})

	if err != nil {
		t.Fatalf("Failed to restore: %v", err)
	}

	if restoreBackup.Status != service.BackupStatusRestoring {
		t.Error("Expected status to be restoring")
	}
}

func TestBackupService_GetStats(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "stats-1",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-stats",
	})

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "stats-2",
		Type:     service.BackupTypeIncremental,
		TenantID: "tenant-stats",
	})

	stats := bs.GetStats("tenant-stats")

	if stats.TotalBackups < 2 {
		t.Errorf("Expected at least 2 backups, got %d", stats.TotalBackups)
	}
}

func TestBackupService_GetLatestBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	// Create older backup
	bs.CreateBackup(context.Background(), &service.Backup{
		Name:       "older",
		Type:       service.BackupTypeFull,
		TenantID:   "tenant-latest",
		SourceType: "database",
		SourceID:   "db-1",
	})

	// Create newer backup
	backup := &service.Backup{
		Name:       "newer",
		Type:       service.BackupTypeFull,
		TenantID:   "tenant-latest",
		SourceType: "database",
		SourceID:   "db-1",
	}
	bs.CreateBackup(context.Background(), backup)
	bs.StartBackup(backup.ID)
	bs.CompleteBackup(backup.ID, 1024, "abc")

	latest := bs.GetLatestBackup("database", "db-1")
	if latest == nil {
		t.Fatal("Expected latest backup")
	}

	if latest.Name != "newer" {
		t.Errorf("Expected 'newer' backup, got %s", latest.Name)
	}
}

func TestBackupService_SearchBackups(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "production-backup",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-search",
	})

	bs.CreateBackup(context.Background(), &service.Backup{
		Name:     "staging-backup",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-search",
	})

	results := bs.SearchBackups("production")
	if len(results) == 0 {
		t.Error("Expected search results")
	}

	for _, b := range results {
		if b.Name != "production-backup" {
			t.Errorf("Expected 'production-backup', got %s", b.Name)
		}
	}
}

func TestBackupService_AddBackupTags(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	backup := &service.Backup{
		Name:     "tags-test",
		Type:     service.BackupTypeFull,
		TenantID: "tenant-tags",
	}
	bs.CreateBackup(context.Background(), backup)

	err := bs.AddBackupTags(backup.ID, []string{"important", "weekly"})
	if err != nil {
		t.Fatalf("Failed to add tags: %v", err)
	}

	retrieved, _ := bs.GetBackup(backup.ID)

	tagSet := make(map[string]bool)
	for _, t := range retrieved.Tags {
		tagSet[t] = true
	}

	if !tagSet["important"] || !tagSet["weekly"] {
		t.Error("Tags not added correctly")
	}
}

func TestBackupService_BackupTypes(t *testing.T) {
	types := []service.BackupType{
		service.BackupTypeFull,
		service.BackupTypeIncremental,
		service.BackupTypeDifferential,
		service.BackupTypeSnapshot,
	}

	dir, _ := os.MkdirTemp("", "backups")
	defer os.RemoveAll(dir)

	bs := service.NewBackupService(dir)

	for _, backupType := range types {
		backup := &service.Backup{
			Name:     string(backupType),
			Type:     backupType,
			TenantID: "tenant-types",
		}

		err := bs.CreateBackup(context.Background(), backup)
		if err != nil {
			t.Errorf("Failed to create %s backup: %v", backupType, err)
		}
	}
}

func TestBackupService_BackupToJSON(t *testing.T) {
	backup := &service.Backup{
		ID:        "backup-1",
		Name:      "test",
		Type:      service.BackupTypeFull,
		TenantID:  "tenant-1",
		Status:    service.BackupStatusCompleted,
		CreatedAt: time.Now(),
	}

	data, err := backup.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestBackupService_ScheduleToJSON(t *testing.T) {
	schedule := &service.BackupSchedule{
		ID:        "schedule-1",
		Name:      "Daily",
		Cron:      "0 2 * * *",
		TenantID:  "tenant-1",
		CreatedAt: time.Now(),
	}

	data, err := schedule.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}