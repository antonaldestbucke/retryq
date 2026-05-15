package quota_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/quota"
)

func newJob(id string) *job.Job {
	return job.New(id, []byte(`{}`))
}

func keyByID(j *job.Job) string { return j.ID }

func TestWithQuota_AllowsWhenUnderLimit(t *testing.T) {
	q := quota.New(2, time.Minute, quota.WithClock(fixedClock(time.Now())))
	var called bool
	handler := func(j *job.Job) error { called = true; return nil }

	mw := quota.WithQuota(q, keyByID)(handler)
	if err := mw(newJob("j1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestWithQuota_BlocksWhenExhausted(t *testing.T) {
	q := quota.New(1, time.Minute, quota.WithClock(fixedClock(time.Now())))
	calls := 0
	handler := func(j *job.Job) error { calls++; return nil }

	mw := quota.WithQuota(q, keyByID)(handler)
	j := newJob("same-key")
	_ = mw(j)
	err := mw(j)

	if !errors.Is(err, quota.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected handler called once, got %d", calls)
	}
}

func TestWithQuota_DoesNotSwallowHandlerError(t *testing.T) {
	sentinel := errors.New("handler error")
	q := quota.New(10, time.Minute, quota.WithClock(fixedClock(time.Now())))
	handler := func(j *job.Job) error { return sentinel }

	mw := quota.WithQuota(q, keyByID)(handler)
	if err := mw(newJob("k")); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWithQuota_IndependentKeys(t *testing.T) {
	q := quota.New(1, time.Minute, quota.WithClock(fixedClock(time.Now())))
	handler := func(j *job.Job) error { return nil }
	mw := quota.WithQuota(q, keyByID)(handler)

	if err := mw(newJob("a")); err != nil {
		t.Fatalf("unexpected error for key a: %v", err)
	}
	if err := mw(newJob("b")); err != nil {
		t.Fatalf("unexpected error for key b: %v", err)
	}
}
