package timeout

import (
	"context"
	"time"

	"github.com/example/retryq/job"
)

// Middleware is the standard middleware signature used across retryq.
type Middleware func(Handler) Handler

// WithJobTimeout returns a Middleware that enforces a per-job execution
// deadline of d. It is designed to be composed with middleware.Chain.
//
//	 chain := middleware.Chain(handler,
//	     timeout.WithJobTimeout(5*time.Second),
//	 )
func WithJobTimeout(d time.Duration) Middleware {
	return func(next Handler) Handler {
		return WithTimeout(d, func(ctx context.Context, j *job.Job) error {
			return next(ctx, j)
		})
	}
}
