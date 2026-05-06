// Package ratelimit provides a token-bucket rate limiter for controlling
// the rate at which jobs are dequeued and processed by workers.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter controls the rate of job processing using a token bucket algorithm.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// Option configures a Limiter.
type Option func(*Limiter)

// WithClock sets a custom clock function (useful for testing).
func WithClock(fn func() time.Time) Option {
	return func(l *Limiter) {
		l.clock = fn
	}
}

// New creates a Limiter that allows up to burst tokens initially and
// refills at ratePerSec tokens per second.
func New(ratePerSec float64, burst int, opts ...Option) *Limiter {
	l := &Limiter{
		tokens: float64(burst),
		max:    float64(burst),
		rate:   ratePerSec,
		clock:  time.Now,
	}
	for _, o := range opts {
		o(l)
	}
	l.lastTick = l.clock()
	return l
}

// Allow reports whether one token is available and consumes it.
// It refills tokens based on elapsed time since the last call.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Tokens returns the current number of available tokens.
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tokens
}
