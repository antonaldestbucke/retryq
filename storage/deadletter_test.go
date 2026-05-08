package storage_test

import (
	"testing"
	"time"

	"github.com/example/retryq/storage"
)

func deadRecord(id string) storage.Record {
	return storage.Record{
		ID:          id,
		Payload:     []byte(`{"task":"notify"}`),
		Attempts:    5,
		MaxAttempts: 5,
		LastError:   "service unavailable",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestDeadLetterStore_AddAndGet(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	r := deadRecord("dl-1")
	dl.Add(r, "max retries exceeded")

	e, ok := dl.Get("dl-1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Record.ID != "dl-1" {
		t.Errorf("ID mismatch: got %q", e.Record.ID)
	}
	if e.Reason != "max retries exceeded" {
		t.Errorf("Reason mismatch: got %q", e.Reason)
	}
	if e.DeadAt.IsZero() {
		t.Error("DeadAt should not be zero")
	}
}

func TestDeadLetterStore_Get_Missing(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	_, ok := dl.Get("nope")
	if ok {
		t.Error("expected ok=false for missing entry")
	}
}

func TestDeadLetterStore_Remove(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	dl.Add(deadRecord("dl-2"), "exhausted")
	dl.Remove("dl-2")

	_, ok := dl.Get("dl-2")
	if ok {
		t.Error("expected entry to be removed")
	}
}

// TestDeadLetterStore_Remove_Missing verifies that removing a non-existent
// entry does not panic or affect the store's size.
func TestDeadLetterStore_Remove_Missing(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	dl.Add(deadRecord("dl-3"), "exhausted")
	dl.Remove("does-not-exist")

	if dl.Size() != 1 {
		t.Errorf("expected size 1 after removing missing entry, got %d", dl.Size())
	}
}

func TestDeadLetterStore_All(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	for _, id := range []string{"x", "y", "z"} {
		dl.Add(deadRecord(id), "done")
	}

	all := dl.All()
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
}

func TestDeadLetterStore_Size(t *testing.T) {
	dl := storage.NewDeadLetterStore()
	if dl.Size() != 0 {
		t.Error("expected size 0 on empty store")
	}
	dl.Add(deadRecord("s1"), "err")
	dl.Add(deadRecord("s2"), "err")
	if dl.Size() != 2 {
		t.Errorf("expected size 2, got %d", dl.Size())
	}
}
