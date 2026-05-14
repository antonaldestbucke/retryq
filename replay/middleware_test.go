package replay_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/retryq/replay"
	"github.com/example/retryq/storage"
)

func TestWithDeadLetterReplay_PassesThroughSuccess(t *testing.T) {
	dead := storage.NewDeadLetterStore()
	mw := replay.WithDeadLetterReplay(dead)

	handler := mw(func(_ context.Context, j *job.Job) error {
		return nil
	})

	j := newJob("ok")
	if err := handler(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	records, _ := dead.List()
	if len(records) != 0 {
		t.Errorf("expected no dead-letter records, got %d", len(records))
	}
}

func TestWithDeadLetterReplay_PassesThroughNonExhaustedError(t *testing.T) {
	dead := storage.NewDeadLetterStore()
	mw := replay.WithDeadLetterReplay(dead)
	want := errors.New("transient")

	handler := mw(func(_ context.Context, j *job.Job) error {
		return want
	})

	j := newJob("retry-me")
	// Attempts < MaxAttempts so not exhausted.
	got := handler(context.Background(), j)
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	records, _ := dead.List()
	if len(records) != 0 {
		t.Errorf("expected no dead-letter records, got %d", len(records))
	}
}

func TestWithDeadLetterReplay_StoresExhaustedJob(t *testing.T) {
	dead := storage.NewDeadLetterStore()
	mw := replay.WithDeadLetterReplay(dead)

	handler := mw(func(_ context.Context, j *job.Job) error {
		return errors.New("final failure")
	})

	j := newJob("exhausted")
	// Exhaust the job.
	for !j.IsExhausted() {
		j.MarkFailed(errors.New("x"), 0)
	}

	if err := handler(context.Background(), j); err != nil {
		t.Fatalf("expected nil (dead-lettered), got %v", err)
	}
	records, _ := dead.List()
	if len(records) != 1 {
		t.Fatalf("expected 1 dead-letter record, got %d", len(records))
	}
	if records[0].Job.ID != "exhausted" {
		t.Errorf("unexpected job ID %q", records[0].Job.ID)
	}
}
