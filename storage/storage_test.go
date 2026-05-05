package storage_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/storage"
)

func sampleRecord(id string) storage.Record {
	return storage.Record{
		ID:          id,
		Payload:     []byte(`{"task":"send_email"}`),
		Attempts:    1,
		MaxAttempts: 5,
		LastError:   "timeout",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Meta:        map[string]string{"queue": "default"},
	}
}

func TestMemoryStore_SaveAndLoad(t *testing.T) {
	s := storage.NewMemoryStore()
	r := sampleRecord("job-1")

	if err := s.Save(r); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	got, err := s.Load("job-1")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if got.ID != r.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, r.ID)
	}
	if got.LastError != r.LastError {
		t.Errorf("LastError mismatch: got %q, want %q", got.LastError, r.LastError)
	}
}

func TestMemoryStore_Load_NotFound(t *testing.T) {
	s := storage.NewMemoryStore()
	_, err := s.Load("missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := storage.NewMemoryStore()
	r := sampleRecord("job-2")
	_ = s.Save(r)

	if err := s.Delete("job-2"); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
	_, err := s.Load("job-2")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStore_List(t *testing.T) {
	s := storage.NewMemoryStore()
	for _, id := range []string{"a", "b", "c"} {
		_ = s.Save(sampleRecord(id))
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 records, got %d", len(list))
	}
}

func TestMemoryStore_Save_Overwrites(t *testing.T) {
	s := storage.NewMemoryStore()
	r := sampleRecord("job-3")
	_ = s.Save(r)

	r.Attempts = 4
	r.LastError = "connection refused"
	_ = s.Save(r)

	got, _ := s.Load("job-3")
	if got.Attempts != 4 {
		t.Errorf("expected Attempts=4, got %d", got.Attempts)
	}
	if got.LastError != "connection refused" {
		t.Errorf("unexpected LastError: %q", got.LastError)
	}
}
