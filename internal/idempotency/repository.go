package idempotency

import "context"

type IdempotencyRepository interface {
	GetByKey(ctx context.Context, key string) (*IdempotencyRecord, error)
	CreateIfNotExists(ctx context.Context, record *IdempotencyRecord) (*IdempotencyRecord, bool, error)
	Save(ctx context.Context, record *IdempotencyRecord) error
}
