// Package replay provides utilities for replaying dead-lettered jobs
// back into a queue for reprocessing.
package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/storage"
)

// Enqueuer is the interface required to re-enqueue a job.
type Enqueuer interface {
	Enqueue(j *job.Job) error
}

// Replayer reads jobs from a DeadLetterStore and re-enqueues them.
type Replayer struct {
	dead    *storage.DeadLetterStore
	queue   Enqueuer
	resetAt func(*job.Job)
}

// Option configures a Replayer.
type Option func(*Replayer)

// WithResetAttempts resets a job's attempt counter before re-enqueueing,
// giving it a full retry budget again.
func WithResetAttempts() Option {
	return func(r *Replayer) {
		r.resetAt = func(j *job.Job) {
			j.Attempts = 0
			j.LastError = ""
			j.RetryAt = time.Time{}
		}
	}
}

// New creates a new Replayer that moves jobs from the dead-letter store
// into the provided queue.
func New(dead *storage.DeadLetterStore, queue Enqueuer, opts ...Option) *Replayer {
	r := &Replayer{
		dead:  dead,
		queue: queue,
		resetAt: func(_ *job.Job) {}, // no-op by default
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// ReplayAll re-enqueues every job currently in the dead-letter store.
// It returns the number of jobs successfully replayed and any accumulated errors.
func (r *Replayer) ReplayAll(ctx context.Context) (int, error) {
	records, err := r.dead.List()
	if err != nil {
		return 0, fmt.Errorf("replay: list dead-letter jobs: %w", err)
	}

	var errs []error
	replayed := 0
	for _, rec := range records {
		if ctx.Err() != nil {
			break
		}
		r.resetAt(rec.Job)
		if enqErr := r.queue.Enqueue(rec.Job); enqErr != nil {
			errs = append(errs, fmt.Errorf("replay: enqueue %s: %w", rec.Job.ID, enqErr))
			continue
		}
		if delErr := r.dead.Remove(rec.Job.ID); delErr != nil {
			errs = append(errs, fmt.Errorf("replay: remove %s from dead-letter: %w", rec.Job.ID, delErr))
		}
		replayed++
	}

	if len(errs) > 0 {
		return replayed, fmt.Errorf("replay: %d error(s): %v", len(errs), errs)
	}
	return replayed, nil
}

// ReplayOne re-enqueues a single job by ID from the dead-letter store.
func (r *Replayer) ReplayOne(ctx context.Context, id string) error {
	rec, err := r.dead.Get(id)
	if err != nil {
		return fmt.Errorf("replay: get %s: %w", id, err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	r.resetAt(rec.Job)
	if err := r.queue.Enqueue(rec.Job); err != nil {
		return fmt.Errorf("replay: enqueue %s: %w", id, err)
	}
	if err := r.dead.Remove(id); err != nil {
		return fmt.Errorf("replay: remove %s from dead-letter: %w", id, err)
	}
	return nil
}
