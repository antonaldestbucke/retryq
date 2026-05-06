package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/example/retryq/job"
	"github.com/example/retryq/metrics"
	"github.com/example/retryq/middleware"
)

// ExampleChain demonstrates composing multiple middlewares around a job handler.
func ExampleChain() {
	logger := log.New(os.Stdout, "", 0)
	m := metrics.New()

	base := func(ctx context.Context, j *job.Job) error {
		if j.Type == "bad" {
			return errors.New("unsupported job type")
		}
		return nil
	}

	h := middleware.Chain(
		base,
		middleware.WithRecovery(),
		middleware.WithMetrics(m),
		middleware.WithLogging(logger),
	)

	goodJob, _ := job.New("email", []byte(`{}`))
	_ = h(context.Background(), goodJob)

	snap := m.Snapshot()
	fmt.Printf("processed=%d failed=%d\n", snap.Processed, snap.Failed)
}
