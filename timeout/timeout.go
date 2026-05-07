// Package timeout provides job execution timeout enforcement.
package timeout

import (
	"context"
	"errors"
	"time"

	"github.com/example/retryq/job"
)

// ErrTimeout is returned when a job handler exceeds its allowed duration.
var ErrTimeout = errors.New("job timed out")

// Handler is the signature for a job processing function.
type Handler func(ctx context.Context, j *job.Job) error

// WithTimeout wraps a Handler and cancels execution if it exceeds d.
// If the handler does not respect context cancellation the goroutine may
// linger, but the caller will receive ErrTimeout promptly.
func WithTimeout(d time.Duration, h Handler) Handler {
	return func(ctx context.Context, j *job.Job) error {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		type result struct {
			err error
		}
		ch := make(chan result, 1)

		go func() {
			ch <- result{err: h(ctx, j)}
		}()

		select {
		case r := <-ch:
			return r.err
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return ErrTimeout
			}
			return ctx.Err()
		}
	}
}

// IsTimeout reports whether err is (or wraps) ErrTimeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}
