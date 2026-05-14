// Package replay provides utilities for recovering dead-lettered jobs.
//
// When a job exhausts its retry budget it is moved to a DeadLetterStore.
// The replay package lets operators inspect those jobs and re-enqueue them
// for another round of processing — optionally resetting the attempt counter
// so the job receives a full retry budget.
//
// # Basic usage
//
//	dead := storage.NewDeadLetterStore()
//	q    := queue.New()
//
//	r := replay.New(dead, q, replay.WithResetAttempts())
//
//	// Re-enqueue all dead-lettered jobs.
//	n, err := r.ReplayAll(ctx)
//
//	// Or replay a single job by ID.
//	err = r.ReplayOne(ctx, jobID)
//
// # Middleware
//
// WithDeadLetterReplay wraps a handler so that exhausted jobs are
// automatically stored in the dead-letter store rather than returned as
// errors, keeping the worker loop clean.
//
//	handler = replay.WithDeadLetterReplay(dead)(handler)
package replay
