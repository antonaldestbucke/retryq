package window_test

import (
	"testing"
	"time"

	"github.com/user/retryq/window"
)

func TestCount_EmptyIsZero(t *testing.T) {
	c := window.New(time.Second)
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAdd_IncrementsCount(t *testing.T) {
	c := window.New(time.Second)
	c.Add(3)
	c.Add(2)
	if got := c.Count(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCount_EvictsExpiredEntries(t *testing.T) {
	now := time.Unix(1_000, 0)
	clock := func() time.Time { return now }

	c := window.New(time.Second, window.WithClock(clock))
	c.Add(10)

	// advance time beyond the window
	now = now.Add(2 * time.Second)
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0 after expiry, got %d", got)
	}
}

func TestCount_PartialEviction(t *testing.T) {
	now := time.Unix(1_000, 0)
	clock := func() time.Time { return now }

	c := window.New(time.Second, window.WithClock(clock))
	c.Add(5)

	now = now.Add(500 * time.Millisecond)
	c.Add(3)

	// advance just past the first entry
	now = now.Add(600 * time.Millisecond)
	if got := c.Count(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestReset_ClearsAll(t *testing.T) {
	c := window.New(time.Second)
	c.Add(7)
	c.Reset()
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}
