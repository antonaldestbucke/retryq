package throttle_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/retryq/job"
	"github.com/yourorg/retryq/throttle"
)

// fakeClock is a manually-advanced clock for deterministic tests.
type fakeClock struct {
	now    time.Time
	slept  time.Duration
}

func (f *fakeClock) Now() time.Time        { return f.now }
func (f *fakeClock) Sleep(d time.Duration) { f.slept += d; f.now = f.now.Add(d) }

func newJob(id string) *job.Job {
	j, _ := job.New(id, []byte(`{}`))
	return j
}

func TestThrottle_FirstCallNoWait(t *testing.T) {
	clk := &fakeClock{now: time.Now()}
	th := throttle.New(100*time.Millisecond, throttle.WithClock(clk))

	var called bool
	h := th.Wrap(func(j *job.Job) error {
		called = true
		return nil
	})

	if err := h(newJob("j1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
	if clk.slept != 0 {
		t.Fatalf("expected no sleep on first call, got %v", clk.slept)
	}
}

func TestThrottle_SecondCallWaits(t *testing.T) {
	clk := &fakeClock{now: time.Now()}
	interval := 200 * time.Millisecond
	th := throttle.New(interval, throttle.WithClock(clk))

	noop := func(j *job.Job) error { return nil }
	wrapped := th.Wrap(noop)

	// First call — advances last.
	_ = wrapped(newJob("j1"))

	// Advance clock by only half the interval.
	clk.now = clk.now.Add(interval / 2)
	clk.slept = 0

	_ = wrapped(newJob("j2"))

	if clk.slept == 0 {
		t.Fatal("expected throttle to sleep on second call")
	}
	if clk.slept > interval {
		t.Fatalf("slept too long: %v (max %v)", clk.slept, interval)
	}
}

func TestThrottle_NoWaitAfterFullInterval(t *testing.T) {
	clk := &fakeClock{now: time.Now()}
	interval := 100 * time.Millisecond
	th := throttle.New(interval, throttle.WithClock(clk))

	noop := func(j *job.Job) error { return nil }
	wrapped := th.Wrap(noop)

	_ = wrapped(newJob("j1"))

	// Advance well past the interval.
	clk.now = clk.now.Add(interval * 2)
	clk.slept = 0

	_ = wrapped(newJob("j2"))

	if clk.slept != 0 {
		t.Fatalf("expected no sleep, got %v", clk.slept)
	}
}

func TestThrottle_PropagatesHandlerError(t *testing.T) {
	th := throttle.New(0)
	sentinel := errors.New("boom")
	wrapped := th.Wrap(func(j *job.Job) error { return sentinel })

	if err := wrapped(newJob("j1")); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestThrottle_ConcurrentCallsDoNotPanic(t *testing.T) {
	th := throttle.New(time.Millisecond)
	var count int64
	wrapped := th.Wrap(func(j *job.Job) error {
		atomic.AddInt64(&count, 1)
		return nil
	})

	const goroutines = 10
	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			_ = wrapped(newJob("j"))
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
	if atomic.LoadInt64(&count) != goroutines {
		t.Fatalf("expected %d calls, got %d", goroutines, count)
	}
}
