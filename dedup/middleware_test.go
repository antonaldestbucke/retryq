package dedup_test

import (
	"errors"
	"testing"

	"github.com/example/retryq/dedup"
	"github.com/example/retryq/job"
)

func TestWithDedup_AllowsFirstJob(t *testing.T) {
	store := dedup.New()
	called := false
	h := dedup.WithDedup(store, func(j *job.Job) error {
		called = true
		return nil
	})

	j := newJob("unique-1")
	if err := h(j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestWithDedup_BlocksDuplicate(t *testing.T) {
	store := dedup.New()
	h := dedup.WithDedup(store, func(j *job.Job) error { return nil })

	j := newJob("dup-job")
	// First call records the key but the handler removes it on success.
	// Call twice without intervening success to simulate an in-flight duplicate.
	store.IsDuplicate(j) // manually seed the store

	if err := h(j); !errors.Is(err, dedup.ErrDuplicate) {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestWithDedup_KeyRemovedOnSuccess(t *testing.T) {
	store := dedup.New()
	h := dedup.WithDedup(store, func(j *job.Job) error { return nil })

	j := newJob("success-job")
	if err := h(j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After success the key should be cleared; a second call must not be a duplicate.
	if err := h(j); errors.Is(err, dedup.ErrDuplicate) {
		t.Fatal("key should have been removed after successful processing")
	}
}

func TestWithDedup_KeyRetainedOnError(t *testing.T) {
	store := dedup.New()
	handlerErr := errors.New("processing failed")
	h := dedup.WithDedup(store, func(j *job.Job) error { return handlerErr })

	j := newJob("fail-job")
	if err := h(j); !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
	// Key should still be present; next attempt should be treated as duplicate.
	if err := h(j); !errors.Is(err, dedup.ErrDuplicate) {
		t.Fatalf("expected ErrDuplicate on retry, got %v", err)
	}
}
