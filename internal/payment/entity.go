package payment

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

type Payment struct {
	ID        string
	ProductID int64
	Amount    int
	Currency  string
	Status    Status
	CreatedAt time.Time
}

func NewPayment(id string, productID int64, amount int, currency string) (*Payment, error) {
	if id == "" {
		return nil, errors.New("payment ID cannot be empty")
	}

	return &Payment{
		ID:        id,
		ProductID: productID,
		Amount:    amount,
		Currency:  currency,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}, nil
}

func (p *Payment) MarkSuccess() {
	p.Status = StatusSuccess
}

func (p *Payment) MarkFailed() {
	p.Status = StatusFailed
}
