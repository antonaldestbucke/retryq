package sampling

import (
	"github.com/yourorg/retryq/job"
)

// Handler is a function that processes a job.
type Handler func(*job.Job) error

// WithSampling returns middleware that probabilistically skips jobs.
// Jobs that are not sampled in have their handler skipped and ErrSkipped
// is returned. The caller may choose to treat ErrSkipped as a no-op.
func WithSampling(s *Sampler) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(j *job.Job) error {
			if !s.Allow(j) {
				return ErrSkipped
			}
			return next(j)
		}
	}
}
