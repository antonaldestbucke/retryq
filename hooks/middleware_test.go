package hooks_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/hooks"
	"github.com/example/retryq/job"
)

func newExhaustedJob() *job.Job {
	j := job.New("test-type", []byte(`{}`), 1, time.Millisecond)
	j.MarkFailed(errors.New("first failure")) // exhausts maxAttempts=1
	return j
}

func TestWithHooks_EmitsSucceeded(t *testing.T) {
	r := hooks.New()
	var emitted hooks.EventType
	r.On(hooks.EventSucceeded, func(e hooks.EventType, j *job.Job) { emitted = e })

	h := hooks.WithHooks(r, func(j *job.Job) error { return nil })
	if err := h(newJob()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emitted != hooks.EventSucceeded {
		t.Errorf("expected EventSucceeded, got %q", emitted)
	}
}

func TestWithHooks_EmitsFailed_WhenNotExhausted(t *testing.T) {
	r := hooks.New()
	var emitted hooks.EventType
	r.On(hooks.EventFailed, func(e hooks.EventType, j *job.Job) { emitted = e })

	handlerErr := errors.New("transient error")
	h := hooks.WithHooks(r, func(j *job.Job) error { return handlerErr })

	j := newJob() // maxAttempts=3, attempts=0 → not exhausted after handler
	err := h(j)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handlerErr, got %v", err)
	}
	if emitted != hooks.EventFailed {
		t.Errorf("expected EventFailed, got %q", emitted)
	}
}

func TestWithHooks_EmitsDeadLetter_WhenExhausted(t *testing.T) {
	r := hooks.New()
	var emitted hooks.EventType
	r.On(hooks.EventDeadLetter, func(e hooks.EventType, j *job.Job) { emitted = e })

	handlerErr := errors.New("final failure")
	h := hooks.WithHooks(r, func(j *job.Job) error { return handlerErr })

	j := newExhaustedJob()
	h(j) //nolint:errcheck

	if emitted != hooks.EventDeadLetter {
		t.Errorf("expected EventDeadLetter, got %q", emitted)
	}
}

func TestWithHooks_ReturnsHandlerError(t *testing.T) {
	r := hooks.New()
	want := errors.New("boom")
	h := hooks.WithHooks(r, func(j *job.Job) error { return want })

	got := h(newJob())
	if !errors.Is(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}
