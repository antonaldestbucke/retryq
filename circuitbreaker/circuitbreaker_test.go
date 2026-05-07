package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/example/retryq/circuitbreaker"
)

func TestNew_DefaultsToClosedState(t *testing.T) {
	cb := circuitbreaker.New()
	if cb.CurrentState() != circuitbreaker.StateClosed {
		t.Fatalf("expected closed state, got %v", cb.CurrentState())
	}
}

func TestAllow_ClosedCircuit(t *testing.T) {
	cb := circuitbreaker.New()
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil error on closed circuit, got %v", err)
	}
}

func TestRecordFailure_OpensCircuitAtThreshold(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.WithThreshold(3))
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.CurrentState() != circuitbreaker.StateOpen {
		t.Fatalf("expected open state after threshold failures")
	}
	if err := cb.Allow(); err != circuitbreaker.ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.WithThreshold(1))
	cb.RecordFailure()
	cb.RecordSuccess()
	if cb.CurrentState() != circuitbreaker.StateClosed {
		t.Fatalf("expected closed state after success")
	}
}

func TestAllow_HalfOpenAfterTimeout(t *testing.T) {
	now := time.Now()
	cb := circuitbreaker.New(
		circuitbreaker.WithThreshold(1),
		circuitbreaker.WithResetTimeout(10*time.Second),
		circuitbreaker.WithClock(func() time.Time { return now }),
	)
	cb.RecordFailure()

	// still within reset window — should be open
	if err := cb.Allow(); err != circuitbreaker.ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen within window, got %v", err)
	}

	// advance clock past reset timeout
	now = now.Add(11 * time.Second)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil after reset timeout, got %v", err)
	}
	if cb.CurrentState() != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected half-open state, got %v", cb.CurrentState())
	}
}

func TestHalfOpen_FailureReopensCircuit(t *testing.T) {
	now := time.Now()
	cb := circuitbreaker.New(
		circuitbreaker.WithThreshold(1),
		circuitbreaker.WithResetTimeout(5*time.Second),
		circuitbreaker.WithClock(func() time.Time { return now }),
	)
	cb.RecordFailure()
	now = now.Add(6 * time.Second)
	_ = cb.Allow() // transitions to half-open
	cb.RecordFailure()
	if cb.CurrentState() != circuitbreaker.StateOpen {
		t.Fatalf("expected open after failure in half-open state")
	}
}
