package hooks_test

import (
	"testing"
	"time"

	"github.com/example/retryq/hooks"
	"github.com/example/retryq/job"
)

func newJob() *job.Job {
	return job.New("test-type", []byte(`{"key":"value"}`), 3, time.Second)
}

func TestNew_EmptyRegistry(t *testing.T) {
	r := hooks.New()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	// Emit on empty registry should not panic
	r.Emit(hooks.EventEnqueued, newJob())
}

func TestOn_And_Emit(t *testing.T) {
	r := hooks.New()
	var called []hooks.EventType

	r.On(hooks.EventEnqueued, func(e hooks.EventType, j *job.Job) {
		called = append(called, e)
	})
	r.On(hooks.EventSucceeded, func(e hooks.EventType, j *job.Job) {
		called = append(called, e)
	})

	r.Emit(hooks.EventEnqueued, newJob())
	r.Emit(hooks.EventSucceeded, newJob())

	if len(called) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(called))
	}
	if called[0] != hooks.EventEnqueued {
		t.Errorf("expected EventEnqueued, got %s", called[0])
	}
	if called[1] != hooks.EventSucceeded {
		t.Errorf("expected EventSucceeded, got %s", called[1])
	}
}

func TestOn_MultipleHandlers_SameEvent(t *testing.T) {
	r := hooks.New()
	count := 0

	r.On(hooks.EventFailed, func(e hooks.EventType, j *job.Job) { count++ })
	r.On(hooks.EventFailed, func(e hooks.EventType, j *job.Job) { count++ })
	r.On(hooks.EventFailed, func(e hooks.EventType, j *job.Job) { count++ })

	r.Emit(hooks.EventFailed, newJob())

	if count != 3 {
		t.Errorf("expected 3 handler calls, got %d", count)
	}
}

func TestEmit_PassesJobToHandler(t *testing.T) {
	r := hooks.New()
	j := newJob()
	var received *job.Job

	r.On(hooks.EventDeadLetter, func(e hooks.EventType, jj *job.Job) {
		received = jj
	})
	r.Emit(hooks.EventDeadLetter, j)

	if received != j {
		t.Error("handler did not receive the expected job")
	}
}

func TestClear_RemovesHandlers(t *testing.T) {
	r := hooks.New()
	called := false

	r.On(hooks.EventEnqueued, func(e hooks.EventType, j *job.Job) { called = true })
	r.Clear(hooks.EventEnqueued)
	r.Emit(hooks.EventEnqueued, newJob())

	if called {
		t.Error("handler should not have been called after Clear")
	}
}
