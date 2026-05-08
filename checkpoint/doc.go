// Package checkpoint provides lightweight job progress tracking for retryq.
//
// Long-running jobs can persist intermediate state at each meaningful step.
// If the job fails and is retried, the handler can resume from the last saved
// checkpoint instead of restarting from scratch, reducing redundant work.
//
// Basic usage:
//
//	store := checkpoint.New()
//
//	// Inside a job handler:
//	record, err := store.Load(job.ID)
//	startStep := 0
//	if err == nil {
//		startStep = record.Step
//	}
//
//	for step := startStep; step < totalSteps; step++ {
//		// ... do work ...
//		store.Save(job.ID, step+1, map[string]string{"cursor": cursor})
//	}
//
//	// On success, clean up the checkpoint.
//	store.Delete(job.ID)
package checkpoint
