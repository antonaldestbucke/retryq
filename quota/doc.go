// Package quota provides per-key job processing rate limiting over a rolling
// time window.
//
// # Overview
//
// A Quota tracks how many times each key (e.g. tenant ID, user ID) has been
// allowed to process a job within a configurable window. Once the limit is
// reached, further calls to Allow return false until the window resets.
//
// # Basic usage
//
//	q := quota.New(100, time.Minute)
//
//	if q.Allow(tenantID) {
//		// process the job
//	}
//
// # Middleware
//
// WithQuota wraps a job handler and enforces the quota automatically:
//
//	mw := quota.WithQuota(q, func(j *job.Job) string {
//		return j.Meta["tenant"]
//	})
//
// If the quota is exceeded the middleware returns ErrQuotaExceeded without
// calling the underlying handler.
package quota
