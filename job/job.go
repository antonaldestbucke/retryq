package job

import (
	"time"
)

// Status represents the current state of a job.
type Status string

const (
	StatusPending  Status = "pending"
	StatusRunning  Status = "running"
	StatusFailed   Status = "failed"
	StatusDead     Status = "dead"
	StatusComplete Status = "complete"
)

// Job represents a unit of work managed by the retry queue.
type Job struct {
	ID          string
	Payload     []byte
	Status      Status
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	NextRunAt   time.Time
	LastError   string
}

// New creates a new Job with sensible defaults.
func New(id string, payload []byte, maxAttempts int) *Job {
	now := time.Now().UTC()
	return &Job{
		ID:          id,
		Payload:     payload,
		Status:      StatusPending,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
		NextRunAt:   now,
	}
}

// IsExhausted reports whether the job has exceeded its maximum attempts.
func (j *Job) IsExhausted() bool {
	return j.Attempts >= j.MaxAttempts
}

// MarkFailed records a failure and advances the attempt counter.
// It moves the job to Dead status if attempts are exhausted.
func (j *Job) MarkFailed(err error, nextRunAt time.Time) {
	j.Attempts++
	j.LastError = err.Error()
	j.UpdatedAt = time.Now().UTC()

	if j.IsExhausted() {
		j.Status = StatusDead
	} else {
		j.Status = StatusFailed
		j.NextRunAt = nextRunAt
	}
}

// MarkComplete sets the job status to complete.
func (j *Job) MarkComplete() {
	j.Status = StatusComplete
	j.UpdatedAt = time.Now().UTC()
}
