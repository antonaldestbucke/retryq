package metrics_test

import (
	"testing"

	"github.com/example/retryq/metrics"
)

func TestNew(t *testing.T) {
	m := metrics.New()
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
}

func TestSnapshot_ZeroValues(t *testing.T) {
	m := metrics.New()
	s := m.Snapshot()

	if s.Enqueued != 0 || s.Dequeued != 0 || s.Requeued != 0 ||
		s.Completed != 0 || s.Failed != 0 || s.DeadLettered != 0 {
		t.Errorf("expected all zero snapshot, got %+v", s)
	}
}

func TestSnapshot_AfterIncrements(t *testing.T) {
	m := metrics.New()

	m.Enqueued.Add(5)
	m.Dequeued.Add(4)
	m.Requeued.Add(2)
	m.Completed.Add(3)
	m.Failed.Add(1)
	m.DeadLettered.Add(1)

	s := m.Snapshot()

	if s.Enqueued != 5 {
		t.Errorf("Enqueued: want 5, got %d", s.Enqueued)
	}
	if s.Dequeued != 4 {
		t.Errorf("Dequeued: want 4, got %d", s.Dequeued)
	}
	if s.Requeued != 2 {
		t.Errorf("Requeued: want 2, got %d", s.Requeued)
	}
	if s.Completed != 3 {
		t.Errorf("Completed: want 3, got %d", s.Completed)
	}
	if s.Failed != 1 {
		t.Errorf("Failed: want 1, got %d", s.Failed)
	}
	if s.DeadLettered != 1 {
		t.Errorf("DeadLettered: want 1, got %d", s.DeadLettered)
	}
}

func TestSnapshot_IsImmutable(t *testing.T) {
	m := metrics.New()
	m.Enqueued.Add(1)

	s1 := m.Snapshot()
	m.Enqueued.Add(10)
	s2 := m.Snapshot()

	if s1.Enqueued != 1 {
		t.Errorf("s1.Enqueued should be 1, got %d", s1.Enqueued)
	}
	if s2.Enqueued != 11 {
		t.Errorf("s2.Enqueued should be 11, got %d", s2.Enqueued)
	}
}
