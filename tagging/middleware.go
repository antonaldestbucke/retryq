package tagging

import (
	"context"

	"github.com/example/retryq/job"
)

// Handler is the standard job-processing function signature.
type Handler func(ctx context.Context, j *job.Job) error

// WithTagFilter returns middleware that skips jobs not carrying ALL of the
// required tags. Skipped jobs are acknowledged as successful so they are not
// requeued; the caller should route them to an appropriate queue instead.
func WithTagFilter(store *Store, required ...string) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			if !store.Has(j, required...) {
				return nil
			}
			return next(ctx, j)
		}
	}
}

// WithAutoTag returns middleware that automatically applies a fixed set of tags
// to every job that passes through, then delegates to the next handler.
// Tags are added (not replaced) so pre-existing tags are preserved.
func WithAutoTag(store *Store, tags ...string) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(ctx context.Context, j *job.Job) error {
			store.Add(j, tags...)
			return next(ctx, j)
		}
	}
}
