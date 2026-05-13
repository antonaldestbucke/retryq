// Package pause provides a pausable gate that can temporarily halt job
// processing without stopping the worker entirely.
package pause

import (
	"context"
	"sync"

	"github.com/your-org/retryq/job"
)

// Handler is a function that processes a job.
type Handler func(ctx context.Context, j *job.Job) error

// Gate controls whether job processing is allowed to proceed.
type Gate struct {
	mu     sync.RWMutex
	paused bool
	cond   *sync.Cond
}

// New creates a new Gate in the running (unpaused) state.
func New() *Gate {
	g := &Gate{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// Pause halts processing. Calls to Wait will block until Resume is called.
func (g *Gate) Pause() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = true
}

// Resume allows processing to continue and unblocks any waiting goroutines.
func (g *Gate) Resume() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = false
	g.cond.Broadcast()
}

// IsPaused reports whether the gate is currently paused.
func (g *Gate) IsPaused() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.paused
}

// Wait blocks until the gate is not paused or ctx is cancelled.
// Returns ctx.Err() if the context is cancelled while waiting.
func (g *Gate) Wait(ctx context.Context) error {
	// Fast path: not paused.
	g.mu.RLock()
	if !g.paused {
		g.mu.RUnlock()
		return nil
	}
	g.mu.RUnlock()

	// Slow path: wait for resume or context cancellation.
	done := make(chan struct{})
	go func() {
		g.mu.Lock()
		for g.paused {
			g.cond.Wait()
		}
		g.mu.Unlock()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Wake the waiting goroutine so it can exit cleanly.
		g.cond.Broadcast()
		return ctx.Err()
	}
}
