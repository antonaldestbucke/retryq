// Package checkpoint provides job progress tracking, allowing long-running
// jobs to save intermediate state and resume from the last known checkpoint
// rather than restarting from scratch on failure.
package checkpoint

import (
	"errors"
	"sync"
	"time"
)

// ErrNotFound is returned when no checkpoint exists for the given job ID.
var ErrNotFound = errors.New("checkpoint: not found")

// Record holds the saved progress for a single job.
type Record struct {
	JobID     string
	Step      int
	Meta      map[string]string
	SavedAt   time.Time
}

// Store persists and retrieves checkpoint records.
type Store struct {
	mu      sync.RWMutex
	records map[string]Record
	clock   func() time.Time
}

// Option is a functional option for Store.
type Option func(*Store)

// WithClock overrides the clock used to timestamp records.
func WithClock(fn func() time.Time) Option {
	return func(s *Store) { s.clock = fn }
}

// New creates a new in-memory checkpoint Store.
func New(opts ...Option) *Store {
	s := &Store{
		records: make(map[string]Record),
		clock:   time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Save persists the current step and metadata for the given job ID.
func (s *Store) Save(jobID string, step int, meta map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	copy := make(map[string]string, len(meta))
	for k, v := range meta {
		copy[k] = v
	}
	s.records[jobID] = Record{
		JobID:   jobID,
		Step:    step,
		Meta:    copy,
		SavedAt: s.clock(),
	}
}

// Load retrieves the checkpoint for the given job ID.
// Returns ErrNotFound if no checkpoint exists.
func (s *Store) Load(jobID string) (Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.records[jobID]
	if !ok {
		return Record{}, ErrNotFound
	}
	return r, nil
}

// Delete removes the checkpoint for the given job ID.
func (s *Store) Delete(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, jobID)
}

// Len returns the number of stored checkpoints.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
