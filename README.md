# Idempotent Payment Simulator

A **payment processing service** written in Go that guarantees safe retries through idempotency — a critical requirement in real-world distributed systems.

In production environments, clients often retry payment requests due to timeouts, network failures, or user behavior. Without proper safeguards, retries can result in duplicate charges and inconsistent system state. This service solves that by requiring clients to send an `Idempotency-Key` header with every payment request. When the same key is received more than once, the system returns the original stored response instead of re-processing the payment, ensuring each logical operation is executed **exactly once**.

---

## Table of Contents

- [Technical Highlights](#technical-highlights)
- [Architecture](#architecture)
  - [Module Overview](#module-overview)
  - [Service Functions](#service-functions)
- [Idempotency Design](#idempotency-design)
  - [Transaction Flow](#transaction-flow)
  - [Design Principles](#design-principles)
- [API Reference](#api-reference)
  - [Endpoints](#endpoints)
- [Request & Response Examples](#request--response-examples)
- [Getting Started](#getting-started)
- [Testing](#testing)
- [Observability](#observability)

---

## Technical Highlights

| Feature | Details |
|---|---|
| Language | Go |
| Database | PostgreSQL |
| Idempotency | Header-based (`Idempotency-Key`) with deterministic request hashing |
| Concurrency | Atomic DB constraints prevent duplicate inserts under concurrent load |
| Transactions | Payment creation and idempotency record update happen atomically |
| Testing | Unit tests and full HTTP integration tests |
| Logging | Structured log output via a dedicated logger module |

---

## Architecture

### Module Overview

```
idempotent-payment/
├── idempotency/     # Idempotency key management: insert, conflict detection, hash validation
├── payment/         # Core payment logic: creation and retrieval
├── product/         # Product catalog: creation, retrieval, deletion
├── storage/         # PostgreSQL driver and connection management
├── http/            # HTTP server, routing, middleware, and handlers
└── logger/          # Structured logging
```

**`idempotency`** — The heart of the service. Manages idempotency key lifecycle (`in_progress` → `completed` | `failed`), validates request hashes to detect mismatched retries, and exposes a higher-order `Execute` function that wraps any business operation with idempotency guarantees.

**`payment`** — Handles payment creation and retrieval. Always invoked through the idempotency layer; never called directly from the HTTP layer for write operations.

**`product`** — Manages the product catalog that payments reference. Supports create, lookup, and delete.

**`storage`** — PostgreSQL driver abstraction. Provides the connection pool and transaction helpers consumed by all other modules.

**`http`** — HTTP server entrypoint. Defines routes, binds handlers, and enforces middleware (idempotency key extraction, request validation, logging).

**`logger`** — Structured, leveled logging used across all modules for observability.

### Service Functions

| Module | Functions |
|---|---|
| `payment` | `Create`, `GetByID`, `Health` |
| `idempotency` | `Execute` (higher-order function) |
| `product` | `Create`, `GetByID`, `Delete` |

---

## Idempotency Design

### Transaction Flow

```mermaid
flowchart TD
A[Receive Request + Idempotency-Key] --> B[Compute Request Hash]
B --> C[BEGIN TRANSACTION]
C --> D{Try INSERT Idempotency\nstatus=in_progress}
D -->|Success| E[Create Payment]
E --> F[Build HTTP Response]
F --> G[Update Idempotency\nstatus=completed\nstore response]
G --> H[COMMIT]
H --> I[Return Response]
D -->|Conflict| J[Load Existing Idempotency Record]
J --> K{Hash Matches?}
K -->|No| L[Return 409 Conflict\nRequest mismatch]
K -->|Yes| M{Status?}
M -->|completed| N[Return Stored Response]
M -->|in_progress| O[Return 409 Processing]
M -->|failed| P[Return Stored Failed Response]
```

### Design Principles

- The `Idempotency-Key` registers the **client's intent** before executing any side effect.
- Only one request can successfully insert a given key — guaranteed by a unique DB constraint.
- On conflict, the system first validates that the incoming request is identical to the original via a **deterministic request hash**. A mismatch returns `409 Conflict`.
- The response returned is always determined by the **persisted status** of the record, not by re-executing logic.
- Payment creation and the `IdempotencyRecord` status update are committed in the **same transaction**, preventing partial states.
- Deterministic business errors (e.g. invalid amount, unknown product) are persisted as `failed` so retries consistently receive the same error response.

---
## API Reference
 
All write endpoints require an `Idempotency-Key` header.
 
### Endpoints
 
| Method | Path | Description | Idempotency-Key |
|---|---|---|---|
| `POST` | `/payments` | Create a payment | Required |
| `GET` | `/payments/:id` | Get payment by ID | — |
| `POST` | `/products` | Create a product | — |
| `GET` | `/products/:id` | Get product by ID | — |
| `DELETE` | `/products/:id` | Delete a product | — |
| `GET` | `/health` | Health check | — |
 
---
 
### Create Payment
 
```
POST /payments
```
 
**Headers**
 
| Header | Required | Description |
|---|---|---|
| `Content-Type` | Yes | `application/json` |
| `Idempotency-Key` | Yes | Client-generated unique key (UUID recommended) |
 
**Request Body**
 
```json
{
  "product_id": "uuid",
  "amount":     100,
  "currency":   "USD"
}
```
 
**Responses**
 
| Status | Description |
|---|---|
| `201 Created` | Payment successfully created |
| `200 OK` | Duplicate request — returns original stored response |
| `409 Conflict` | Key exists with a different request hash, or payment is still `in_progress` |
| `422 Unprocessable Entity` | Validation error |
 
---
 
## Request & Response Examples
 
### First Payment Request (new key)
 
```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: a1b2c3d4-0000-0000-0000-000000000001" \
  -d '{"product_id": "prod-uuid", "amount": 100, "currency": "USD"}'
```
 
```json
HTTP/1.1 201 Created
 
{
  "id":         "pay-uuid",
  "product_id": "prod-uuid",
  "amount":     100,
  "currency":   "USD",
  "status":     "success",
  "created_at": "2025-01-01T12:00:00Z"
}
```
 
### Duplicate Request (same key, same body)
 
```bash
# Exact same request — safe retry
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: a1b2c3d4-0000-0000-0000-000000000001" \
  -d '{"product_id": "prod-uuid", "amount": 100, "currency": "USD"}'
```
 
```json
HTTP/1.1 200 OK
 
{
  "id":         "pay-uuid",
  "product_id": "prod-uuid",
  "amount":     100,
  "currency":   "USD",
  "status":     "success",
  "created_at": "2025-01-01T12:00:00Z"
}
```
 
### Conflicting Request (same key, different body)
 
```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: a1b2c3d4-0000-0000-0000-000000000001" \
  -d '{"product_id": "prod-uuid", "amount": 999, "currency": "USD"}'
```
 
```json
HTTP/1.1 409 Conflict
 
{
  "error": "idempotency key already used with a different request"
}
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+

### Setup

```bash
# Clone the repository
git clone https://github.com/your-org/idempotent-payment.git
cd idempotent-payment

# Copy and configure environment variables
cp .env.example .env

# Run database migrations
make migrate

# Start the server
go run ./cmd/server
```

### Environment Variables

| Variable | Description | Example |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/payments` |
| `SERVER_PORT` | HTTP server port | `8080` |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `error`) | `info` |

---

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector (recommended for concurrency correctness)
go test -race ./...

# Run only HTTP integration tests
go test ./http/...
```

The test suite covers:

- **Unit tests** — idempotency hash logic, payment and product service functions
- **HTTP integration tests** — full request lifecycle including duplicate detection, conflict handling, and concurrent retry scenarios

---

## Observability

The `logger` module emits structured JSON logs at every significant step of the request lifecycle:

```json
{"level":"info",  "msg":"payment created",    "payment_id":"...", "idempotency_key":"..."}
{"level":"info",  "msg":"duplicate request",   "idempotency_key":"...", "status":"completed"}
{"level":"warn",  "msg":"hash mismatch",        "idempotency_key":"..."}
{"level":"error", "msg":"transaction failed",   "error":"..."}
```

Log fields are consistent across modules, making them easy to aggregate in tools like Datadog, Loki, or any JSON-aware log pipeline.