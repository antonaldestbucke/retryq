package replay_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/replay"
	"github.com/example/retryq/storage"
)

// stubQueue records enqueued jobs.
type stubQueue struct {
	mu   sync.Mutex
	jobs []*job.Job
	err  error
}

func (q *stubQueue) Enqueue(j *job.Job) error {
	if q.err != nil {
		return q.err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, j)
	return nil
}

func newDeadStore(t *testing.T, jobs ...*job.Job) *storage.DeadLetterStore {
	t.Helper()
	store := storage.NewDeadLetterStore()
	for _, j := range jobs {
		if err := store.Add(j, "test error"); err != nil {
			t.Fatalf("setup: add job: %v", err)
		}
	}
	return store
}

func newJob(id string) *job.Job {
	return job.New(id, "test", []byte(`{}`))
}

func TestReplayAll_Empty(t *testing.T) {
	q := &stubQueue{}
	r := replay.New(storage.NewDeadLetterStore(), q)
	n, err := r.ReplayAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 replayed, got %d", n)
	}
}

func TestReplayAll_ReenqueuesAll(t *testing.T) {
	j1, j2 := newJob("a"), newJob("b")
	store := newDeadStore(t, j1, j2)
	q := &stubQueue{}
	r := replay.New(store, q)

	n, err := r.ReplayAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 replayed, got %d", n)
	}
	if len(q.jobs) != 2 {
		t.Errorf("expected 2 jobs in queue, got %d", len(q.jobs))
	}
	records, _ := store.List()
	if len(records) != 0 {
		t.Errorf("expected dead-letter store to be empty, got %d records", len(records))
	}
}

func TestReplayAll_WithResetAttempts(t *testing.T) {
	j := newJob("x")
	j.Attempts = 5
	j.LastError = "boom"
	j.RetryAt = time.Now().Add(time.Hour)
	store := newDeadStore(t, j)
	q := &stubQueue{}
	r := replay.New(store, q, replay.WithResetAttempts())

	_, err := r.ReplayAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(q.jobs))
	}
	got := q.jobs[0]
	if got.Attempts != 0 {
		t.Errorf("expected Attempts=0, got %d", got.Attempts)
	}
	if got.LastError != "" {
		t.Errorf("expected empty LastError, got %q", got.LastError)
	}
}

func TestReplayAll_EnqueueError(t *testing.T) {
	store := newDeadStore(t, newJob("fail"))
	q := &stubQueue{err: errors.New("queue full")}
	r := replay.New(store, q)

	n, err := r.ReplayAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != 0 {
		t.Errorf("expected 0 replayed, got %d", n)
	}
}

func TestReplayOne_Success(t *testing.T) {
	j := newJob("solo")
	store := newDeadStore(t, j)
	q := &stubQueue{}
	r := replay.New(store, q)

	if err := r.ReplayOne(context.Background(), "solo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(q.jobs))
	}
	records, _ := store.List()
	if len(records) != 0 {
		t.Errorf("expected dead-letter store empty, got %d", len(records))
	}
}

func TestReplayOne_NotFound(t *testing.T) {
	r := replay.New(storage.NewDeadLetterStore(), &stubQueue{})
	if err := r.ReplayOne(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for missing job, got nil")
	}
}

func TestReplayAll_ContextCancelled(t *testing.T) {
	store := newDeadStore(t, newJob("c1"), newJob("c2"))
	q := &stubQueue{}
	r := replay.New(store, q)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Should not panic; may replay 0 or partial.
	_, _ = r.ReplayAll(ctx)
}
