// Package sampling provides probabilistic job sampling middleware for retryq.
// It allows a fraction of jobs to be processed while skipping the rest,
// useful for load shedding, A/B testing, or gradual rollouts.
package sampling

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/yourorg/retryq/job"
)

// ErrSkipped is returned by the middleware when a job is sampled out.
var ErrSkipped = errors.New("sampling: job skipped")

// Source is a function that returns a float64 in [0.0, 1.0).
type Source func() float64

// Sampler decides probabilistically whether a job should be processed.
type Sampler struct {
	mu   sync.Mutex
	rate float64
	src  Source
}

// Option configures a Sampler.
type Option func(*Sampler)

// WithSource replaces the default random source with a custom one.
// Useful for deterministic testing.
func WithSource(src Source) Option {
	return func(s *Sampler) {
		s.src = src
	}
}

// New creates a Sampler with the given sample rate in [0.0, 1.0].
// A rate of 1.0 allows all jobs through; 0.0 skips all jobs.
func New(rate float64, opts ...Option) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := &Sampler{
		rate: rate,
		src:  r.Float64,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Allow returns true if the job should be processed based on the sample rate.
func (s *Sampler) Allow(_ *job.Job) bool {
	s.mu.Lock()
	v := s.src()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sample rate.
func (s *Sampler) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rate
}

// SetRate updates the sample rate at runtime.
func (s *Sampler) SetRate(rate float64) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	s.mu.Lock()
	s.rate = rate
	s.mu.Unlock()
}
