package hooks

import (
	"github.com/example/retryq/job"
	"github.com/example/retryq/worker"
)

// WithHooks wraps a worker.HandlerFunc and emits lifecycle events
// via the provided Registry on success, failure, and dead-letter transitions.
func WithHooks(r *Registry, next worker.HandlerFunc) worker.HandlerFunc {
	return func(j *job.Job) error {
		err := next(j)
		if err == nil {
			r.Emit(EventSucceeded, j)
			return nil
		}

		if j.IsExhausted() {
			r.Emit(EventDeadLetter, j)
		} else {
			r.Emit(EventFailed, j)
		}
		return err
	}
}
