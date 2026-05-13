package pause_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/retryq/job"
	"github.com/your-org/retryq/pause"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func TestWithPause_PassesThroughWhenNotPaused(t *testing.T) {
	g := pause.New()
	j := newJob(t)

	called := false
	h := pause.WithPause(g, func(_ context.Context, _ *job.Job) error {
		called = true
		return nil
	})

	if err := h(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestWithPause_BlocksWhilePaused(t *testing.T) {
	g := pause.New()
	g.Pause()
	j := newJob(t)

	var called bool
	h := pause.WithPause(g, func(_ context.Context, _ *job.Job) error {
		called = true
		return nil
	})

	go func() {
		time.Sleep(20 * time.Millisecond)
		g.Resume()
	}()

	if err := h(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called after resume")
	}
}

func TestWithPause_ReturnsContextErrorWhenCancelled(t *testing.T) {
	g := pause.New()
	g.Pause()
	j := newJob(t)

	h := pause.WithPause(g, func(_ context.Context, _ *job.Job) error {
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	err := h(ctx, j)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestWithPause_PropagatesHandlerError(t *testing.T) {
	g := pause.New()
	j := newJob(t)
	sentinel := errors.New("handler error")

	h := pause.WithPause(g, func(_ context.Context, _ *job.Job) error {
		return sentinel
	})

	err := h(context.Background(), j)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
