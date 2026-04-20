package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"idempotent-payment/internal/idempotency"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type IdempotencyRepository struct {
	db  DBTX
	log *slog.Logger
}

var idempotencyTracer = otel.Tracer("idempotent-payment/internal/storage/postgres/idempotency_repo")

func NewIdempotencyRepository(db DBTX, logger *slog.Logger) *IdempotencyRepository {
	return &IdempotencyRepository{db: db, log: logger}
}

func (r *IdempotencyRepository) GetByKey(
	ctx context.Context,
	key string,
) (*idempotency.IdempotencyRecord, error) {
	ctx, span := idempotencyTracer.Start(ctx, "IdempotencyRepository.GetByKey")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "idempotency"),
		attribute.String("idempotency.key", key),
	)

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to scan idempotency key", "error", err, "key", key)
		return nil, ErrKeyNotFound
	}

	return &record, nil
}

func (r *IdempotencyRepository) CreateIfNotExists(
	ctx context.Context,
	record *idempotency.IdempotencyRecord,
) (*idempotency.IdempotencyRecord, bool, error) {
	ctx, span := idempotencyTracer.Start(ctx, "IdempotencyRepository.CreateIfNotExists")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "idempotency"),
		attribute.String("idempotency.key", record.Key),
	)

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
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				r.log.ErrorContext(ctx, "failed to get existing idempotency record", "error", err, "idempotency_key", record.Key)
				return nil, false, err
			}
			return existing, false, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to scan idempotency record", "error", err, "idempotency_key", record.Key)
		return nil, false, fmt.Errorf("create idempotency record: %w", err)
	}

	return &created, true, nil
}

func (r *IdempotencyRepository) Save(
	ctx context.Context,
	record *idempotency.IdempotencyRecord,
) error {
	ctx, span := idempotencyTracer.Start(ctx, "IdempotencyRepository.Save")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "idempotency"),
		attribute.String("idempotency.key", record.Key),
	)

	query := `
		UPDATE idempotency
		SET status = $2,
		    http_status = $3,
		    response = $4,
		    updated_at = $5
		WHERE key = $1
	`

	cmdTag, err := r.db.Exec(ctx, query,
		record.Key,
		record.Status,
		record.HTTPStatus,
		record.Response,
		record.UpdatedAt,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to update idempotency record", "error", err, "idempotency_key", record.Key)
		return fmt.Errorf("save idempotency record: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return idempotency.ErrNotFound
	}

	return nil
}
