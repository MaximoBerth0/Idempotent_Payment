package postgres

import (
	"context"
	"errors"
	"fmt"

	"idempotent-payment/internal/payment"

	"github.com/jackc/pgx/v5"
)

type PaymentRepository struct {
	db DBTX
}

func NewPaymentRepository(db DBTX) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(
	ctx context.Context,
	p *payment.Payment,
) error {

	query := `
		INSERT INTO payments (id, amount, status)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query,
		p.ID,
		p.Amount,
		p.Status,
	)

	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	return nil
}

func (r *PaymentRepository) GetByID(
	ctx context.Context,
	id string,
) (*payment.Payment, error) {

	query := `
		SELECT id, amount, status
		FROM payments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var p payment.Payment

	err := row.Scan(
		&p.ID,
		&p.Amount,
		&p.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payment.ErrNotFound
		}
		return nil, fmt.Errorf("get payment by id: %w", err)
	}

	return &p, nil
}

func (r *PaymentRepository) Save(
	ctx context.Context,
	p *payment.Payment,
) error {

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
		return fmt.Errorf("save payment: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return payment.ErrNotFound
	}

	return nil
}
