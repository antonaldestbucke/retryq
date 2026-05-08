package checkpoint_test

import (
	"testing"
	"time"

	"github.com/example/retryq/checkpoint"
)

var fixedTime = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedTime }

func TestSave_And_Load(t *testing.T) {
	s := checkpoint.New(checkpoint.WithClock(fixedClock))

	meta := map[string]string{"cursor": "42", "batch": "3"}
	s.Save("job-1", 5, meta)

	r, err := s.Load("job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.JobID != "job-1" {
		t.Errorf("expected job-1, got %s", r.JobID)
	}
	if r.Step != 5 {
		t.Errorf("expected step 5, got %d", r.Step)
	}
	if r.Meta["cursor"] != "42" {
		t.Errorf("expected cursor=42, got %s", r.Meta["cursor"])
	}
	if !r.SavedAt.Equal(fixedTime) {
		t.Errorf("unexpected SavedAt: %v", r.SavedAt)
	}
}

func TestLoad_NotFound(t *testing.T) {
	s := checkpoint.New()
	_, err := s.Load("missing")
	if err != checkpoint.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSave_MetaIsolation(t *testing.T) {
	s := checkpoint.New()
	meta := map[string]string{"key": "original"}
	s.Save("job-2", 1, meta)

	// Mutate the original map after saving.
	meta["key"] = "mutated"

	r, _ := s.Load("job-2")
	if r.Meta["key"] != "original" {
		t.Errorf("expected isolated copy, got %s", r.Meta["key"])
	}
}

func TestDelete_RemovesRecord(t *testing.T) {
	s := checkpoint.New()
	s.Save("job-3", 2, nil)
	s.Delete("job-3")

	_, err := s.Load("job-3")
	if err != checkpoint.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestLen_TracksCount(t *testing.T) {
	s := checkpoint.New()
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.Save("a", 0, nil)
	s.Save("b", 1, nil)
	if s.Len() != 2 {
		t.Errorf("expected 2, got %d", s.Len())
	}
	s.Delete("a")
	if s.Len() != 1 {
		t.Errorf("expected 1 after delete, got %d", s.Len())
	}
}

func TestSave_Overwrites(t *testing.T) {
	s := checkpoint.New()
	s.Save("job-4", 1, map[string]string{"x": "1"})
	s.Save("job-4", 9, map[string]string{"x": "9"})

	r, _ := s.Load("job-4")
	if r.Step != 9 {
		t.Errorf("expected step 9, got %d", r.Step)
	}
}
