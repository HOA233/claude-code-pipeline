package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// LockService provides distributed locking capabilities
type LockService struct {
	mu      sync.RWMutex
	locks   map[string]*Lock
	waiters map[string][]*LockWaiter
}

// Lock represents a distributed lock
type Lock struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	Token     string    `json:"token"`
	TTL       int       `json:"ttl"` // seconds
	AcquiredAt time.Time `json:"acquired_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LockWaiter represents a lock waiter
type LockWaiter struct {
	Owner     string
	Token     string
	Notified  chan struct{}
	ExpiresAt time.Time
}

// LockOptions for acquiring locks
type LockOptions struct {
	Key      string                 `json:"key"`
	Owner    string                 `json:"owner"`
	TTL      int                    `json:"ttl"` // seconds
	Wait     bool                   `json:"wait"` // wait for lock if unavailable
	Timeout  int                    `json:"timeout"` // max wait time in seconds
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewLockService creates a new lock service
func NewLockService() *LockService {
	return &LockService{
		locks:   make(map[string]*Lock),
		waiters: make(map[string][]*LockWaiter),
	}
}

// Acquire acquires a lock
func (s *LockService) Acquire(ctx context.Context, opts *LockOptions) (*Lock, error) {
	if opts.Key == "" {
		return nil, errors.New("key is required")
	}
	if opts.Owner == "" {
		return nil, errors.New("owner is required")
	}

	ttl := opts.TTL
	if ttl == 0 {
		ttl = 30 // default 30 seconds
	}

	token := generateID()
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if lock exists and is not expired
	if existing, exists := s.locks[opts.Key]; exists {
		if time.Now().Before(existing.ExpiresAt) {
			// Lock is held by someone else
			if opts.Wait {
				// Add to waiters
				waiter := &LockWaiter{
					Owner:     opts.Owner,
					Token:     token,
					Notified:  make(chan struct{}),
					ExpiresAt: now.Add(time.Duration(opts.Timeout) * time.Second),
				}
				s.waiters[opts.Key] = append(s.waiters[opts.Key], waiter)
				s.mu.Unlock() // Unlock to allow other operations

				// Wait for notification or timeout
				timeout := time.Duration(opts.Timeout) * time.Second
				if opts.Timeout == 0 {
					timeout = time.Minute // default wait
				}

				select {
				case <-waiter.Notified:
					s.mu.Lock()
					// Lock was released, try to acquire again
					lock := &Lock{
						ID:         generateID(),
						Key:        opts.Key,
						Owner:      opts.Owner,
						Token:      token,
						TTL:        ttl,
						AcquiredAt: time.Now(),
						ExpiresAt:  time.Now().Add(time.Duration(ttl) * time.Second),
						Metadata:   opts.Metadata,
					}
					s.locks[opts.Key] = lock
					return lock, nil
				case <-time.After(timeout):
					s.mu.Lock()
					// Remove waiter
					s.removeWaiter(opts.Key, token)
					return nil, errors.New("lock acquisition timeout")
				case <-ctx.Done():
					s.mu.Lock()
					s.removeWaiter(opts.Key, token)
					return nil, ctx.Err()
				}
			}
			return nil, fmt.Errorf("lock is held by %s", existing.Owner)
		}
		// Lock expired, remove it
		delete(s.locks, opts.Key)
	}

	// Acquire the lock
	lock := &Lock{
		ID:         generateID(),
		Key:        opts.Key,
		Owner:      opts.Owner,
		Token:      token,
		TTL:        ttl,
		AcquiredAt: now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
		Metadata:   opts.Metadata,
	}

	s.locks[opts.Key] = lock

	return lock, nil
}

// Release releases a lock
func (s *LockService) Release(ctx context.Context, key, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[key]
	if !exists {
		return errors.New("lock not found")
	}

	if lock.Token != token {
		return errors.New("invalid lock token")
	}

	delete(s.locks, key)

	// Notify next waiter
	if waiters := s.waiters[key]; len(waiters) > 0 {
		waiter := waiters[0]
		s.waiters[key] = waiters[1:]
		close(waiter.Notified)
	}

	return nil
}

// Extend extends a lock's TTL
func (s *LockService) Extend(ctx context.Context, key, token string, additionalTTL int) (*Lock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[key]
	if !exists {
		return nil, errors.New("lock not found")
	}

	if lock.Token != token {
		return nil, errors.New("invalid lock token")
	}

	if time.Now().After(lock.ExpiresAt) {
		delete(s.locks, key)
		return nil, errors.New("lock has expired")
	}

	lock.ExpiresAt = time.Now().Add(time.Duration(additionalTTL) * time.Second)
	lock.TTL = lock.TTL + additionalTTL

	return lock, nil
}

// GetLock gets lock information
func (s *LockService) GetLock(key string) (*Lock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lock, exists := s.locks[key]
	if !exists {
		return nil, errors.New("lock not found")
	}

	// Check if expired
	if time.Now().After(lock.ExpiresAt) {
		return nil, errors.New("lock has expired")
	}

	return lock, nil
}

// IsLocked checks if a key is locked
func (s *LockService) IsLocked(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lock, exists := s.locks[key]
	if !exists {
		return false
	}

	return time.Now().Before(lock.ExpiresAt)
}

// ListLocks lists all locks
func (s *LockService) ListLocks(owner string) []*Lock {
	s.mu.RLock()
	defer s.mu.RUnlock()

	locks := make([]*Lock, 0)
	now := time.Now()

	for _, lock := range s.locks {
		if now.After(lock.ExpiresAt) {
			continue
		}
		if owner == "" || lock.Owner == owner {
			locks = append(locks, lock)
		}
	}

	return locks
}

// ForceRelease forcefully releases a lock (admin operation)
func (s *LockService) ForceRelease(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.locks, key)

	// Notify all waiters
	for _, waiter := range s.waiters[key] {
		close(waiter.Notified)
	}
	delete(s.waiters, key)

	return nil
}

// CleanupExpired removes expired locks
func (s *LockService) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for key, lock := range s.locks {
		if now.After(lock.ExpiresAt) {
			delete(s.locks, key)
			count++
		}
	}

	return count
}

// removeWaiter removes a waiter from the queue
func (s *LockService) removeWaiter(key, token string) {
	waiters := s.waiters[key]
	for i, w := range waiters {
		if w.Token == token {
			s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
			break
		}
	}
}

// GetStats returns lock service statistics
func (s *LockService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := 0
	expired := 0
	now := time.Now()

	for _, lock := range s.locks {
		if now.Before(lock.ExpiresAt) {
			active++
		} else {
			expired++
		}
	}

	waiters := 0
	for _, w := range s.waiters {
		waiters += len(w)
	}

	return map[string]interface{}{
		"active_locks":  active,
		"expired_locks": expired,
		"total_locks":   len(s.locks),
		"waiters":       waiters,
	}
}

// Mutex provides a simpler mutex-style interface
type Mutex struct {
	service *LockService
	key     string
	token   string
}

// NewMutex creates a mutex for a key
func NewMutex(service *LockService, key string) *Mutex {
	return &Mutex{
		service: service,
		key:     key,
	}
}

// Lock acquires the mutex
func (m *Mutex) Lock(ctx context.Context, owner string, ttl int) error {
	lock, err := m.service.Acquire(ctx, &LockOptions{
		Key:   m.key,
		Owner: owner,
		TTL:   ttl,
	})
	if err != nil {
		return err
	}
	m.token = lock.Token
	return nil
}

// Unlock releases the mutex
func (m *Mutex) Unlock(ctx context.Context) error {
	return m.service.Release(ctx, m.key, m.token)
}

// WithLock executes a function while holding a lock
func (s *LockService) WithLock(ctx context.Context, key, owner string, ttl int, fn func() error) error {
	lock, err := s.Acquire(ctx, &LockOptions{
		Key:   key,
		Owner: owner,
		TTL:   ttl,
		Wait:  true,
	})
	if err != nil {
		return err
	}

	defer s.Release(ctx, key, lock.Token)

	return fn()
}

// Semaphore provides a counting semaphore
type Semaphore struct {
	service *SemaphoreService
	key     string
	count   int
	tokens  []string
}

// SemaphoreService provides semaphore operations
type SemaphoreService struct {
	mu     sync.RWMutex
	semaphores map[string]*semaphoreState
}

type semaphoreState struct {
	count    int
	max      int
	acquired []string // tokens
	waiters  []*semWaiter
}

type semWaiter struct {
	token    string
	notified chan struct{}
}

// NewSemaphoreService creates a new semaphore service
func NewSemaphoreService() *SemaphoreService {
	return &SemaphoreService{
		semaphores: make(map[string]*semaphoreState),
	}
}

// CreateSemaphore creates a semaphore with max count
func (s *SemaphoreService) CreateSemaphore(key string, max int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.semaphores[key]; exists {
		return errors.New("semaphore already exists")
	}

	s.semaphores[key] = &semaphoreState{
		count:    0,
		max:      max,
		acquired: make([]string, 0),
		waiters:  make([]*semWaiter, 0),
	}

	return nil
}

// Acquire acquires a semaphore permit
func (s *SemaphoreService) Acquire(ctx context.Context, key string) (string, error) {
	s.mu.Lock()

	state, exists := s.semaphores[key]
	if !exists {
		s.mu.Unlock()
		return "", errors.New("semaphore not found")
	}

	if state.count < state.max {
		token := generateID()
		state.count++
		state.acquired = append(state.acquired, token)
		s.mu.Unlock()
		return token, nil
	}

	// Wait for availability
	waiter := &semWaiter{
		token:    generateID(),
		notified: make(chan struct{}),
	}
	state.waiters = append(state.waiters, waiter)
	s.mu.Unlock()

	select {
	case <-waiter.notified:
		s.mu.Lock()
		state.count++
		state.acquired = append(state.acquired, waiter.token)
		s.mu.Unlock()
		return waiter.token, nil
	case <-ctx.Done():
		s.mu.Lock()
		// Remove waiter
		for i, w := range state.waiters {
			if w.token == waiter.token {
				state.waiters = append(state.waiters[:i], state.waiters[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		return "", ctx.Err()
	}
}

// Release releases a semaphore permit
func (s *SemaphoreService) Release(key, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.semaphores[key]
	if !exists {
		return errors.New("semaphore not found")
	}

	// Find and remove token
	for i, t := range state.acquired {
		if t == token {
			state.acquired = append(state.acquired[:i], state.acquired[i+1:]...)
			state.count--

			// Notify next waiter
			if len(state.waiters) > 0 {
				waiter := state.waiters[0]
				state.waiters = state.waiters[1:]
				close(waiter.notified)
			}

			return nil
		}
	}

	return errors.New("invalid token")
}

// ToJSON serializes a lock to JSON
func (l *Lock) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}