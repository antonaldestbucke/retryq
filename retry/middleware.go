package retry

import (
	"context"
	"time"

	"github.com/example/retryq/job"
)

// BackoffFunc computes the delay before the next retry attempt.
// attempt is zero-indexed (0 = first failure).
type BackoffFunc func(attempt int) time.Duration

// Handler is the signature for a job processing function.
type Handler func(ctx context.Context, j *job.Job) error

// WithBackoff wraps handler with a retry loop driven by backoff.
// On each failure the job's MarkFailed is called and, if the job is not yet
// exhausted, execution sleeps for backoff(attempt) before retrying.
// If the context is cancelled the loop exits immediately.
//
// This middleware is intended for in-process retries within a single
// dequeue cycle. For queue-level retries (requeue after backoff) prefer
// the worker/queue requeue path instead.
func WithBackoff(backoff BackoffFunc, next Handler) Handler {
	return func(ctx context.Context, j *job.Job) error {
		var attempt int
		for {
			err := next(ctx, j)
			if err == nil {
				return nil
			}

			j.MarkFailed(err)
			if j.IsExhausted() {
				return err
			}

			delay := backoff(attempt)
			attempt++

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
}
