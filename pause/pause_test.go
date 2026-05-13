package pause_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/retryq/pause"
)

func TestNew_NotPaused(t *testing.T) {
	g := pause.New()
	if g.IsPaused() {
		t.Fatal("expected gate to start unpaused")
	}
}

func TestPause_And_IsPaused(t *testing.T) {
	g := pause.New()
	g.Pause()
	if !g.IsPaused() {
		t.Fatal("expected gate to be paused")
	}
}

func TestResume_ClearsPause(t *testing.T) {
	g := pause.New()
	g.Pause()
	g.Resume()
	if g.IsPaused() {
		t.Fatal("expected gate to be unpaused after Resume")
	}
}

func TestWait_ReturnImmediatelyWhenNotPaused(t *testing.T) {
	g := pause.New()
	ctx := context.Background()
	if err := g.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWait_BlocksUntilResumed(t *testing.T) {
	g := pause.New()
	g.Pause()

	var reached atomic.Bool
	go func() {
		time.Sleep(20 * time.Millisecond)
		g.Resume()
	}()

	ctx := context.Background()
	if err := g.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reached.Store(true)

	if !reached.Load() {
		t.Fatal("Wait did not unblock after Resume")
	}
}

func TestWait_RespectsContextCancellation(t *testing.T) {
	g := pause.New()
	g.Pause()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := g.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestWait_MultipleWaitersUnblockedByResume(t *testing.T) {
	g := pause.New()
	g.Pause()

	var count atomic.Int32
	for i := 0; i < 5; i++ {
		go func() {
			_ = g.Wait(context.Background())
			count.Add(1)
		}()
	}

	time.Sleep(20 * time.Millisecond)
	g.Resume()
	time.Sleep(30 * time.Millisecond)

	if n := count.Load(); n != 5 {
		t.Fatalf("expected 5 goroutines unblocked, got %d", n)
	}
}
