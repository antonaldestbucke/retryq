package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dstotijn/retryq/middleware"
	"github.com/dstotijn/retryq/tracing"
)

func TestWithTracing_RecordsSuccessSpan(t *testing.T) {
	tr := tracing.New()
	chain := middleware.Chain(
		tracing.WithTracing(tr),
	)

	handler := chain(func(_ context.Context, j *job.Job) error {
		return nil
	})

	if err := handler(context.Background(), newJob()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := len(tr.Spans()); n != 1 {
		t.Fatalf("expected 1 span, got %d", n)
	}
}

func TestWithTracing_RecordsErrorSpan(t *testing.T) {
	sentinel := errors.New("boom")
	tr := tracing.New()
	chain := middleware.Chain(tracing.WithTracing(tr))

	handler := chain(func(_ context.Context, j *job.Job) error {
		return sentinel
	})

	err := handler(context.Background(), newJob())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !errors.Is(spans[0].Err, sentinel) {
		t.Errorf("expected span to carry error")
	}
}

func TestWithTracing_MultipleJobs(t *testing.T) {
	tr := tracing.New()
	chain := middleware.Chain(tracing.WithTracing(tr))
	handler := chain(func(_ context.Context, j *job.Job) error { return nil })

	for i := 0; i < 5; i++ {
		_ = handler(context.Background(), newJob())
	}
	if n := len(tr.Spans()); n != 5 {
		t.Fatalf("expected 5 spans, got %d", n)
	}
}
