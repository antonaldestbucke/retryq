package dedup_test

import (
	"testing"
	"time"

	"github.com/example/retryq/dedup"
	"github.com/example/retryq/job"
)

func newJob(id string) *job.Job {
	j, _ := job.New(id, []byte(`{}`))
	return j
}

func TestIsDuplicate_FirstSeen(t *testing.T) {
	s := dedup.New()
	j := newJob("job-1")
	if s.IsDuplicate(j) {
		t.Fatal("expected first occurrence not to be a duplicate")
	}
}

func TestIsDuplicate_SecondSeen(t *testing.T) {
	s := dedup.New()
	j := newJob("job-2")
	s.IsDuplicate(j)
	if !s.IsDuplicate(j) {
		t.Fatal("expected second occurrence to be a duplicate")
	}
}

func TestIsDuplicate_AfterWindowExpiry(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	s := dedup.New(
		dedup.WithWindow(1*time.Second),
		dedup.WithClock(func() time.Time { return now }),
	)
	j := newJob("job-3")
	s.IsDuplicate(j)

	// Advance clock past the window.
	now = now.Add(2 * time.Second)
	if s.IsDuplicate(j) {
		t.Fatal("expected expired key not to be a duplicate")
	}
}

func TestRemove_AllowsReenqueue(t *testing.T) {
	s := dedup.New()
	j := newJob("job-4")
	s.IsDuplicate(j)
	s.Remove(j.ID)
	if s.IsDuplicate(j) {
		t.Fatal("expected key to be removed")
	}
}

func TestLen(t *testing.T) {
	s := dedup.New()
	s.IsDuplicate(newJob("a"))
	s.IsDuplicate(newJob("b"))
	if s.Len() != 2 {
		t.Fatalf("expected 2 keys, got %d", s.Len())
	}
}

func TestIsDuplicate_EmptyID(t *testing.T) {
	s := dedup.New()
	j := &struct{ ID string }{}
	_ = j
	// A job with an empty ID should never be flagged as duplicate.
	emptyJob := newJob("")
	if s.IsDuplicate(emptyJob) {
		t.Fatal("empty ID should not be deduplicated")
	}
}
