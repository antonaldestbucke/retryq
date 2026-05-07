// Package retry provides configurable backoff policies for job retry scheduling.
package retry

import (
	"math"
	"math/rand"
	"time"
)

// Policy defines how retry delays are calculated for failed jobs.
type Policy struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     bool
}

// DefaultPolicy returns a Policy with sensible exponential backoff defaults.
func DefaultPolicy() Policy {
	return Policy{
		BaseDelay:  1 * time.Second,
		MaxDelay:   5 * time.Minute,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// NextDelay calculates the delay before the next retry attempt.
// attempt is zero-indexed (0 = first retry).
func (p Policy) NextDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	backoff := float64(p.BaseDelay) * math.Pow(p.Multiplier, float64(attempt))

	if backoff > float64(p.MaxDelay) {
		backoff = float64(p.MaxDelay)
	}

	delay := time.Duration(backoff)

	if p.Jitter && delay > 0 {
		// Add up to 20% random jitter to spread out retries.
		jitter := time.Duration(rand.Int63n(int64(delay / 5)))
		delay += jitter
	}

	return delay
}

// RetryAt returns the absolute time at which the next retry should occur.
func (p Policy) RetryAt(attempt int, now time.Time) time.Time {
	return now.Add(p.NextDelay(attempt))
}
