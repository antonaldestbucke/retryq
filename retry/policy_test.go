package retry_test

import (
	"testing"
	"time"

	"github.com/user/retryq/retry"
)

func TestDefaultPolicy_Fields(t *testing.T) {
	p := retry.DefaultPolicy()

	if p.BaseDelay != 1*time.Second {
		t.Errorf("expected BaseDelay 1s, got %v", p.BaseDelay)
	}
	if p.MaxDelay != 5*time.Minute {
		t.Errorf("expected MaxDelay 5m, got %v", p.MaxDelay)
	}
	if p.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", p.Multiplier)
	}
	if !p.Jitter {
		t.Error("expected Jitter to be true")
	}
}

func TestNextDelay_IncreasesWithAttempt(t *testing.T) {
	p := retry.Policy{
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Minute,
		Multiplier: 2.0,
		Jitter:     false,
	}

	d0 := p.NextDelay(0)
	d1 := p.NextDelay(1)
	d2 := p.NextDelay(2)

	if d0 != 1*time.Second {
		t.Errorf("attempt 0: expected 1s, got %v", d0)
	}
	if d1 != 2*time.Second {
		t.Errorf("attempt 1: expected 2s, got %v", d1)
	}
	if d2 != 4*time.Second {
		t.Errorf("attempt 2: expected 4s, got %v", d2)
	}
}

func TestNextDelay_CapsAtMaxDelay(t *testing.T) {
	p := retry.Policy{
		BaseDelay:  1 * time.Second,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		Jitter:     false,
	}

	for _, attempt := range []int{5, 10, 20} {
		d := p.NextDelay(attempt)
		if d > p.MaxDelay {
			t.Errorf("attempt %d: delay %v exceeds MaxDelay %v", attempt, d, p.MaxDelay)
		}
	}
}

func TestNextDelay_NegativeAttemptClamped(t *testing.T) {
	p := retry.Policy{
		BaseDelay:  1 * time.Second,
		MaxDelay:   1 * time.Minute,
		Multiplier: 2.0,
		Jitter:     false,
	}

	d := p.NextDelay(-3)
	if d != 1*time.Second {
		t.Errorf("expected 1s for negative attempt, got %v", d)
	}
}

func TestRetryAt_IsInFuture(t *testing.T) {
	p := retry.Policy{
		BaseDelay:  2 * time.Second,
		MaxDelay:   1 * time.Minute,
		Multiplier: 2.0,
		Jitter:     false,
	}

	now := time.Now()
	at := p.RetryAt(0, now)

	if !at.After(now) {
		t.Errorf("expected RetryAt to be after now, got %v", at)
	}
	if at.Sub(now) != 2*time.Second {
		t.Errorf("expected 2s delay, got %v", at.Sub(now))
	}
}
