// Package batch provides utilities for enqueuing multiple jobs at once
// with optional rate limiting and error aggregation.
package batch

import (
	"errors"
	"fmt"

	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
)

// Enqueuer is the interface satisfied by queue.Queue.
type Enqueuer interface {
	Enqueue(j *job.Job) error
}

// Result holds the outcome of a single enqueue attempt within a batch.
type Result struct {
	Job *job.Job
	Err error
}

// Summary is returned after a batch enqueue operation.
type Summary struct {
	Total    int
	Enqueued int
	Failed   int
	Errors   []Result
}

// HasErrors reports whether any jobs failed to enqueue.
func (s Summary) HasErrors() bool {
	return s.Failed > 0
}

// Enqueue submits all provided jobs to the queue, collecting any errors.
// It never stops early — all jobs are attempted regardless of failures.
func Enqueue(q Enqueuer, jobs []*job.Job) Summary {
	s := Summary{Total: len(jobs)}
	for _, j := range jobs {
		if err := q.Enqueue(j); err != nil {
			s.Failed++
			s.Errors = append(s.Errors, Result{Job: j, Err: err})
		} else {
			s.Enqueued++
		}
	}
	return s
}

// EnqueueStrict submits all jobs and returns a combined error if any failed.
func EnqueueStrict(q Enqueuer, jobs []*job.Job) error {
	s := Enqueue(q, jobs)
	if !s.HasErrors() {
		return nil
	}
	errs := make([]error, len(s.Errors))
	for i, r := range s.Errors {
		errs[i] = fmt.Errorf("job %s: %w", r.Job.ID, r.Err)
	}
	return errors.Join(errs...)
}

// FromPayloads is a convenience function that constructs jobs from a slice of
// payloads using the provided factory function, then enqueues them all.
func FromPayloads(q Enqueuer, payloads [][]byte, maxAttempts int) Summary {
	jobs := make([]*job.Job, len(payloads))
	for i, p := range payloads {
		jobs[i] = job.New(p, maxAttempts)
	}
	return Enqueue(q, jobs)
}

// ensure queue.Queue satisfies Enqueuer at compile time.
var _ Enqueuer = (*queue.Queue)(nil)
