package dedup

import (
	"errors"

	"github.com/example/retryq/job"
)

// ErrDuplicate is returned when a job is identified as a duplicate.
var ErrDuplicate = errors.New("dedup: duplicate job")

// Handler is the signature for a job handler function.
type Handler func(*job.Job) error

// WithDedup wraps a Handler and skips processing if the job is a duplicate.
// On successful completion the key is removed so the job can be re-enqueued later.
func WithDedup(store *Store, next Handler) Handler {
	return func(j *job.Job) error {
		if store.IsDuplicate(j) {
			return ErrDuplicate
		}
		err := next(j)
		if err == nil {
			// Free the slot so the same logical job can run again after success.
			store.Remove(j.ID)
		}
		return err
	}
}
