// Package metrics provides counters and gauges for monitoring retryq internals.
package metrics

import "sync/atomic"

// Metrics holds runtime counters for queue and worker activity.
type Metrics struct {
	Enqueued    atomic.Int64
	Dequeued    atomic.Int64
	Requeued    atomic.Int64
	Completed   atomic.Int64
	Failed      atomic.Int64
	DeadLettered atomic.Int64
}

// New returns a new zero-valued Metrics instance.
func New() *Metrics {
	return &Metrics{}
}

// Snapshot returns a point-in-time copy of all counters.
func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		Enqueued:     m.Enqueued.Load(),
		Dequeued:     m.Dequeued.Load(),
		Requeued:     m.Requeued.Load(),
		Completed:    m.Completed.Load(),
		Failed:       m.Failed.Load(),
		DeadLettered: m.DeadLettered.Load(),
	}
}

// Snapshot is an immutable copy of Metrics at a point in time.
type Snapshot struct {
	Enqueued     int64
	Dequeued     int64
	Requeued     int64
	Completed    int64
	Failed       int64
	DeadLettered int64
}

// InFlight returns the number of jobs currently being processed.
// This is calculated as Dequeued minus the sum of Completed, Failed,
// and DeadLettered jobs.
func (s Snapshot) InFlight() int64 {
	return s.Dequeued - s.Completed - s.Failed - s.DeadLettered
}
