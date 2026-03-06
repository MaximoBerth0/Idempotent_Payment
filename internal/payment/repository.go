package payment

import "context"

type PaymentRepository interface {
	Health(ctx context.Context) error
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	Save(ctx context.Context, payment *Payment) error
}
