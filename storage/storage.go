// Package storage provides persistence interfaces and implementations
// for retryq job queues, enabling dead-letter and retry state to survive
// process restarts.
package storage

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ErrNotFound is returned when a job cannot be located by ID.
var ErrNotFound = errors.New("storage: job not found")

// Record is the persisted representation of a job.
type Record struct {
	ID          string            `json:"id"`
	Payload     []byte            `json:"payload"`
	Attempts    int               `json:"attempts"`
	MaxAttempts int               `json:"max_attempts"`
	LastError   string            `json:"last_error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// Store defines the interface for job persistence backends.
type Store interface {
	// Save persists or updates a record.
	Save(r Record) error
	// Load retrieves a record by ID.
	Load(id string) (Record, error)
	// Delete removes a record by ID.
	Delete(id string) error
	// List returns all stored records.
	List() ([]Record, error)
}

// MemoryStore is an in-memory Store implementation, safe for concurrent use.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewMemoryStore returns an initialised MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string][]byte)}
}

func (m *MemoryStore) Save(r Record) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.data[r.ID] = b
	m.mu.Unlock()
	return nil
}

func (m *MemoryStore) Load(id string) (Record, error) {
	m.mu.RLock()
	b, ok := m.data[id]
	m.mu.RUnlock()
	if !ok {
		return Record{}, ErrNotFound
	}
	var r Record
	return r, json.Unmarshal(b, &r)
}

func (m *MemoryStore) Delete(id string) error {
	m.mu.Lock()
	delete(m.data, id)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStore) List() ([]Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	records := make([]Record, 0, len(m.data))
	for _, b := range m.data {
		var r Record
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
