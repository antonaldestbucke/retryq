// Package retry provides retry policies and delay calculation utilities.
package retry

import (
	"math/rand"
	"time"
)

// JitterStrategy defines how jitter is applied to a retry delay.
type JitterStrategy int

const (
	// JitterNone applies no jitter — delays are deterministic.
	JitterNone JitterStrategy = iota

	// JitterFull randomises the delay uniformly between 0 and the computed delay.
	JitterFull

	// JitterEqual randomises the delay between half and the full computed delay.
	JitterEqual
)

// Jitter holds configuration for applying jitter to a base delay.
type Jitter struct {
	Strategy JitterStrategy
	rng      *rand.Rand
}

// NewJitter returns a Jitter with the given strategy and a new random source
// seeded from the current time.
func NewJitter(strategy JitterStrategy) *Jitter {
	return &Jitter{
		Strategy: strategy,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Apply returns a jittered version of base according to the configured strategy.
// If base is zero or negative it is returned unchanged.
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}

	switch j.Strategy {
	case JitterFull:
		// Uniform [0, base)
		return time.Duration(j.rng.Int63n(int64(base)))
	case JitterEqual:
		// Uniform [base/2, base)
		half := base / 2
		return half + time.Duration(j.rng.Int63n(int64(base-half)))
	default:
		return base
	}
}
