package sampling_test

import (
	"errors"
	"testing"

	"github.com/yourorg/retryq/job"
	"github.com/yourorg/retryq/sampling"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func TestNew_DefaultRate(t *testing.T) {
	s := sampling.New(0.5)
	if s.Rate() != 0.5 {
		t.Fatalf("expected rate 0.5, got %v", s.Rate())
	}
}

func TestNew_ClampsRate(t *testing.T) {
	if sampling.New(-1).Rate() != 0.0 {
		t.Fatal("expected rate clamped to 0.0")
	}
	if sampling.New(2.0).Rate() != 1.0 {
		t.Fatal("expected rate clamped to 1.0")
	}
}

func TestAllow_AlwaysWhenRateOne(t *testing.T) {
	s := sampling.New(1.0)
	j := newJob(t)
	for i := 0; i < 20; i++ {
		if !s.Allow(j) {
			t.Fatal("expected Allow=true with rate=1.0")
		}
	}
}

func TestAllow_NeverWhenRateZero(t *testing.T) {
	s := sampling.New(0.0)
	j := newJob(t)
	for i := 0; i < 20; i++ {
		if s.Allow(j) {
			t.Fatal("expected Allow=false with rate=0.0")
		}
	}
}

func TestSetRate_UpdatesRate(t *testing.T) {
	s := sampling.New(0.5)
	s.SetRate(0.9)
	if s.Rate() != 0.9 {
		t.Fatalf("expected rate 0.9, got %v", s.Rate())
	}
}

func TestWithSampling_AllowsJob(t *testing.T) {
	s := sampling.New(1.0)
	handled := false
	h := sampling.WithSampling(s)(func(j *job.Job) error {
		handled = true
		return nil
	})
	if err := h(newJob(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Fatal("expected handler to be called")
	}
}

func TestWithSampling_SkipsJob(t *testing.T) {
	s := sampling.New(0.0)
	h := sampling.WithSampling(s)(func(j *job.Job) error {
		t.Fatal("handler should not be called")
		return nil
	})
	err := h(newJob(t))
	if !errors.Is(err, sampling.ErrSkipped) {
		t.Fatalf("expected ErrSkipped, got %v", err)
	}
}

func TestWithSampling_PropagatesHandlerError(t *testing.T) {
	s := sampling.New(1.0)
	sentinel := errors.New("handler error")
	h := sampling.WithSampling(s)(func(j *job.Job) error {
		return sentinel
	})
	if err := h(newJob(t)); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWithSource_Deterministic(t *testing.T) {
	calls := 0
	src := func() float64 {
		calls++
		return 0.3 // always below 0.5 → always allowed
	}
	s := sampling.New(0.5, sampling.WithSource(src))
	j := newJob(t)
	for i := 0; i < 5; i++ {
		if !s.Allow(j) {
			t.Fatal("expected Allow=true with src returning 0.3")
		}
	}
	if calls != 5 {
		t.Fatalf("expected 5 src calls, got %d", calls)
	}
}
