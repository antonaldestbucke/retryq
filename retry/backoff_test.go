package retry

import (
	"testing"
	"time"
)

func TestExponentialBackoff_Defaults(t *testing.T) {
	b := DefaultExponentialBackoff()
	if b.BaseDelay != 500*time.Millisecond {
		t.Errorf("expected BaseDelay 500ms, got %v", b.BaseDelay)
	}
	if b.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", b.Multiplier)
	}
	if b.MaxDelay != 5*time.Minute {
		t.Errorf("expected MaxDelay 5m, got %v", b.MaxDelay)
	}
}

func TestExponentialBackoff_GrowsWithAttempt(t *testing.T) {
	b := DefaultExponentialBackoff()
	prev := b.Next(0)
	for attempt := 1; attempt <= 5; attempt++ {
		curr := b.Next(attempt)
		if curr <= prev {
			t.Errorf("attempt %d: expected delay > %v, got %v", attempt, prev, curr)
		}
		prev = curr
	}
}

func TestExponentialBackoff_CapsAtMaxDelay(t *testing.T) {
	b := &ExponentialBackoff{
		BaseDelay:  100 * time.Millisecond,
		Multiplier: 10.0,
		MaxDelay:   1 * time.Second,
	}
	for attempt := 5; attempt <= 10; attempt++ {
		d := b.Next(attempt)
		if d != 1*time.Second {
			t.Errorf("attempt %d: expected cap 1s, got %v", attempt, d)
		}
	}
}

func TestExponentialBackoff_NegativeAttemptClamped(t *testing.T) {
	b := DefaultExponentialBackoff()
	if b.Next(-3) != b.Next(0) {
		t.Error("negative attempt should behave like attempt 0")
	}
}

func TestLinearBackoff_GrowsLinearly(t *testing.T) {
	b := &LinearBackoff{
		BaseDelay: 1 * time.Second,
		Increment: 500 * time.Millisecond,
		MaxDelay:  0,
	}
	expected := []time.Duration{
		1 * time.Second,
		1500 * time.Millisecond,
		2 * time.Second,
	}
	for i, want := range expected {
		got := b.Next(i)
		if got != want {
			t.Errorf("attempt %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestLinearBackoff_CapsAtMaxDelay(t *testing.T) {
	b := &LinearBackoff{
		BaseDelay: 1 * time.Second,
		Increment: 1 * time.Second,
		MaxDelay:  3 * time.Second,
	}
	for attempt := 5; attempt <= 10; attempt++ {
		d := b.Next(attempt)
		if d != 3*time.Second {
			t.Errorf("attempt %d: expected cap 3s, got %v", attempt, d)
		}
	}
}

func TestConstantBackoff_AlwaysSame(t *testing.T) {
	b := &ConstantBackoff{Delay: 2 * time.Second}
	for attempt := 0; attempt <= 10; attempt++ {
		if d := b.Next(attempt); d != 2*time.Second {
			t.Errorf("attempt %d: expected 2s, got %v", attempt, d)
		}
	}
}
