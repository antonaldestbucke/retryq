package ratelimit_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/retryq/ratelimit"
)

func TestNew_InitialTokens(t *testing.T) {
	l := ratelimit.New(10, 5)
	if got := l.Tokens(); got != 5 {
		t.Fatalf("expected 5 tokens, got %v", got)
	}
}

func TestAllow_ConsumesToken(t *testing.T) {
	l := ratelimit.New(10, 3)
	if !l.Allow() {
		t.Fatal("expected Allow to return true")
	}
	if got := l.Tokens(); got != 2 {
		t.Fatalf("expected 2 tokens after one Allow, got %v", got)
	}
}

func TestAllow_DeniesWhenEmpty(t *testing.T) {
	l := ratelimit.New(0, 1)
	l.Allow() // consume the only token
	if l.Allow() {
		t.Fatal("expected Allow to return false when no tokens remain")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	l := ratelimit.New(10, 0, ratelimit.WithClock(clock))

	if l.Allow() {
		t.Fatal("expected false with zero burst and no elapsed time")
	}

	// Advance time by 1 second — should add 10 tokens (capped at burst=0... wait, burst=10)
	l2 := ratelimit.New(10, 0, ratelimit.WithClock(clock))
	now = time.Unix(1, 0)
	if !l2.Allow() {
		t.Fatal("expected Allow to succeed after 1s with rate=10")
	}
	_ = l
}

func TestAllow_CapsAtBurst(t *testing.T) {
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	l := ratelimit.New(5, 3, ratelimit.WithClock(clock))
	now = time.Unix(100, 0) // large elapsed time
	l.Allow()              // trigger refill

	if got := l.Tokens(); got > 3 {
		t.Fatalf("tokens should be capped at burst=3, got %v", got)
	}
}

func TestAllow_Concurrent(t *testing.T) {
	l := ratelimit.New(1000, 100)
	var allowed int64
	doneCh := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				if l.Allow() {
					atomic.AddInt64(&allowed, 1)
				}
			}
			doneCh <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-doneCh
	}
	if atomic.LoadInt64(&allowed) > 100 {
		t.Fatalf("allowed %d > burst 100", allowed)
	}
}
