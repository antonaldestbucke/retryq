package priority_test

import (
	"testing"

	"github.com/example/retryq/job"
	"github.com/example/retryq/priority"
)

func newJob(id string) *job.Job {
	return job.New(id, "test", []byte(`{}`))
}

func TestNew_EmptyQueue(t *testing.T) {
	q := priority.New()
	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got len %d", q.Len())
	}
}

func TestDequeue_Empty(t *testing.T) {
	q := priority.New()
	if got := q.Dequeue(); got != nil {
		t.Fatalf("expected nil from empty queue, got %v", got)
	}
}

func TestEnqueue_IncreasesLen(t *testing.T) {
	q := priority.New()
	q.Enqueue(newJob("a"), priority.Normal)
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestDequeue_ReturnsHighestPriorityFirst(t *testing.T) {
	q := priority.New()
	low := newJob("low")
	norm := newJob("norm")
	high := newJob("high")

	q.Enqueue(low, priority.Low)
	q.Enqueue(high, priority.High)
	q.Enqueue(norm, priority.Normal)

	got := q.Dequeue()
	if got.ID != high.ID {
		t.Fatalf("expected high-priority job first, got %s", got.ID)
	}

	got = q.Dequeue()
	if got.ID != norm.ID {
		t.Fatalf("expected normal-priority job second, got %s", got.ID)
	}

	got = q.Dequeue()
	if got.ID != low.ID {
		t.Fatalf("expected low-priority job last, got %s", got.ID)
	}
}

func TestDequeue_SamePriorityFIFO(t *testing.T) {
	q := priority.New()
	a := newJob("a")
	b := newJob("b")
	q.Enqueue(a, priority.Normal)
	q.Enqueue(b, priority.Normal)

	first := q.Dequeue()
	if first.ID != a.ID {
		t.Fatalf("expected job 'a' first for equal priority, got %s", first.ID)
	}
}

func TestDequeue_DecreasesLen(t *testing.T) {
	q := priority.New()
	q.Enqueue(newJob("x"), priority.High)
	q.Dequeue()
	if q.Len() != 0 {
		t.Fatalf("expected len 0 after dequeue, got %d", q.Len())
	}
}
