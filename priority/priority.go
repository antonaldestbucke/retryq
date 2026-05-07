// Package priority provides a priority-aware job queue that processes
// higher-priority jobs before lower-priority ones.
package priority

import (
	"container/heap"
	"sync"

	"github.com/example/retryq/job"
)

// Level represents a job priority level.
type Level int

const (
	Low    Level = 0
	Normal Level = 5
	High   Level = 10
)

// entry is an internal heap element.
type entry struct {
	j        *job.Job
	priority Level
	index    int
}

// entryHeap implements heap.Interface.
type entryHeap []*entry

func (h entryHeap) Len() int            { return len(h) }
func (h entryHeap) Less(i, j int) bool  { return h[i].priority > h[j].priority }
func (h entryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *entryHeap) Push(x any) {
	e := x.(*entry)
	e.index = len(*h)
	*h = append(*h, e)
}
func (h *entryHeap) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return e
}

// Queue is a thread-safe priority job queue.
type Queue struct {
	mu   sync.Mutex
	heap entryHeap
}

// New creates an empty priority Queue.
func New() *Queue {
	h := entryHeap{}
	heap.Init(&h)
	return &Queue{heap: h}
}

// Enqueue adds a job with the given priority level.
func (q *Queue) Enqueue(j *job.Job, p Level) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.heap, &entry{j: j, priority: p})
}

// Dequeue removes and returns the highest-priority job.
// Returns nil if the queue is empty.
func (q *Queue) Dequeue() *job.Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.heap.Len() == 0 {
		return nil
	}
	e := heap.Pop(&q.heap).(*entry)
	return e.j
}

// Len returns the number of jobs currently in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.heap.Len()
}
