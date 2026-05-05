// Package storage provides pluggable persistence backends for retryq.
//
// It exposes two primary abstractions:
//
//   - Store: a key-value interface for saving and loading job Records.
//     The built-in MemoryStore satisfies this interface and is suitable
//     for testing or ephemeral workloads.
//
//   - DeadLetterStore: a dedicated store for jobs that have exhausted all
//     retry attempts. Entries can be inspected, replayed, or purged by
//     operators.
//
// # Record
//
// A Record is the serialisable snapshot of a job's state, including its
// payload, attempt counters, last error message, and arbitrary metadata.
//
// # Extending
//
// To use a durable backend (e.g. Redis, PostgreSQL), implement the Store
// interface and pass it to the queue or worker constructors.
package storage
