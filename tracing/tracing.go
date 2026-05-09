// Package tracing provides job execution tracing with span recording.
package tracing

import (
	"context"
	"sync"
	"time"

	"github.com/dstotijn/retryq/job"
)

// Span represents a single traced job execution.
type Span struct {
	JobID     string
	Queue     string
	Attempt   int
	StartedAt time.Time
	EndedAt   time.Time
	Err       error
}

// Duration returns the elapsed time of the span.
func (s Span) Duration() time.Duration {
	return s.EndedAt.Sub(s.StartedAt)
}

// Tracer records spans for job executions.
type Tracer struct {
	mu    sync.Mutex
	spans []Span
	clock func() time.Time
}

// Option configures a Tracer.
type Option func(*Tracer)

// WithClock sets a custom clock function (useful for testing).
func WithClock(fn func() time.Time) Option {
	return func(t *Tracer) { t.clock = fn }
}

// New creates a new Tracer.
func New(opts ...Option) *Tracer {
	t := &Tracer{clock: time.Now}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Record starts and ends a span around the given handler call.
func (t *Tracer) Record(ctx context.Context, j *job.Job, handler func(context.Context, *job.Job) error) error {
	start := t.clock()
	err := handler(ctx, j)
	end := t.clock()

	t.mu.Lock()
	t.spans = append(t.spans, Span{
		JobID:     j.ID,
		Queue:     j.Queue,
		Attempt:   j.Attempts,
		StartedAt: start,
		EndedAt:   end,
		Err:       err,
	})
	t.mu.Unlock()

	return err
}

// Spans returns a copy of all recorded spans.
func (t *Tracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Span, len(t.spans))
	copy(out, t.spans)
	return out
}

// Clear removes all recorded spans.
func (t *Tracer) Clear() {
	t.mu.Lock()
	t.spans = t.spans[:0]
	t.mu.Unlock()
}
