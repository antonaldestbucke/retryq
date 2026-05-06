// Package hooks provides lifecycle event callbacks for job processing.
// Hooks allow external code to react to job state transitions such as
// enqueue, success, failure, and dead-letter without modifying core logic.
package hooks

import "github.com/example/retryq/job"

// EventType represents a job lifecycle event.
type EventType string

const (
	EventEnqueued   EventType = "enqueued"
	EventSucceeded  EventType = "succeeded"
	EventFailed     EventType = "failed"
	EventDeadLetter EventType = "dead_letter"
)

// Handler is a function invoked when a lifecycle event occurs.
type Handler func(event EventType, j *job.Job)

// Registry holds registered handlers for each event type.
type Registry struct {
	handlers map[EventType][]Handler
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{
		handlers: make(map[EventType][]Handler),
	}
}

// On registers a Handler for the given EventType.
func (r *Registry) On(event EventType, h Handler) {
	r.handlers[event] = append(r.handlers[event], h)
}

// Emit calls all handlers registered for the given EventType.
// Handlers are invoked synchronously in registration order.
func (r *Registry) Emit(event EventType, j *job.Job) {
	for _, h := range r.handlers[event] {
		h(event, j)
	}
}

// Clear removes all handlers for the given EventType.
func (r *Registry) Clear(event EventType) {
	delete(r.handlers, event)
}
