// Package priority implements a thread-safe, heap-backed priority queue
// for retryq jobs.
//
// Jobs are enqueued with an explicit priority Level (Low, Normal, or High).
// Dequeue always returns the highest-priority job available, making it
// suitable for use cases where certain job types (e.g. payment processing)
// must be handled before background tasks.
//
// Example usage:
//
//	q := priority.New()
//	q.Enqueue(paymentJob, priority.High)
//	q.Enqueue(reportJob,  priority.Low)
//
//	next := q.Dequeue() // returns paymentJob
package priority
