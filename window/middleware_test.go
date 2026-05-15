package window_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/retryq/job"
	"github.com/user/retryq/window"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func noop(_ context.Context, _ *job.Job) error { return nil }

func TestWithWindow_AllowsUnderLimit(t *testing.T) {
	c := window.New(time.Second)
	h := window.WithWindow(c, 3, noop)

	for i := 0; i < 3; i++ {
		if err := h(context.Background(), newJob(t)); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, err)
		}
	}
}

func TestWithWindow_RejectsOverLimit(t *testing.T) {
	c := window.New(time.Second)
	h := window.WithWindow(c, 2, noop)

	// consume the budget
	_ = h(context.Background(), newJob(t))
	_ = h(context.Background(), newJob(t))

	err := h(context.Background(), newJob(t))
	if err != window.ErrWindowExceeded {
		t.Fatalf("expected ErrWindowExceeded, got %v", err)
	}
}

func TestWithWindow_AllowsAfterWindowRollover(t *testing.T) {
	now := time.Unix(2_000, 0)
	clock := func() time.Time { return now }

	c := window.New(time.Second, window.WithClock(clock))
	h := window.WithWindow(c, 1, noop)

	// consume the single-event budget
	_ = h(context.Background(), newJob(t))

	// roll the window forward
	now = now.Add(2 * time.Second)

	if err := h(context.Background(), newJob(t)); err != nil {
		t.Fatalf("expected nil after window rollover, got %v", err)
	}
}

func TestWithWindow_PropagatesHandlerError(t *testing.T) {
	c := window.New(time.Second)
	sentinel := context.DeadlineExceeded
	h := window.WithWindow(c, 10, func(_ context.Context, _ *job.Job) error {
		return sentinel
	})

	if err := h(context.Background(), newJob(t)); err != sentinel {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
