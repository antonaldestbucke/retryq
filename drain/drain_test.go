package drain_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/retryq/drain"
)

func TestNew_InitialState(t *testing.T) {
	d := drain.New()
	if got := d.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight, got %d", got)
	}
}

func TestAcquire_IncrementsInflight(t *testing.T) {
	d := drain.New()
	if !d.Acquire() {
		t.Fatal("expected Acquire to return true")
	}
	if got := d.Inflight(); got != 1 {
		t.Fatalf("expected 1 inflight, got %d", got)
	}
	d.Release()
}

func TestRelease_DecrementsInflight(t *testing.T) {
	d := drain.New()
	d.Acquire()
	d.Release()
	if got := d.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight after release, got %d", got)
	}
}

func TestAcquire_ReturnsFalseAfterClose(t *testing.T) {
	d := drain.New()
	d.Close()
	if d.Acquire() {
		t.Fatal("expected Acquire to return false after Close")
	}
}

func TestWait_ReturnsWhenInflightZero(t *testing.T) {
	d := drain.New()

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		d.Acquire()
		go func() {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
			d.Release()
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	wg.Wait()
}

func TestWait_ReturnsContextError_OnTimeout(t *testing.T) {
	d := drain.New()
	d.Acquire() // never released

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := d.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	d.Release() // cleanup
}

func TestWaitWithTimeout_Convenience(t *testing.T) {
	d := drain.New()
	d.Acquire()
	go func() {
		time.Sleep(20 * time.Millisecond)
		d.Release()
	}()

	if err := d.WaitWithTimeout(time.Second); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
