package tracing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dstotijn/retryq/job"
	"github.com/dstotijn/retryq/tracing"
)

func newJob() *job.Job {
	j, _ := job.New("test-queue", []byte(`{}`))
	return j
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_EmptySpans(t *testing.T) {
	tr := tracing.New()
	if got := tr.Spans(); len(got) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(got))
	}
}

func TestRecord_SuccessSpan(t *testing.T) {
	now := time.Now()
	tr := tracing.New(tracing.WithClock(fixedClock(now)))
	j := newJob()

	err := tr.Record(context.Background(), j, func(_ context.Context, _ *job.Job) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Err != nil {
		t.Errorf("expected nil err in span, got %v", spans[0].Err)
	}
	if spans[0].JobID != j.ID {
		t.Errorf("expected job ID %s, got %s", j.ID, spans[0].JobID)
	}
}

func TestRecord_ErrorSpan(t *testing.T) {
	sentinel := errors.New("handler error")
	tr := tracing.New()
	j := newJob()

	err := tr.Record(context.Background(), j, func(_ context.Context, _ *job.Job) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !errors.Is(spans[0].Err, sentinel) {
		t.Errorf("expected sentinel in span err, got %v", spans[0].Err)
	}
}

func TestRecord_SpanDuration(t *testing.T) {
	base := time.Now()
	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(50 * time.Millisecond)
	}
	tr := tracing.New(tracing.WithClock(clock))
	_ = tr.Record(context.Background(), newJob(), func(_ context.Context, _ *job.Job) error { return nil })

	spans := tr.Spans()
	if d := spans[0].Duration(); d != 50*time.Millisecond {
		t.Errorf("expected 50ms duration, got %v", d)
	}
}

func TestSpans_IsCopy(t *testing.T) {
	tr := tracing.New()
	_ = tr.Record(context.Background(), newJob(), func(_ context.Context, _ *job.Job) error { return nil })

	a := tr.Spans()
	a[0].Queue = "mutated"
	b := tr.Spans()
	if b[0].Queue == "mutated" {
		t.Error("Spans() should return an independent copy")
	}
}

func TestClear_RemovesSpans(t *testing.T) {
	tr := tracing.New()
	_ = tr.Record(context.Background(), newJob(), func(_ context.Context, _ *job.Job) error { return nil })
	tr.Clear()
	if got := tr.Spans(); len(got) != 0 {
		t.Fatalf("expected 0 spans after Clear, got %d", len(got))
	}
}
