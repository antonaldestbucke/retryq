package middleware

import (
	"context"
	"errors"

	"github.com/example/retryq/job"
	"github.com/example/retryq/ratelimit"
)

// ErrRateLimited is returned by the WithRateLimit middleware when the
// rate limiter denies the request.
var ErrRateLimited = errors.New("retryq: rate limit exceeded")

// WithRateLimit returns a Handler middleware that gates job execution
// behind the provided Limiter. If no token is available the job handler
// is not called and ErrRateLimited is returned, causing the worker to
// requeue the job according to its retry policy.
//
//	limiter := ratelimit.New(50, 5)
//	handler := middleware.Chain(myHandler, middleware.WithRateLimit(limiter))
func WithRateLimit(l *ratelimit.Limiter) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			if !l.Allow() {
				return ErrRateLimited
			}
			return next(ctx, j)
		}
	}
}
