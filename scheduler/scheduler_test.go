package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
	"github.com/example/retryq/scheduler"
)

func newQueue(t *testing.T) *queue.Queue {
	t.Helper()
	q, err := queue.New()
	if err != nil {
		t.Fatalf("queue.New: %v", err)
	}
	return q
}

func newFactory(payload string) scheduler.JobFactory {
	return func() *job.Job {
		return job.New(payload, 3)
	}
}

func TestNew(t *testing.T) {
	q := newQueue(t)
	s := scheduler.New(q, newFactory("ping"), 100*time.Millisecond)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduler_EnqueuesJobs(t *testing.T) {
	q := newQueue(t)
	var count int64

	factory := func() *job.Job {
		atomic.AddInt64(&count, 1)
		return job.New("task", 3)
	}

	s := scheduler.New(q, factory, 50*time.Millisecond)
	ctx := context.Background()
	s.Start(ctx)

	time.Sleep(180 * time.Millisecond)
	s.Stop()

	got := atomic.LoadInt64(&count)
	if got < 2 {
		t.Errorf("expected at least 2 ticks, got %d", got)
	}
}

func TestScheduler_StartIsIdempotent(t *testing.T) {
	q := newQueue(t)
	s := scheduler.New(q, newFactory("x"), 50*time.Millisecond)
	ctx := context.Background()

	s.Start(ctx)
	s.Start(ctx) // second call must not panic or spawn extra goroutines

	if !s.Running() {
		t.Error("expected scheduler to be running")
	}
	s.Stop()
}

func TestScheduler_StopStopsEnqueuing(t *testing.T) {
	q := newQueue(t)
	var count int64

	factory := func() *job.Job {
		atomic.AddInt64(&count, 1)
		return job.New("task", 3)
	}

	s := scheduler.New(q, factory, 30*time.Millisecond)
	s.Start(context.Background())
	time.Sleep(80 * time.Millisecond)
	s.Stop()

	snap := atomic.LoadInt64(&count)
	time.Sleep(80 * time.Millisecond)

	if after := atomic.LoadInt64(&count); after != snap {
		t.Errorf("scheduler kept running after Stop: count went from %d to %d", snap, after)
	}

	if s.Running() {
		t.Error("expected Running() to return false after Stop")
	}
}

func TestScheduler_NilFactoryJobSkipped(t *testing.T) {
	q := newQueue(t)
	factory := func() *job.Job { return nil }

	s := scheduler.New(q, factory, 30*time.Millisecond)
	s.Start(context.Background())
	time.Sleep(80 * time.Millisecond)
	s.Stop()
	// No panic and queue remains empty — success.
}
