package replay

import (
	"context"
	"fmt"

	"github.com/example/retryq/job"
	"github.com/example/retryq/storage"
)

// Handler is a function that processes a job.
type Handler func(ctx context.Context, j *job.Job) error

// WithDeadLetterReplay returns a middleware that, upon detecting an exhausted
// job returned from the inner handler, stores it in the dead-letter store
// instead of propagating the error. Non-exhausted failures pass through
// unchanged so the worker can requeue them normally.
func WithDeadLetterReplay(dead *storage.DeadLetterStore) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			err := next(ctx, j)
			if err == nil {
				return nil
			}
			if j.IsExhausted() {
				if dlErr := dead.Add(j, err.Error()); dlErr != nil {
					return fmt.Errorf("dead-letter store: %w", dlErr)
				}
				// Return nil so the worker does not attempt to requeue.
				return nil
			}
			return err
		}
	}
}
