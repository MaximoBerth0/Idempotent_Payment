package payment

import "errors"

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

type Payment struct {
	ID        string
	ProductID int64
	Status    Status
}

func NewPayment(id string, productID int64) (*Payment, error) {
	if id == "" {
		return nil, errors.New("payment ID cannot be empty")
	}

	return &Payment{
		ID:        id,
		ProductID: productID,
		Status:    StatusPending,
	}, nil
}

func (p *Payment) MarkSuccess() {
	p.Status = StatusSuccess
}

func (p *Payment) MarkFailed() {
	p.Status = StatusFailed
}
