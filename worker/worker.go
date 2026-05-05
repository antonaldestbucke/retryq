// Package worker provides a concurrent job processor that pulls jobs
// from a queue, executes a handler function, and handles retries and
// dead-letter routing automatically.
package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
)

// HandlerFunc is the function signature for processing a job.
type HandlerFunc func(ctx context.Context, j *job.Job) error

// Worker pulls jobs from a queue and processes them with the provided handler.
type Worker struct {
	q       *queue.Queue
	handler HandlerFunc
	concurrency int
	pollInterval time.Duration
}

// Option is a functional option for configuring a Worker.
type Option func(*Worker)

// WithConcurrency sets the number of concurrent goroutines processing jobs.
func WithConcurrency(n int) Option {
	return func(w *Worker) {
		if n > 0 {
			w.concurrency = n
		}
	}
}

// WithPollInterval sets how often the worker polls for new jobs when idle.
func WithPollInterval(d time.Duration) Option {
	return func(w *Worker) {
		if d > 0 {
			w.pollInterval = d
		}
	}
}

// New creates a Worker that processes jobs from q using handler.
func New(q *queue.Queue, handler HandlerFunc, opts ...Option) *Worker {
	w := &Worker{
		q:            q,
		handler:      handler,
		concurrency:  1,
		pollInterval: 500 * time.Millisecond,
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Run starts the worker and blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, w.concurrency)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		default:
		}

		j, ok := w.q.Dequeue()
		if !ok {
			select {
			case <-ctx.Done():
				wg.Wait()
				return
			case <-time.After(w.pollInterval):
				continue
			}
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(j *job.Job) {
			defer wg.Done()
			defer func() { <-sem }()
			w.process(ctx, j)
		}(j)
	}
}

func (w *Worker) process(ctx context.Context, j *job.Job) {
	if err := w.handler(ctx, j); err != nil {
		log.Printf("job %s failed (attempt %d): %v", j.ID, j.Attempts, err)
		w.q.Requeue(j)
		return
	}
	w.q.Complete(j)
}
