package queue

import (
	"errors"
	"sync"

	"github.com/user/retryq/job"
)

// ErrQueueClosed is returned when operations are attempted on a closed queue.
var ErrQueueClosed = errors.New("queue: queue is closed")

// Queue manages pending jobs and routes exhausted jobs to a dead-letter store.
type Queue struct {
	mu       sync.Mutex
	pending  []*job.Job
	dead     []*job.Job
	closed   bool
}

// New creates and returns an empty Queue.
func New() *Queue {
	return &Queue{}
}

// Enqueue adds a job to the pending list.
func (q *Queue) Enqueue(j *job.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}

	q.pending = append(q.pending, j)
	return nil
}

// Dequeue removes and returns the next pending job, or nil if none exist.
func (q *Queue) Dequeue() (*job.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, ErrQueueClosed
	}

	if len(q.pending) == 0 {
		return nil, nil
	}

	j := q.pending[0]
	q.pending = q.pending[1:]
	return j, nil
}

// Requeue re-adds a failed job. If the job is exhausted it is moved to dead-letter.
func (q *Queue) Requeue(j *job.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}

	if j.IsExhausted() {
		q.dead = append(q.dead, j)
		return nil
	}

	q.pending = append(q.pending, j)
	return nil
}

// PendingCount returns the number of jobs waiting to be processed.
func (q *Queue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// DeadLetterCount returns the number of exhausted jobs.
func (q *Queue) DeadLetterCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.dead)
}

// DeadLetterJobs returns a copy of all dead-letter jobs.
func (q *Queue) DeadLetterJobs() []*job.Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	copy := make([]*job.Job, len(q.dead))
	for i, j := range q.dead {
		copy[i] = j
	}
	return copy
}

// Close marks the queue as closed, preventing further enqueue/dequeue operations.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
}
