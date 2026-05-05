# retryq

Lightweight job retry queue with exponential backoff and dead-letter support.

## Installation

```bash
go get github.com/yourusername/retryq
```

## Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/yourusername/retryq"
)

func main() {
    q := retryq.New(retryq.Config{
        MaxRetries:  5,
        BaseDelay:   500 * time.Millisecond,
        MaxDelay:    30 * time.Second,
        DeadLetter:  true,
    })

    q.Enqueue(func() error {
        err := doSomeWork()
        if err != nil {
            return err // job will be retried with exponential backoff
        }
        return nil
    })

    q.OnDeadLetter(func(job retryq.Job) {
        fmt.Printf("job failed after all retries: %v\n", job.LastError)
    })

    q.Start()
}
```

## Features

- Exponential backoff with configurable base delay and ceiling
- Dead-letter queue for jobs that exhaust all retries
- Concurrency-safe job processing
- Zero external dependencies

## Configuration

| Option       | Default | Description                          |
|--------------|---------|--------------------------------------|
| `MaxRetries` | `3`     | Maximum number of retry attempts     |
| `BaseDelay`  | `1s`    | Initial delay between retries        |
| `MaxDelay`   | `60s`   | Maximum delay cap for backoff        |
| `DeadLetter` | `false` | Enable dead-letter queue             |

## License

MIT © [yourusername](https://github.com/yourusername)