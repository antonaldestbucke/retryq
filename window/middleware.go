package window

import (
	"context"
	"errors"

	"github.com/user/retryq/job"
)

// ErrWindowExceeded is returned when the sliding-window limit is breached.
var ErrWindowExceeded = errors.New("window: rate limit exceeded")

// Handler is the standard job handler signature used across retryq.
type Handler func(ctx context.Context, j *job.Job) error

// WithWindow wraps a Handler and rejects jobs once the counter exceeds
// maxEvents within the configured window. The counter must be created
// separately so it can be shared across multiple workers if desired.
func WithWindow(c *Counter, maxEvents int, next Handler) Handler {
	return func(ctx context.Context, j *job.Job) error {
		if c.Count() >= maxEvents {
			return ErrWindowExceeded
		}
		c.Add(1)
		return next(ctx, j)
	}
}
