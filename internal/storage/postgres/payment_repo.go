package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"idempotent-payment/internal/payment"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type PaymentRepository struct {
	db   DBTX
	pool *pgxpool.Pool
	log  *slog.Logger
}

var paymentTracer = otel.Tracer("idempotent-payment/internal/storage/postgres/payment_repo")

func NewPaymentRepository(db DBTX, logger *slog.Logger, pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db, log: logger, pool: pool}
}

func (r *PaymentRepository) conn(ctx context.Context) DBTX {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *PaymentRepository) Create(
	ctx context.Context,
	p *payment.Payment,
) error {
	ctx, span := paymentTracer.Start(ctx, "PaymentRepository.Create")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "payments"),
	)

	query := `
        INSERT INTO payments (id, product_id, amount, currency, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(ctx, query,
		p.ID,
		p.ProductID,
		p.Amount,
		p.Currency,
		p.Status,
		p.CreatedAt,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to insert payment", "error", err, "payment_id", p.ID)
		return err
	}

	return nil
}

func (r *PaymentRepository) GetByID(
	ctx context.Context,
	id string,
) (*payment.Payment, error) {

	ctx, span := paymentTracer.Start(ctx, "PaymentRepository.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "payments"),
		attribute.String("payment.id", id),
	)

	query := `
		SELECT id, product_id, amount, currency, status, created_at
		FROM payments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var p payment.Payment

	err := row.Scan(
		&p.ID,
		&p.ProductID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payment.ErrNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to scan payment", "error", err, "payment_id", id)
		return nil, fmt.Errorf("get payment by id: %w", err)
	}

	return &p, nil
}

func (r *PaymentRepository) Save(
	ctx context.Context,
	p *payment.Payment,
) error {
	ctx, span := paymentTracer.Start(ctx, "PaymentRepository.Save")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "payments"),
		attribute.String("payment.id", p.ID),
	)

	query := `
		UPDATE payments
		SET amount = $2, status = $3
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query,
		p.ID,
		p.Amount,
		p.Status,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to update payment", "error", err, "payment_id", p.ID)
		return fmt.Errorf("save payment: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return payment.ErrNotFound
	}

	return nil
}

func (r *PaymentRepository) Health(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
