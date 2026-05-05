package queue_test

import (
	"testing"

	"github.com/user/retryq/job"
	"github.com/user/retryq/queue"
)

func newJob(id string) *job.Job {
	return job.New(id, "test-type", []byte(`{}`), 3)
}

func TestEnqueueAndDequeue(t *testing.T) {
	q := queue.New()

	j := newJob("job-1")
	if err := q.Enqueue(j); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}

	if q.PendingCount() != 1 {
		t.Fatalf("expected 1 pending job, got %d", q.PendingCount())
	}

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected dequeue error: %v", err)
	}
	if got == nil || got.ID != j.ID {
		t.Fatalf("expected job %s, got %v", j.ID, got)
	}
	if q.PendingCount() != 0 {
		t.Fatalf("expected 0 pending jobs after dequeue, got %d", q.PendingCount())
	}
}

func TestDequeue_Empty(t *testing.T) {
	q := queue.New()
	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected error on empty dequeue: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil job on empty queue, got %v", got)
	}
}

func TestRequeue_NotExhausted(t *testing.T) {
	q := queue.New()
	j := newJob("job-2")
	j.MarkFailed()

	if err := q.Requeue(j); err != nil {
		t.Fatalf("unexpected requeue error: %v", err)
	}
	if q.PendingCount() != 1 {
		t.Errorf("expected 1 pending, got %d", q.PendingCount())
	}
	if q.DeadLetterCount() != 0 {
		t.Errorf("expected 0 dead-letter, got %d", q.DeadLetterCount())
	}
}

func TestRequeue_Exhausted(t *testing.T) {
	q := queue.New()
	j := job.New("job-3", "test-type", []byte(`{}`), 1)
	j.MarkFailed()

	if err := q.Requeue(j); err != nil {
		t.Fatalf("unexpected requeue error: %v", err)
	}
	if q.PendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", q.PendingCount())
	}
	if q.DeadLetterCount() != 1 {
		t.Errorf("expected 1 dead-letter, got %d", q.DeadLetterCount())
	}
}

func TestClose(t *testing.T) {
	q := queue.New()
	q.Close()

	if err := q.Enqueue(newJob("job-4")); err != queue.ErrQueueClosed {
		t.Errorf("expected ErrQueueClosed on enqueue, got %v", err)
	}
	if _, err := q.Dequeue(); err != queue.ErrQueueClosed {
		t.Errorf("expected ErrQueueClosed on dequeue, got %v", err)
	}
}

func TestDeadLetterJobs_ReturnsCopy(t *testing.T) {
	q := queue.New()
	j := job.New("job-5", "test-type", []byte(`{}`), 1)
	j.MarkFailed()
	q.Requeue(j)

	dead := q.DeadLetterJobs()
	dead[0] = nil // mutate returned slice

	if q.DeadLetterJobs()[0] == nil {
		t.Error("DeadLetterJobs should return a copy, not a reference to internal slice")
	}
}
