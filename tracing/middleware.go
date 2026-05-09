package tracing

import (
	"context"

	"github.com/dstotijn/retryq/job"
	"github.com/dstotijn/retryq/middleware"
)

// WithTracing returns a middleware that records a span for every job execution.
func WithTracing(t *Tracer) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, j *job.Job) error {
			return t.Record(ctx, j, next)
		}
	}
}
