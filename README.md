# Idempotent Payment Simulator

---

This project simulates a **payment processing service** written in Go that guarantees safe retries using idempotency, a critical requirement in real-world distributed systems.

In production environments, clients often retry payment requests due to timeouts, network failures, or user behavior. Without proper safeguards, these retries can result in duplicate charges and inconsistent system state.

The service is implemented in **Go** with **PostgreSQL** and requires clients to send an `Idempotency-Key` header with each payment request. When the same key is received more than once, the system returns the original stored response instead of processing the payment again, ensuring that each logical payment is executed only once.

---

## Technical Highlights

* Deterministic request hashing
* Atomic database constraints
* Transaction-safe payment processing
* Concurrency-safe idempotency handling
* High-performance Go HTTP server



