// Package tagging provides job tagging and filtering support.
// Tags are arbitrary string labels attached to a job that can be
// used to route, filter, or categorize work at runtime.
package tagging

import (
	"sync"

	"github.com/example/retryq/job"
)

// Store holds a mapping from job ID to its set of tags.
type Store struct {
	mu   sync.RWMutex
	tags map[string]map[string]struct{}
}

// New returns an initialised Store.
func New() *Store {
	return &Store{tags: make(map[string]map[string]struct{})}
}

// Set replaces all tags for the given job.
func (s *Store) Set(j *job.Job, tags ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	set := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		set[t] = struct{}{}
	}
	s.tags[j.ID] = set
}

// Add appends one or more tags to a job without removing existing ones.
func (s *Store) Add(j *job.Job, tags ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tags[j.ID] == nil {
		s.tags[j.ID] = make(map[string]struct{})
	}
	for _, t := range tags {
		s.tags[j.ID][t] = struct{}{}
	}
}

// Remove deletes specific tags from a job.
func (s *Store) Remove(j *job.Job, tags ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range tags {
		delete(s.tags[j.ID], t)
	}
}

// Has reports whether a job carries all of the specified tags.
func (s *Store) Has(j *job.Job, tags ...string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := s.tags[j.ID]
	for _, t := range tags {
		if _, ok := set[t]; !ok {
			return false
		}
	}
	return true
}

// Get returns a snapshot of all tags for a job.
func (s *Store) Get(j *job.Job) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := s.tags[j.ID]
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	return out
}

// Delete removes all tag data for a job.
func (s *Store) Delete(j *job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, j.ID)
}
