package postgres

import (
	"context"
	"errors"

	"idempotent-payment/internal/idempotency"

	"github.com/jackc/pgx/v5"
)

type IdempotencyRepository struct {
	db DBTX
}

func NewIdempotencyRepository(db DBTX) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) GetByKey(
	ctx context.Context,
	key string,
) (*idempotency.IdempotencyRecord, error) {

	query := `
		SELECT key, status, http_status, response, request_hash, created_at, updated_at
		FROM idempotency
		WHERE key = $1
	`

	row := r.db.QueryRow(ctx, query, key)

	var record idempotency.IdempotencyRecord

	err := row.Scan(
		&record.Key,
		&record.Status,
		&record.HTTPStatus,
		&record.Response,
		&record.RequestHash,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}

func (r *IdempotencyRepository) CreateIfNotExists(
	ctx context.Context,
	record *idempotency.IdempotencyRecord,
) (*idempotency.IdempotencyRecord, bool, error) {

	query := `
		INSERT INTO idempotency 
		(key, status, http_status, response, request_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO NOTHING
		RETURNING key, status, http_status, response, request_hash, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		record.Key,
		record.Status,
		record.HTTPStatus,
		record.Response,
		record.RequestHash,
		record.CreatedAt,
		record.UpdatedAt,
	)

	var created idempotency.IdempotencyRecord

	err := row.Scan(
		&created.Key,
		&created.Status,
		&created.HTTPStatus,
		&created.Response,
		&created.RequestHash,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, err := r.GetByKey(ctx, record.Key)
			return existing, false, err
		}
		return nil, false, err
	}

	return &created, true, nil
}

func (r *IdempotencyRepository) Save(
	ctx context.Context,
	record *idempotency.IdempotencyRecord,
) error {

	query := `
		UPDATE idempotency
		SET status = $2,
		    http_status = $3,
		    response = $4,
		    updated_at = $5
		WHERE key = $1
	`

	_, err := r.db.Exec(ctx, query,
		record.Key,
		record.Status,
		record.HTTPStatus,
		record.Response,
		record.UpdatedAt,
	)

	return err
}
