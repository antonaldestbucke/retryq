package timeout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/timeout"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func TestWithTimeout_CompletesBeforeDeadline(t *testing.T) {
	h := timeout.WithTimeout(100*time.Millisecond, func(ctx context.Context, j *job.Job) error {
		return nil
	})

	if err := h(context.Background(), newJob(t)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWithTimeout_ReturnsHandlerError(t *testing.T) {
	sentinel := errors.New("handler error")
	h := timeout.WithTimeout(100*time.Millisecond, func(ctx context.Context, j *job.Job) error {
		return sentinel
	})

	err := h(context.Background(), newJob(t))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWithTimeout_TimesOut(t *testing.T) {
	h := timeout.WithTimeout(20*time.Millisecond, func(ctx context.Context, j *job.Job) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	err := h(context.Background(), newJob(t))
	if !timeout.IsTimeout(err) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestWithTimeout_RespectsParentCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h := timeout.WithTimeout(100*time.Millisecond, func(ctx context.Context, j *job.Job) error {
		<-ctx.Done()
		return ctx.Err()
	})

	err := h(ctx, newJob(t))
	if err == nil {
		t.Fatal("expected an error from cancelled context")
	}
}

func TestIsTimeout_TrueForErrTimeout(t *testing.T) {
	if !timeout.IsTimeout(timeout.ErrTimeout) {
		t.Fatal("IsTimeout should return true for ErrTimeout")
	}
}

func TestIsTimeout_FalseForOtherErrors(t *testing.T) {
	if timeout.IsTimeout(errors.New("other")) {
		t.Fatal("IsTimeout should return false for unrelated errors")
	}
}
