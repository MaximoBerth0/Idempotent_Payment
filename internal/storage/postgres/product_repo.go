package postgres

import (
	"context"
	"errors"
	"fmt"
	products "idempotent-payment/internal/product"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductRepository struct {
	db DBTX
}

func NewProductRepository(db DBTX) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *products.Product) error {
	query := `
		INSERT INTO products (name, price, active, currency, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		product.Name,
		product.Price,
		product.Active,
		product.Currency,
		product.CreatedAt,
	)

	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*products.Product, error) {
	query := `
		SELECT id, name, price, active, currency, created_at
		FROM products
		WHERE id = $1
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
			return nil, nil // ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by id %d: %w", id, err)
	}

	product, err := products.NewProduct(pID, name, price, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid product data from db (id %d): %w", id, err)
	}

	product.Active = active
	product.CreatedAt = createdAt

	return product, nil
}
