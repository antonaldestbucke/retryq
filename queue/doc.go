// Package queue provides a thread-safe in-memory job queue with dead-letter
// support for the retryq system.
//
// Jobs are enqueued and dequeued in FIFO order. When a failed job is requeued
// via [Queue.Requeue], the queue inspects whether the job has exhausted its
// retry attempts (via [job.Job.IsExhausted]). Exhausted jobs are automatically
// routed to an internal dead-letter list instead of being re-added to the
// pending pool.
//
// Basic usage:
//
//	q := queue.New()
//
//	j := job.New("id-1", "email", payload, 5)
//	q.Enqueue(j)
//
//	next, _ := q.Dequeue()
//	// process next ...
//	next.MarkFailed()
//	q.Requeue(next) // re-added to pending or moved to dead-letter
//
// The queue is safe for concurrent use by multiple goroutines.
package queue
