package payment

import "context"

type Service struct {
	repo PaymentRepository
}

func NewService(repo PaymentRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	amount int64,
	id string,
) (*Payment, error) {

	payment, err := NewPayment(id, amount)
	if err != nil {
		return nil, err
	}

	// Simulation of payment processing logic
	if amount > 100000 {
		payment.MarkFailed()
	} else {
		payment.MarkSuccess()
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}
