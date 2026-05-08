package fanout_test

import (
	"errors"
	"testing"

	"github.com/example/retryq/fanout"
	"github.com/example/retryq/job"
)

// stubQueue is a minimal Enqueuer used in tests.
type stubQueue struct {
	jobs []*job.Job
	err  error
}

func (s *stubQueue) Enqueue(j *job.Job) error {
	if s.err != nil {
		return s.err
	}
	s.jobs = append(s.jobs, j)
	return nil
}

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func TestNew_RequiresAtLeastOneQueue(t *testing.T) {
	_, err := fanout.New()
	if err == nil {
		t.Fatal("expected error when no queues provided")
	}
}

func TestNew_ReturnsNonNil(t *testing.T) {
	q := &stubQueue{}
	f, err := fanout.New(q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil Fanout")
	}
}

func TestDispatch_SendsToAllQueues(t *testing.T) {
	q1, q2 := &stubQueue{}, &stubQueue{}
	f, _ := fanout.New(q1, q2)
	j := newJob(t)

	if err := f.Dispatch(j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q1.jobs) != 1 || len(q2.jobs) != 1 {
		t.Errorf("expected 1 job in each queue, got %d and %d", len(q1.jobs), len(q2.jobs))
	}
}

func TestDispatch_ContinuesPastFailure(t *testing.T) {
	q1 := &stubQueue{err: errors.New("enqueue failed")}
	q2 := &stubQueue{}
	f, _ := fanout.New(q1, q2)
	j := newJob(t)

	err := f.Dispatch(j)
	if err == nil {
		t.Fatal("expected error from failing queue")
	}
	// q2 should still have received the job
	if len(q2.jobs) != 1 {
		t.Errorf("expected q2 to receive job despite q1 failure")
	}
}

func TestDispatch_ReturnsDispatchError(t *testing.T) {
	q1 := &stubQueue{err: errors.New("oops")}
	q2 := &stubQueue{err: errors.New("also oops")}
	f, _ := fanout.New(q1, q2)

	err := f.Dispatch(newJob(t))
	var de *fanout.DispatchError
	if !errors.As(err, &de) {
		t.Fatalf("expected *fanout.DispatchError, got %T", err)
	}
	if len(de.Errs) != 2 {
		t.Errorf("expected 2 inner errors, got %d", len(de.Errs))
	}
}

func TestLen_ReturnsQueueCount(t *testing.T) {
	q1, q2, q3 := &stubQueue{}, &stubQueue{}, &stubQueue{}
	f, _ := fanout.New(q1, q2, q3)
	if f.Len() != 3 {
		t.Errorf("expected Len()=3, got %d", f.Len())
	}
}
