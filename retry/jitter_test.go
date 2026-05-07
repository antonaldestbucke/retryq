package retry

import (
	"testing"
	"time"
)

func TestJitterNone_ReturnsSameDelay(t *testing.T) {
	j := NewJitter(JitterNone)
	base := 4 * time.Second

	for i := 0; i < 20; i++ {
		got := j.Apply(base)
		if got != base {
			t.Fatalf("JitterNone: expected %v, got %v", base, got)
		}
	}
}

func TestJitterFull_InRange(t *testing.T) {
	j := NewJitter(JitterFull)
	base := 8 * time.Second

	for i := 0; i < 100; i++ {
		got := j.Apply(base)
		if got < 0 || got >= base {
			t.Fatalf("JitterFull: %v out of range [0, %v)", got, base)
		}
	}
}

func TestJitterEqual_InRange(t *testing.T) {
	j := NewJitter(JitterEqual)
	base := 8 * time.Second
	half := base / 2

	for i := 0; i < 100; i++ {
		got := j.Apply(base)
		if got < half || got >= base {
			t.Fatalf("JitterEqual: %v out of range [%v, %v)", got, half, base)
		}
	}
}

func TestJitter_ZeroBase(t *testing.T) {
	strategies := []JitterStrategy{JitterNone, JitterFull, JitterEqual}

	for _, s := range strategies {
		j := NewJitter(s)
		if got := j.Apply(0); got != 0 {
			t.Errorf("strategy %d: expected 0 for zero base, got %v", s, got)
		}
	}
}

func TestJitter_NegativeBase(t *testing.T) {
	strategies := []JitterStrategy{JitterNone, JitterFull, JitterEqual}
	neg := -1 * time.Second

	for _, s := range strategies {
		j := NewJitter(s)
		if got := j.Apply(neg); got != neg {
			t.Errorf("strategy %d: expected %v for negative base, got %v", s, neg, got)
		}
	}
}
