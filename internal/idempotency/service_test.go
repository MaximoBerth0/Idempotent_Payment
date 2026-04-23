package idempotency

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

// noopTx runs fn directly without a real database transaction (like unit-test helper)
func noopTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type MockRepo struct {
	record  *IdempotencyRecord
	logger  *slog.Logger
	created bool
	err     error
}

func (m *MockRepo) GetByKey(ctx context.Context, key string) (*IdempotencyRecord, error) {
	return m.record, nil
}

func (m *MockRepo) CreateIfNotExists(ctx context.Context, r *IdempotencyRecord) (*IdempotencyRecord, bool, error) {
	return m.record, m.created, m.err
}

func (m *MockRepo) Save(ctx context.Context, r *IdempotencyRecord) error {
	m.record = r
	return nil
}

// Tests the flow of a new request:
// → handler runs
// → response stored
// → returned to user

func TestExecute_NewRequest(t *testing.T) {

	ctx := context.Background()

	mockRepo := &MockRepo{
		created: true,
	}

	service := NewService(mockRepo, slog.New(slog.NewTextHandler(io.Discard, nil)), noopTx)

	handler := func(ctx context.Context) ([]byte, int, error) {
		return []byte("payment ok"), 200, nil
	}

	resp, status, err := service.Execute(
		ctx,
		"key123",
		"hash123",
		handler,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != 200 {
		t.Fatalf("expected 200 got %d", status)
	}

	if string(resp) != "payment ok" {
		t.Fatalf("unexpected response")
	}
}

// tests the flow of a duplicate request with same key and hash:
// same key
// same hash
// - do NOT execute handler again
// - return stored response

func TestExecute_DuplicateRequestReturnsStoredResponse(t *testing.T) {

	ctx := context.Background()

	existing := &IdempotencyRecord{
		Key:         "key123",
		RequestHash: "hash123",
		Status:      StatusCompleted,
		Response:    []byte("cached response"),
		HTTPStatus:  200,
	}

	mockRepo := &MockRepo{
		record:  existing,
		created: false,
	}

	service := NewService(mockRepo, slog.New(slog.NewTextHandler(io.Discard, nil)), noopTx)

	handler := func(ctx context.Context) ([]byte, int, error) {
		t.Fatal("handler should not run")
		return nil, 0, nil
	}

	resp, status, err := service.Execute(
		ctx,
		"key123",
		"hash123",
		handler,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(resp) != "cached response" {
		t.Fatalf("expected cached response")
	}

	if status != 200 {
		t.Fatalf("expected status 200")
	}
}
