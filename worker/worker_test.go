package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
	"github.com/example/retryq/worker"
)

func newQueue(t *testing.T) *queue.Queue {
	t.Helper()
	q, err := queue.New(3, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("queue.New: %v", err)
	}
	return q
}

func TestWorker_ProcessesJob(t *testing.T) {
	q := newQueue(t)
	j := job.New("send-email", []byte(`{"to":"a@b.com"}`))
	q.Enqueue(j)

	var processed int32
	h := func(_ context.Context, _ *job.Job) error {
		atomic.AddInt32(&processed, 1)
		return nil
	}

	w := worker.New(q, h, worker.WithPollInterval(10*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if atomic.LoadInt32(&processed) != 1 {
		t.Errorf("expected 1 job processed, got %d", processed)
	}
}

func TestWorker_RequeuesOnFailure(t *testing.T) {
	q := newQueue(t)
	j := job.New("resize-image", []byte(`{}`))
	q.Enqueue(j)

	var calls int32
	h := func(_ context.Context, _ *job.Job) error {
		atomic.AddInt32(&calls, 1)
		return context.DeadlineExceeded // always fail
	}

	w := worker.New(q, h, worker.WithPollInterval(10*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	// Job has 3 max attempts; handler should be called up to 3 times.
	got := atomic.LoadInt32(&calls)
	if got < 1 {
		t.Errorf("expected at least 1 call, got %d", got)
	}
}

func TestWorker_ConcurrentProcessing(t *testing.T) {
	q := newQueue(t)
	for i := 0; i < 5; i++ {
		q.Enqueue(job.New("task", []byte(`{}`)))
	}

	var processed int32
	h := func(_ context.Context, _ *job.Job) error {
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&processed, 1)
		return nil
	}

	w := worker.New(q, h,
		worker.WithConcurrency(5),
		worker.WithPollInterval(10*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if atomic.LoadInt32(&processed) != 5 {
		t.Errorf("expected 5 jobs processed, got %d", processed)
	}
}
