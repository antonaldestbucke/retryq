// Package fanout provides a mechanism to dispatch a single job to multiple
// queues simultaneously, useful for broadcasting work across independent consumers.
package fanout

import (
	"fmt"

	"github.com/example/retryq/job"
)

// Enqueuer is the interface satisfied by any queue that can accept a job.
type Enqueuer interface {
	Enqueue(j *job.Job) error
}

// Fanout dispatches jobs to a fixed set of queues.
type Fanout struct {
	queues []Enqueuer
}

// New returns a Fanout that will broadcast to the provided queues.
// At least one queue must be supplied.
func New(queues ...Enqueuer) (*Fanout, error) {
	if len(queues) == 0 {
		return nil, fmt.Errorf("fanout: at least one queue is required")
	}
	return &Fanout{queues: queues}, nil
}

// Dispatch enqueues j into every registered queue.
// It continues past individual failures and returns a combined error
// listing every queue index that rejected the job.
func (f *Fanout) Dispatch(j *job.Job) error {
	var errs []error
	for i, q := range f.queues {
		if err := q.Enqueue(j); err != nil {
			errs = append(errs, fmt.Errorf("fanout: queue[%d]: %w", i, err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return &DispatchError{Errs: errs}
}

// DispatchError aggregates one or more enqueue failures.
type DispatchError struct {
	Errs []error
}

func (e *DispatchError) Error() string {
	return fmt.Sprintf("fanout: %d queue(s) failed to accept job", len(e.Errs))
}

// Len returns the number of queues registered with this Fanout.
func (f *Fanout) Len() int { return len(f.queues) }
