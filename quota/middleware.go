package quota

import (
	"errors"

	"github.com/example/retryq/job"
)

// ErrQuotaExceeded is returned when a job's key has exhausted its quota.
var ErrQuotaExceeded = errors.New("quota: limit exceeded for key")

// Handler is a function that processes a job.
type Handler func(*job.Job) error

// KeyFunc extracts a quota key from a job.
type KeyFunc func(*job.Job) string

// WithQuota returns a middleware that enforces per-key processing limits.
// If the key derived from the job has exhausted its quota the job is not
// forwarded to next and ErrQuotaExceeded is returned instead.
func WithQuota(q *Quota, key KeyFunc) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(j *job.Job) error {
			k := key(j)
			if !q.Allow(k) {
				return ErrQuotaExceeded
			}
			return next(j)
		}
	}
}
