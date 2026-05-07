// Package retry provides retry policies, backoff strategies, and jitter
// utilities for controlling how failed jobs are rescheduled.
//
// # Policies
//
// A [Policy] defines the maximum number of attempts and the maximum delay
// between retries. Use [DefaultPolicy] to get a sensible starting point, or
// construct your own:
//
//	policy := retry.Policy{
//		MaxAttempts: 5,
//		MaxDelay:    2 * time.Minute,
//	}
//
// # Backoff
//
// [DefaultExponentialBackoff] computes delays that grow exponentially with
// each attempt. [LinearBackoff] grows at a constant rate instead.
//
// # Jitter
//
// [NewJitter] wraps any backoff function with a jitter strategy to spread
// retries across time and avoid thundering-herd problems. Three strategies
// are available: JitterNone, JitterFull, and JitterEqual.
package retry
