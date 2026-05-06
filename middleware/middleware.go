// Package middleware provides hook-based middleware for job processing in retryq.
// Middleware functions wrap the job handler to add cross-cutting concerns such as
// logging, metrics tracking, and panic recovery.
package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/metrics"
)

// Handler is a function that processes a job.
type Handler func(ctx context.Context, j *job.Job) error

// Middleware wraps a Handler to add additional behavior.
type Middleware func(next Handler) Handler

// Chain applies a list of middlewares to a handler in order.
func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// WithLogging returns a Middleware that logs job start, completion, and errors.
func WithLogging(logger *log.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			start := time.Now()
			logger.Printf("[retryq] starting job id=%s type=%s attempt=%d", j.ID, j.Type, j.Attempts)
			err := next(ctx, j)
			elapsed := time.Since(start)
			if err != nil {
				logger.Printf("[retryq] job failed id=%s type=%s elapsed=%s err=%v", j.ID, j.Type, elapsed, err)
			} else {
				logger.Printf("[retryq] job completed id=%s type=%s elapsed=%s", j.ID, j.Type, elapsed)
			}
			return err
		}
	}
}

// WithMetrics returns a Middleware that records job metrics.
func WithMetrics(m *metrics.Metrics) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			err := next(ctx, j)
			if err != nil {
				m.IncFailed()
			} else {
				m.IncProcessed()
			}
			return err
		}
	}
}

// WithRecovery returns a Middleware that recovers from panics and converts them to errors.
func WithRecovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered in job id=%s: %v", j.ID, r)
				}
			}()
			return next(ctx, j)
		}
	}
}
