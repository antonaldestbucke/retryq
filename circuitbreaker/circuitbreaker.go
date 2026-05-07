// Package circuitbreaker provides a circuit breaker middleware for job handlers.
// It tracks consecutive failures and opens the circuit to prevent cascading failures.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // rejecting requests
	StateHalfOpen              // testing recovery
)

// ErrCircuitOpen is returned when the circuit is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker tracks failure rates and opens the circuit when a threshold is exceeded.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
	now          func() time.Time
}

// Option configures a CircuitBreaker.
type Option func(*CircuitBreaker)

// WithThreshold sets the number of consecutive failures before opening.
func WithThreshold(n int) Option {
	return func(cb *CircuitBreaker) { cb.threshold = n }
}

// WithResetTimeout sets how long the circuit stays open before attempting recovery.
func WithResetTimeout(d time.Duration) Option {
	return func(cb *CircuitBreaker) { cb.resetTimeout = d }
}

// WithClock overrides the time source (for testing).
func WithClock(fn func() time.Time) Option {
	return func(cb *CircuitBreaker) { cb.now = fn }
}

// New creates a CircuitBreaker with the given options.
func New(opts ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		threshold:    5,
		resetTimeout: 30 * time.Second,
		now:          time.Now,
	}
	for _, o := range opts {
		o(cb)
	}
	return cb
}

// Allow reports whether a request should be allowed through.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if cb.now().Sub(cb.openedAt) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess records a successful operation and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure records a failed operation and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == StateHalfOpen || cb.failures >= cb.threshold {
		cb.state = StateOpen
		cb.openedAt = cb.now()
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
