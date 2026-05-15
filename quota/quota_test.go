package quota_test

import (
	"testing"
	"time"

	"github.com/example/retryq/quota"
)

func fixedClock(t time.Time) quota.Clock { return func() time.Time { return t } }

func TestAllow_UnderLimit(t *testing.T) {
	q := quota.New(3, time.Minute, quota.WithClock(fixedClock(time.Now())))
	for i := 0; i < 3; i++ {
		if !q.Allow("tenant-a") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestAllow_AtLimit(t *testing.T) {
	q := quota.New(2, time.Minute, quota.WithClock(fixedClock(time.Now())))
	q.Allow("k")
	q.Allow("k")
	if q.Allow("k") {
		t.Fatal("expected Allow to return false when limit reached")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	q := quota.New(1, time.Minute, quota.WithClock(fixedClock(time.Now())))
	if !q.Allow("a") {
		t.Fatal("expected true for key a")
	}
	if !q.Allow("b") {
		t.Fatal("expected true for key b")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	now := time.Now()
	var current time.Time = now
	clock := func() time.Time { return current }

	q := quota.New(1, time.Second, quota.WithClock(clock))
	q.Allow("k") // consume the single slot

	if q.Allow("k") {
		t.Fatal("expected false before window expires")
	}

	current = now.Add(2 * time.Second) // advance past window
	if !q.Allow("k") {
		t.Fatal("expected true after window reset")
	}
}

func TestRemaining(t *testing.T) {
	q := quota.New(5, time.Minute, quota.WithClock(fixedClock(time.Now())))
	if r := q.Remaining("k"); r != 5 {
		t.Fatalf("expected 5, got %d", r)
	}
	q.Allow("k")
	q.Allow("k")
	if r := q.Remaining("k"); r != 3 {
		t.Fatalf("expected 3, got %d", r)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	q := quota.New(1, time.Minute, quota.WithClock(fixedClock(time.Now())))
	q.Allow("k")
	if q.Allow("k") {
		t.Fatal("expected false after limit")
	}
	q.Reset("k")
	if !q.Allow("k") {
		t.Fatal("expected true after reset")
	}
}
