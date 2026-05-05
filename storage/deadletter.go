package storage

import (
	"sync"
	"time"
)

// DeadLetterEntry holds a record that has exhausted all retry attempts.
type DeadLetterEntry struct {
	Record    Record
	DeadAt    time.Time
	Reason    string
}

// DeadLetterStore persists exhausted jobs for later inspection or replay.
type DeadLetterStore struct {
	mu      sync.RWMutex
	entries map[string]DeadLetterEntry
}

// NewDeadLetterStore returns an initialised DeadLetterStore.
func NewDeadLetterStore() *DeadLetterStore {
	return &DeadLetterStore{entries: make(map[string]DeadLetterEntry)}
}

// Add inserts or replaces a dead-letter entry.
func (d *DeadLetterStore) Add(r Record, reason string) {
	d.mu.Lock()
	d.entries[r.ID] = DeadLetterEntry{
		Record: r,
		DeadAt: time.Now(),
		Reason: reason,
	}
	d.mu.Unlock()
}

// Get retrieves a dead-letter entry by job ID.
func (d *DeadLetterStore) Get(id string) (DeadLetterEntry, bool) {
	d.mu.RLock()
	e, ok := d.entries[id]
	d.mu.RUnlock()
	return e, ok
}

// Remove deletes a dead-letter entry, e.g. after manual replay.
func (d *DeadLetterStore) Remove(id string) {
	d.mu.Lock()
	delete(d.entries, id)
	d.mu.Unlock()
}

// All returns a snapshot of all dead-letter entries.
func (d *DeadLetterStore) All() []DeadLetterEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]DeadLetterEntry, 0, len(d.entries))
	for _, e := range d.entries {
		out = append(out, e)
	}
	return out
}

// Size returns the number of dead-letter entries.
func (d *DeadLetterStore) Size() int {
	d.mu.RLock()
	n := len(d.entries)
	d.mu.RUnlock()
	return n
}
