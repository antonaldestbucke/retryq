package circuitbreaker

import (
	"context"

	"github.com/example/retryq/job"
	"github.com/example/retryq/worker"
)

// WithCircuitBreaker returns a middleware that gates job execution behind the
// given CircuitBreaker. When the circuit is open the handler is skipped and
// ErrCircuitOpen is returned so the worker can requeue or dead-letter the job.
func WithCircuitBreaker(cb *CircuitBreaker) worker.Middleware {
	return func(next worker.HandlerFunc) worker.HandlerFunc {
		return func(ctx context.Context, j *job.Job) error {
			if err := cb.Allow(); err != nil {
				return err
			}

			err := next(ctx, j)
			if err != nil {
				cb.RecordFailure()
				return err
			}

			cb.RecordSuccess()
			return nil
		}
	}
}
