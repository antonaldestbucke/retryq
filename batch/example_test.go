package batch_test

import (
	"fmt"
	"log"

	"github.com/example/retryq/batch"
	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
	"github.com/example/retryq/storage"
)

func ExampleEnqueue() {
	store := storage.NewMemoryStore()
	dl := storage.NewDeadLetterStore()
	q := queue.New(store, dl)

	jobs := []*job.Job{
		job.New([]byte(`{"task":"send_email","to":"alice@example.com"}`), 3),
		job.New([]byte(`{"task":"send_email","to":"bob@example.com"}`), 3),
		job.New([]byte(`{"task":"send_email","to":"carol@example.com"}`), 3),
	}

	summary := batch.Enqueue(q, jobs)
	fmt.Printf("enqueued: %d, failed: %d\n", summary.Enqueued, summary.Failed)

	if summary.HasErrors() {
		for _, r := range summary.Errors {
			log.Printf("job %s failed: %v", r.Job.ID, r.Err)
		}
	}
	// Output: enqueued: 3, failed: 0
}

func ExampleFromPayloads() {
	store := storage.NewMemoryStore()
	dl := storage.NewDeadLetterStore()
	q := queue.New(store, dl)

	payloads := [][]byte{
		[]byte(`{"task":"resize_image","id":1}`),
		[]byte(`{"task":"resize_image","id":2}`),
	}

	summary := batch.FromPayloads(q, payloads, 5)
	fmt.Printf("total: %d, enqueued: %d\n", summary.Total, summary.Enqueued)
	// Output: total: 2, enqueued: 2
}
