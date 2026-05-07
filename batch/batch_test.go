package batch_test

import (
	"errors"
	"testing"

	"github.com/example/retryq/batch"
	"github.com/example/retryq/job"
)

// stubQueue is a minimal Enqueuer that can be configured to fail.
type stubQueue struct {
	failIDs map[string]bool
	got     []*job.Job
}

func (s *stubQueue) Enqueue(j *job.Job) error {
	if s.failIDs[j.ID] {
		return errors.New("enqueue failed")
	}
	s.got = append(s.got, j)
	return nil
}

func newJobs(n int) []*job.Job {
	out := make([]*job.Job, n)
	for i := range out {
		out[i] = job.New([]byte("payload"), 3)
	}
	return out
}

func TestEnqueue_AllSucceed(t *testing.T) {
	q := &stubQueue{}
	jobs := newJobs(5)
	s := batch.Enqueue(q, jobs)

	if s.Total != 5 || s.Enqueued != 5 || s.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", s)
	}
	if s.HasErrors() {
		t.Fatal("expected no errors")
	}
}

func TestEnqueue_SomeFailures(t *testing.T) {
	jobs := newJobs(3)
	q := &stubQueue{failIDs: map[string]bool{jobs[1].ID: true}}
	s := batch.Enqueue(q, jobs)

	if s.Total != 3 || s.Enqueued != 2 || s.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
	if !s.HasErrors() {
		t.Fatal("expected errors")
	}
	if s.Errors[0].Job.ID != jobs[1].ID {
		t.Errorf("wrong failed job: got %s want %s", s.Errors[0].Job.ID, jobs[1].ID)
	}
}

func TestEnqueueStrict_NoError(t *testing.T) {
	q := &stubQueue{}
	err := batch.EnqueueStrict(q, newJobs(3))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestEnqueueStrict_ReturnsJoinedErrors(t *testing.T) {
	jobs := newJobs(2)
	q := &stubQueue{failIDs: map[string]bool{jobs[0].ID: true, jobs[1].ID: true}}
	err := batch.EnqueueStrict(q, jobs)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestFromPayloads(t *testing.T) {
	q := &stubQueue{}
	payloads := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	s := batch.FromPayloads(q, payloads, 5)

	if s.Total != 3 || s.Enqueued != 3 {
		t.Fatalf("unexpected summary: %+v", s)
	}
	if len(q.got) != 3 {
		t.Fatalf("expected 3 enqueued, got %d", len(q.got))
	}
}
