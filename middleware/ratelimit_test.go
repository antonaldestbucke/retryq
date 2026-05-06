package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/middleware"
	"github.com/example/retryq/ratelimit"
)

func TestWithRateLimit_AllowsWhenTokensAvailable(t *testing.T) {
	l := ratelimit.New(100, 5)
	called := false
	h := middleware.Chain(
		func(_ context.Context, _ *job.Job) error {
			called = true
			return nil
		},
		middleware.WithRateLimit(l),
	)

	j := newJob("rate-allow", 3)
	if err := h(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestWithRateLimit_DeniesWhenExhausted(t *testing.T) {
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	// burst=1, rate=0 so tokens never refill
	l := ratelimit.New(0, 1, ratelimit.WithClock(clock))
	l.Allow() // drain the single token

	called := false
	h := middleware.Chain(
		func(_ context.Context, _ *job.Job) error {
			called = true
			return nil
		},
		middleware.WithRateLimit(l),
	)

	j := newJob("rate-deny", 3)
	err := h(context.Background(), j)
	if !errors.Is(err, middleware.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
	if called {
		t.Fatal("handler should not have been called")
	}
}

func TestWithRateLimit_DoesNotSwallowHandlerError(t *testing.T) {
	l := ratelimit.New(100, 10)
	sentinel := errors.New("handler error")
	h := middleware.Chain(
		func(_ context.Context, _ *job.Job) error { return sentinel },
		middleware.WithRateLimit(l),
	)

	j := newJob("rate-err", 3)
	if err := h(context.Background(), j); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
