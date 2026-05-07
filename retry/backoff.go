// Package retry provides retry policies, jitter strategies, and backoff
// helpers for use with retryq job processing.
package retry

import (
	"math"
	"time"
)

// BackoffStrategy computes the delay before the next retry attempt.
type BackoffStrategy interface {
	Next(attempt int) time.Duration
}

// ExponentialBackoff implements exponential backoff with a configurable
// base delay, multiplier, and maximum delay cap.
type ExponentialBackoff struct {
	BaseDelay  time.Duration
	Multiplier float64
	MaxDelay   time.Duration
}

// DefaultExponentialBackoff returns an ExponentialBackoff with sensible
// defaults: 500ms base, factor of 2, capped at 5 minutes.
func DefaultExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		BaseDelay:  500 * time.Millisecond,
		Multiplier: 2.0,
		MaxDelay:   5 * time.Minute,
	}
}

// Next returns the delay for the given attempt number (0-indexed).
// The delay is clamped to MaxDelay.
func (e *ExponentialBackoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := float64(e.BaseDelay) * math.Pow(e.Multiplier, float64(attempt))
	if delay > float64(e.MaxDelay) {
		return e.MaxDelay
	}
	return time.Duration(delay)
}

// LinearBackoff implements a simple linear backoff strategy where each
// attempt adds a fixed increment to the base delay.
type LinearBackoff struct {
	BaseDelay time.Duration
	Increment time.Duration
	MaxDelay  time.Duration
}

// Next returns the delay for the given attempt using linear growth.
func (l *LinearBackoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := l.BaseDelay + time.Duration(attempt)*l.Increment
	if l.MaxDelay > 0 && delay > l.MaxDelay {
		return l.MaxDelay
	}
	return delay
}

// ConstantBackoff always returns the same delay regardless of attempt.
type ConstantBackoff struct {
	Delay time.Duration
}

// Next returns the constant delay for any attempt.
func (c *ConstantBackoff) Next(_ int) time.Duration {
	return c.Delay
}
