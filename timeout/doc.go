// Package timeout provides deadline enforcement for job handlers in retryq.
//
// Use [WithTimeout] to wrap any handler function and automatically cancel
// execution after a specified duration, returning [ErrTimeout] to the caller.
//
// Use [WithJobTimeout] to obtain a composable [Middleware] suitable for use
// with middleware.Chain:
//
//	chain := middleware.Chain(
//	    myHandler,
//	    timeout.WithJobTimeout(5*time.Second),
//	    middleware.WithLogging(logger),
//	)
//
// [IsTimeout] can be used in retry policies or hook handlers to distinguish
// timeout failures from other error types.
package timeout
