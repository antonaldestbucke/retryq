// Package dedup implements job deduplication for retryq.
//
// It prevents the same logical job from being processed concurrently or
// re-enqueued within a configurable time window. Jobs are identified by
// their ID field.
//
// Basic usage:
//
//	store := dedup.New(dedup.WithWindow(10 * time.Minute))
//
//	handler := dedup.WithDedup(store, func(j *job.Job) error {
//		// process job
//		return nil
//	})
//
// The dedup window begins when the job is first seen. On successful
// completion the key is released, allowing the same job ID to be
// processed again in the future. On failure the key is retained so
// that a racing duplicate is still rejected until the window expires
// or the key is explicitly removed via Store.Remove.
package dedup
