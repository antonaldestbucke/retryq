package middleware_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"

	"github.com/example/retryq/job"
	"github.com/example/retryq/metrics"
	"github.com/example/retryq/middleware"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{"key":"value"}`))
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	return j
}

func TestChain_ExecutesInOrder(t *testing.T) {
	order := []string{}
	mw1 := func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, j *job.Job) error {
			order = append(order, "mw1-before")
			err := next(ctx, j)
			order = append(order, "mw1-after")
			return err
		}
	}
	mw2 := func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, j *job.Job) error {
			order = append(order, "mw2-before")
			err := next(ctx, j)
			order = append(order, "mw2-after")
			return err
		}
	}
	base := func(ctx context.Context, j *job.Job) error {
		order = append(order, "handler")
		return nil
	}
	h := middleware.Chain(base, mw1, mw2)
	if err := h(context.Background(), newJob(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("step %d: got %q, want %q", i, order[i], v)
		}
	}
}

func TestWithLogging_LogsOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	base := func(ctx context.Context, j *job.Job) error { return nil }
	h := middleware.Chain(base, middleware.WithLogging(logger))
	if err := h(context.Background(), newJob(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected log output, got none")
	}
}

func TestWithLogging_LogsOnError(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	base := func(ctx context.Context, j *job.Job) error { return errors.New("boom") }
	h := middleware.Chain(base, middleware.WithLogging(logger))
	_ = h(context.Background(), newJob(t))
	if buf.Len() == 0 {
		t.Error("expected log output, got none")
	}
}

func TestWithMetrics_IncrementsProcessed(t *testing.T) {
	m := metrics.New()
	base := func(ctx context.Context, j *job.Job) error { return nil }
	h := middleware.Chain(base, middleware.WithMetrics(m))
	_ = h(context.Background(), newJob(t))
	if snap := m.Snapshot(); snap.Processed != 1 {
		t.Errorf("expected Processed=1, got %d", snap.Processed)
	}
}

func TestWithMetrics_IncrementsFailed(t *testing.T) {
	m := metrics.New()
	base := func(ctx context.Context, j *job.Job) error { return errors.New("fail") }
	h := middleware.Chain(base, middleware.WithMetrics(m))
	_ = h(context.Background(), newJob(t))
	if snap := m.Snapshot(); snap.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", snap.Failed)
	}
}

func TestWithRecovery_CatchesPanic(t *testing.T) {
	base := func(ctx context.Context, j *job.Job) error {
		panic("unexpected failure")
	}
	h := middleware.Chain(base, middleware.WithRecovery())
	err := h(context.Background(), newJob(t))
	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}
