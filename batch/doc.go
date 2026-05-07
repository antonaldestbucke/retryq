// Package batch provides helpers for submitting multiple jobs to a retryq
// queue in a single operation.
//
// # Batch Enqueue
//
// Use [Enqueue] to submit a slice of jobs, collecting individual errors
// without stopping on the first failure:
//
//	summary := batch.Enqueue(q, jobs)
//	if summary.HasErrors() {
//	    for _, r := range summary.Errors {
//	        log.Printf("failed to enqueue job %s: %v", r.Job.ID, r.Err)
//	    }
//	}
//
// Use [EnqueueStrict] when you want a single combined error:
//
//	if err := batch.EnqueueStrict(q, jobs); err != nil {
//	    return fmt.Errorf("batch enqueue: %w", err)
//	}
//
// Use [FromPayloads] to construct and enqueue jobs from raw byte slices
// in one step.
package batch
