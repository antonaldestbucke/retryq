package pause

import (
	"context"

	"github.com/your-org/retryq/job"
)

// WithPause returns a Handler that blocks while the Gate is paused before
// delegating to next. If the context is cancelled during the wait the error
// is returned immediately without invoking next.
func WithPause(g *Gate, next Handler) Handler {
	return func(ctx context.Context, j *job.Job) error {
		if err := g.Wait(ctx); err != nil {
			return err
		}
		return next(ctx, j)
	}
}
