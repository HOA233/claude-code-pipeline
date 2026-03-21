package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

func TestLockService_New(t *testing.T) {
	lock := service.NewLockService()
	if lock == nil {
		t.Fatal("Expected non-nil lock service")
	}
}

func TestLockService_Acquire(t *testing.T) {
	ls := service.NewLockService()

	lock, err := ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "test-lock",
		Owner: "owner-1",
		TTL:   30,
	})

	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if lock.Key != "test-lock" {
		t.Errorf("Expected key 'test-lock', got '%s'", lock.Key)
	}
	if lock.Owner != "owner-1" {
		t.Errorf("Expected owner 'owner-1', got '%s'", lock.Owner)
	}
	if lock.Token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestLockService_Acquire_MissingKey(t *testing.T) {
	ls := service.NewLockService()

	_, err := ls.Acquire(context.Background(), &service.LockOptions{
		Owner: "owner-1",
	})

	if err == nil {
		t.Error("Expected error for missing key")
	}
}

func TestLockService_Acquire_MissingOwner(t *testing.T) {
	ls := service.NewLockService()

	_, err := ls.Acquire(context.Background(), &service.LockOptions{
		Key: "test-lock",
	})

	if err == nil {
		t.Error("Expected error for missing owner")
	}
}

func TestLockService_Acquire_AlreadyLocked(t *testing.T) {
	ls := service.NewLockService()

	// First acquisition
	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "locked-key",
		Owner: "owner-1",
		TTL:   30,
	})

	// Second acquisition should fail
	_, err := ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "locked-key",
		Owner: "owner-2",
		TTL:   30,
	})

	if err == nil {
		t.Error("Expected error for already locked key")
	}
}

func TestLockService_Release(t *testing.T) {
	ls := service.NewLockService()

	lock, _ := ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "release-test",
		Owner: "owner-1",
		TTL:   30,
	})

	err := ls.Release(context.Background(), "release-test", lock.Token)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Should be able to acquire again
	_, err = ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "release-test",
		Owner: "owner-2",
		TTL:   30,
	})

	if err != nil {
		t.Error("Expected to acquire lock after release")
	}
}

func TestLockService_Release_InvalidToken(t *testing.T) {
	ls := service.NewLockService()

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "invalid-token-test",
		Owner: "owner-1",
		TTL:   30,
	})

	err := ls.Release(context.Background(), "invalid-token-test", "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestLockService_Release_NotFound(t *testing.T) {
	ls := service.NewLockService()

	err := ls.Release(context.Background(), "nonexistent", "token")
	if err == nil {
		t.Error("Expected error for nonexistent lock")
	}
}

func TestLockService_Extend(t *testing.T) {
	ls := service.NewLockService()

	lock, _ := ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "extend-test",
		Owner: "owner-1",
		TTL:   10,
	})

	originalExpiry := lock.ExpiresAt

	extended, err := ls.Extend(context.Background(), "extend-test", lock.Token, 20)
	if err != nil {
		t.Fatalf("Failed to extend lock: %v", err)
	}

	if !extended.ExpiresAt.After(originalExpiry) {
		t.Error("Expected expiry to be extended")
	}
}

func TestLockService_Extend_InvalidToken(t *testing.T) {
	ls := service.NewLockService()

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "extend-invalid",
		Owner: "owner-1",
		TTL:   10,
	})

	_, err := ls.Extend(context.Background(), "extend-invalid", "invalid-token", 20)
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestLockService_GetLock(t *testing.T) {
	ls := service.NewLockService()

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "get-test",
		Owner: "owner-1",
		TTL:   30,
	})

	lock, err := ls.GetLock("get-test")
	if err != nil {
		t.Fatalf("Failed to get lock: %v", err)
	}

	if lock.Owner != "owner-1" {
		t.Error("Lock owner mismatch")
	}
}

func TestLockService_GetLock_NotFound(t *testing.T) {
	ls := service.NewLockService()

	_, err := ls.GetLock("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent lock")
	}
}

func TestLockService_IsLocked(t *testing.T) {
	ls := service.NewLockService()

	if ls.IsLocked("is-locked-test") {
		t.Error("Expected key to not be locked")
	}

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "is-locked-test",
		Owner: "owner-1",
		TTL:   30,
	})

	if !ls.IsLocked("is-locked-test") {
		t.Error("Expected key to be locked")
	}
}

func TestLockService_ListLocks(t *testing.T) {
	ls := service.NewLockService()

	for i := 0; i < 3; i++ {
		ls.Acquire(context.Background(), &service.LockOptions{
			Key:   string(rune('a' + i)),
			Owner: "owner-1",
			TTL:   30,
		})
	}

	locks := ls.ListLocks("owner-1")
	if len(locks) < 3 {
		t.Errorf("Expected at least 3 locks, got %d", len(locks))
	}
}

func TestLockService_ForceRelease(t *testing.T) {
	ls := service.NewLockService()

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "force-release-test",
		Owner: "owner-1",
		TTL:   30,
	})

	err := ls.ForceRelease(context.Background(), "force-release-test")
	if err != nil {
		t.Fatalf("Failed to force release: %v", err)
	}

	if ls.IsLocked("force-release-test") {
		t.Error("Expected key to be unlocked after force release")
	}
}

func TestLockService_CleanupExpired(t *testing.T) {
	ls := service.NewLockService()

	// Create a lock with short TTL
	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "expired-lock",
		Owner: "owner-1",
		TTL:   0, // Will expire immediately
	})

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	count := ls.CleanupExpired()
	if count < 1 {
		t.Errorf("Expected at least 1 expired lock cleaned, got %d", count)
	}
}

func TestLockService_GetStats(t *testing.T) {
	ls := service.NewLockService()

	ls.Acquire(context.Background(), &service.LockOptions{
		Key:   "stats-test",
		Owner: "owner-1",
		TTL:   30,
	})

	stats := ls.GetStats()

	if stats["active_locks"].(int) < 1 {
		t.Error("Expected at least 1 active lock")
	}
}

func TestLockService_WithLock(t *testing.T) {
	ls := service.NewLockService()

	executed := false
	err := ls.WithLock(context.Background(), "with-lock-test", "owner-1", 30, func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("WithLock failed: %v", err)
	}

	if !executed {
		t.Error("Expected function to be executed")
	}

	// Lock should be released after function
	if ls.IsLocked("with-lock-test") {
		t.Error("Expected lock to be released after WithLock")
	}
}

func TestMutex_Lock(t *testing.T) {
	ls := service.NewLockService()
	m := service.NewMutex(ls, "mutex-test")

	err := m.Lock(context.Background(), "owner-1", 30)
	if err != nil {
		t.Fatalf("Failed to lock mutex: %v", err)
	}

	if !ls.IsLocked("mutex-test") {
		t.Error("Expected mutex to be locked")
	}
}

func TestMutex_Unlock(t *testing.T) {
	ls := service.NewLockService()
	m := service.NewMutex(ls, "mutex-unlock-test")

	m.Lock(context.Background(), "owner-1", 30)
	err := m.Unlock(context.Background())

	if err != nil {
		t.Fatalf("Failed to unlock mutex: %v", err)
	}

	if ls.IsLocked("mutex-unlock-test") {
		t.Error("Expected mutex to be unlocked")
	}
}

func TestSemaphoreService_New(t *testing.T) {
	ss := service.NewSemaphoreService()
	if ss == nil {
		t.Fatal("Expected non-nil semaphore service")
	}
}

func TestSemaphoreService_CreateSemaphore(t *testing.T) {
	ss := service.NewSemaphoreService()

	err := ss.CreateSemaphore("test-sem", 3)
	if err != nil {
		t.Fatalf("Failed to create semaphore: %v", err)
	}
}

func TestSemaphoreService_CreateSemaphore_Duplicate(t *testing.T) {
	ss := service.NewSemaphoreService()

	ss.CreateSemaphore("dup-sem", 3)
	err := ss.CreateSemaphore("dup-sem", 3)

	if err == nil {
		t.Error("Expected error for duplicate semaphore")
	}
}

func TestSemaphoreService_Acquire(t *testing.T) {
	ss := service.NewSemaphoreService()
	ss.CreateSemaphore("acquire-test", 3)

	token, err := ss.Acquire(context.Background(), "acquire-test")
	if err != nil {
		t.Fatalf("Failed to acquire semaphore: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestSemaphoreService_Acquire_NotFound(t *testing.T) {
	ss := service.NewSemaphoreService()

	_, err := ss.Acquire(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent semaphore")
	}
}

func TestSemaphoreService_Release(t *testing.T) {
	ss := service.NewSemaphoreService()
	ss.CreateSemaphore("release-test", 3)

	token, _ := ss.Acquire(context.Background(), "release-test")
	err := ss.Release("release-test", token)

	if err != nil {
		t.Fatalf("Failed to release semaphore: %v", err)
	}
}

func TestSemaphoreService_Release_InvalidToken(t *testing.T) {
	ss := service.NewSemaphoreService()
	ss.CreateSemaphore("invalid-token-sem", 3)

	err := ss.Release("invalid-token-sem", "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestSemaphoreService_MaxCount(t *testing.T) {
	ss := service.NewSemaphoreService()
	ss.CreateSemaphore("max-count", 2)

	// Acquire twice
	ss.Acquire(context.Background(), "max-count")
	ss.Acquire(context.Background(), "max-count")

	// Third acquire should wait
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := ss.Acquire(ctx, "max-count")
	if err == nil {
		t.Error("Expected timeout for max count reached")
	}
}