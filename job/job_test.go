package job

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	j := New("abc-123", []byte(`{"task":"send_email"}`), 5)

	if j.ID != "abc-123" {
		t.Errorf("expected ID abc-123, got %s", j.ID)
	}
	if j.Status != StatusPending {
		t.Errorf("expected status pending, got %s", j.Status)
	}
	if j.Attempts != 0 {
		t.Errorf("expected 0 attempts, got %d", j.Attempts)
	}
	if j.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", j.MaxAttempts)
	}
}

func TestIsExhausted(t *testing.T) {
	j := New("x", nil, 3)

	if j.IsExhausted() {
		t.Error("new job should not be exhausted")
	}

	j.Attempts = 3
	if !j.IsExhausted() {
		t.Error("job with attempts == maxAttempts should be exhausted")
	}
}

func TestMarkFailed_NotExhausted(t *testing.T) {
	j := New("y", nil, 3)
	nextRun := time.Now().Add(10 * time.Second)

	j.MarkFailed(errors.New("timeout"), nextRun)

	if j.Status != StatusFailed {
		t.Errorf("expected status failed, got %s", j.Status)
	}
	if j.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", j.Attempts)
	}
	if j.LastError != "timeout" {
		t.Errorf("expected last error 'timeout', got %s", j.LastError)
	}
	if !j.NextRunAt.Equal(nextRun) {
		t.Errorf("NextRunAt not set correctly")
	}
}

func TestMarkFailed_Exhausted(t *testing.T) {
	j := New("z", nil, 1)

	j.MarkFailed(errors.New("fatal"), time.Now().Add(time.Minute))

	if j.Status != StatusDead {
		t.Errorf("expected status dead, got %s", j.Status)
	}
}

func TestMarkComplete(t *testing.T) {
	j := New("w", nil, 3)
	j.MarkComplete()

	if j.Status != StatusComplete {
		t.Errorf("expected status complete, got %s", j.Status)
	}
}
