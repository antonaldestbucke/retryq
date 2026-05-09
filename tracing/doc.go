// Package tracing provides lightweight span-based tracing for retryq job
// executions.
//
// A Tracer records a Span for each job processed through the middleware,
// capturing the job ID, queue name, attempt number, start/end timestamps, and
// any handler error.
//
// Usage:
//
//	tr := tracing.New()
//	mw := tracing.WithTracing(tr)
//
//	// attach to worker via middleware.Chain
//	chain := middleware.Chain(mw)
//
//	// inspect recorded spans after processing
//	for _, span := range tr.Spans() {
//		fmt.Printf("job=%s duration=%s err=%v\n",
//			span.JobID, span.Duration(), span.Err)
//	}
package tracing
