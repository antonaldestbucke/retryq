// Package dedup provides job deduplication support for retryq.
// It prevents duplicate jobs from being enqueued within a configurable window.
package dedup

import (
	"sync"
	"time"

	"github.com/example/retryq/job"
)

// Store tracks in-flight job keys to prevent duplicate enqueues.
type Store struct {
	mu      sync.Mutex
	keys    map[string]time.Time
	window  time.Duration
	clock   func() time.Time
}

// Option configures a Store.
type Option func(*Store)

// WithWindow sets the deduplication window duration.
func WithWindow(d time.Duration) Option {
	return func(s *Store) {
		s.window = d
	}
}

// WithClock overrides the clock used for expiry checks (useful in tests).
func WithClock(fn func() time.Time) Option {
	return func(s *Store) {
		s.clock = fn
	}
}

// New creates a new deduplication Store.
func New(opts ...Option) *Store {
	s := &Store{
		keys:   make(map[string]time.Time),
		window: 5 * time.Minute,
		clock:  time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// IsDuplicate reports whether the given job's key has been seen within the window.
// If not a duplicate, the key is recorded.
func (s *Store) IsDuplicate(j *job.Job) bool {
	key := j.ID
	if key == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	s.evict(now)
	if _, exists := s.keys[key]; exists {
		return true
	}
	s.keys[key] = now.Add(s.window)
	return false
}

// Remove explicitly removes a key from the dedup store.
func (s *Store) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, key)
}

// Len returns the number of tracked keys.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.keys)
}

// evict removes expired keys. Must be called with s.mu held.
func (s *Store) evict(now time.Time) {
	for k, exp := range s.keys {
		if now.After(exp) {
			delete(s.keys, k)
		}
	}
}
