// Package drain provides a graceful shutdown mechanism that waits for
// in-flight jobs to complete before stopping the worker pool.
package drain

import (
	"context"
	"sync"
	"time"
)

// Drainer tracks in-flight jobs and provides a way to wait for all of them
// to finish, optionally bounded by a deadline.
type Drainer struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	closed  bool
	inflight int
}

// New returns a new Drainer.
func New() *Drainer {
	return &Drainer{}
}

// Acquire signals that a new job has started. It returns false if the Drainer
// has already been closed, meaning no new jobs should be accepted.
func (d *Drainer) Acquire() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return false
	}
	d.inflight++
	d.wg.Add(1)
	return true
}

// Release signals that a job has finished.
func (d *Drainer) Release() {
	d.mu.Lock()
	d.inflight--
	d.mu.Unlock()
	d.wg.Done()
}

// Inflight returns the number of currently running jobs.
func (d *Drainer) Inflight() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inflight
}

// Close marks the Drainer as closed so no new jobs can be acquired.
func (d *Drainer) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
}

// Wait blocks until all in-flight jobs complete or the context is cancelled.
// It returns context.DeadlineExceeded or context.Canceled if the context
// expires before all jobs finish.
func (d *Drainer) Wait(ctx context.Context) error {
	d.Close()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WaitWithTimeout is a convenience wrapper around Wait that accepts a
// duration instead of a context.
func (d *Drainer) WaitWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.Wait(ctx)
}
