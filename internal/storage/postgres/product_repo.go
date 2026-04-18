package postgres

import (
	"context"
	"errors"
	"fmt"
	"idempotent-payment/internal/product"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var productTracer = otel.Tracer("idempotent-payment/internal/storage/postgres/product_repo")

type ProductRepository struct {
	db  DBTX
	log *slog.Logger
}

func NewProductRepository(db DBTX, logger *slog.Logger) *ProductRepository {
	return &ProductRepository{db: db, log: logger}
}

func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	ctx, span := productTracer.Start(ctx, "ProductRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "products"),
		attribute.String("product.name", p.Name),
	)

	query := `
        INSERT INTO products (name, price, active, currency, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, p.Name, p.Price, p.Active, p.Currency, p.CreatedAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to insert product", "error", err, "product_name", p.Name)
		return err
	}

	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	ctx, span := productTracer.Start(ctx, "ProductRepository.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "products"),
		attribute.Int64("product.id", id),
	)

	query := `
        SELECT id, name, price, active, currency, created_at
        FROM products WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, id)

	var (
		pID       int
		name      string
		price     int
		active    bool
		currency  string
		createdAt time.Time
	)

	err := row.Scan(&pID, &name, &price, &active, &currency, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound // ← esperado, sin log
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to scan product", "error", err, "product_id", id)
		return nil, fmt.Errorf("get product by id %d: %w", id, err)
	}

	p, err := product.NewProduct(name, price, currency)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "invalid product data in db", "error", err, "product_id", id)
		return nil, fmt.Errorf("invalid product data from db (id %d): %w", id, err)
	}

	p.Active = active
	p.CreatedAt = createdAt
	return p, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := productTracer.Start(ctx, "ProductRepository.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "DELETE"),
		attribute.String("db.table", "products"),
		attribute.Int64("product.id", id),
	)

	cmdTag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.log.ErrorContext(ctx, "failed to delete product", "error", err, "product_id", id)
		return fmt.Errorf("delete product %d: %w", id, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrProductNotFound // ← esperado, sin log
	}

	return nil
}
