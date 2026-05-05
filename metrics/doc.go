// Package metrics exposes lightweight atomic counters that track the
// lifecycle of jobs flowing through retryq.
//
// Usage:
//
//	m := metrics.New()
//
//	// Increment counters as jobs move through the system.
//	m.Enqueued.Add(1)
//	m.Completed.Add(1)
//
//	// Obtain a consistent point-in-time snapshot for reporting.
//	snap := m.Snapshot()
//	fmt.Printf("completed=%d failed=%d dead=%d\n",
//		snap.Completed, snap.Failed, snap.DeadLettered)
//
// All counters are safe for concurrent use without additional locking.
package metrics
