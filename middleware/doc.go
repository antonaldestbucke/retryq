// Package middleware provides composable middleware for retryq job handlers.
//
// Middleware functions follow the pattern:
//
//	type Middleware func(next Handler) Handler
//
// Use Chain to compose multiple middlewares together:
//
//	h := middleware.Chain(
//		myHandler,
//		middleware.WithRecovery(),
//		middleware.WithLogging(logger),
//		middleware.WithMetrics(m),
//	)
//
// Middlewares are applied in the order provided, so the first middleware in the
// list is the outermost wrapper (executed first on the way in, last on the way
// out). In the example above, WithRecovery wraps everything, ensuring panics
// from inner middlewares or the handler itself are caught.
//
// Built-in middlewares:
//   - WithLogging: logs job start, success, and failure with elapsed time.
//   - WithMetrics: increments processed/failed counters on a metrics.Metrics instance.
//   - WithRecovery: catches panics and converts them to errors to prevent worker crashes.
package middleware
