// Package ratelimit implements a thread-safe token-bucket rate limiter
// for use with retryq workers.
//
// A Limiter is created with a sustained rate (tokens per second) and an
// initial burst capacity. Each call to Allow consumes one token; if no
// tokens are available it returns false immediately without blocking.
//
// Example usage:
//
//	limiter := ratelimit.New(100, 10) // 100 jobs/s, burst of 10
//
//	handler := func(ctx context.Context, j *job.Job) error {
//		if !limiter.Allow() {
//			return errors.New("rate limit exceeded")
//		}
//		return process(j)
//	}
package ratelimit
