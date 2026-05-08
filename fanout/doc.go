// Package fanout broadcasts a single job to multiple queues simultaneously.
//
// Use fanout when the same unit of work must be processed independently by
// more than one consumer — for example, sending a notification via email,
// SMS, and push at the same time.
//
// Basic usage:
//
//	f, err := fanout.New(emailQueue, smsQueue, pushQueue)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := f.Dispatch(j); err != nil {
//		// At least one queue rejected the job.
//		var de *fanout.DispatchError
//		if errors.As(err, &de) {
//			for _, e := range de.Errs {
//				log.Println(e)
//			}
//		}
//	}
//
// Dispatch is fail-soft: it attempts every queue regardless of earlier
// failures and returns a DispatchError that aggregates all rejections.
package fanout
