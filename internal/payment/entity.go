package payment

import "errors"

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

type Payment struct {
	ID     string
	Amount int64
	Status Status
}

func NewPayment(id string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	return &Payment{
		ID:     id,
		Amount: amount,
		Status: StatusPending,
	}, nil
}

func (p *Payment) MarkSuccess() {
	p.Status = StatusSuccess
}

func (p *Payment) MarkFailed() {
	p.Status = StatusFailed
}
