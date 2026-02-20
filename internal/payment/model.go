package payment

import "time"

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID        string        `db:"id"`
	Amount    int64         `db:"amount"`
	Currency  string        `db:"currency"`
	Status    PaymentStatus `db:"status"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}
