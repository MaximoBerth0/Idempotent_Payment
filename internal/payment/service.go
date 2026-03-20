package payment

import (
	"context"
	"log/slog"
)

type Service struct {
	repo PaymentRepository
	log  *slog.Logger
}

func NewService(repo PaymentRepository, logger *slog.Logger) *Service {
	return &Service{repo: repo, log: logger}
}

func (s *Service) Create(
	ctx context.Context,
	amount int64,
	id string,
) (*Payment, error) {

	s.log.Info("creating payment",
		"id", id,
		"amount", amount,
	)

	payment, err := NewPayment(id, amount)
	if err != nil {
		return nil, err
	}

	// Simulation of payment processing logic
	if amount > 100000 {
		payment.MarkFailed()

		s.log.Warn("payment failed by rule",
			"id", id,
			"amount", amount,
		)

	} else {
		payment.MarkSuccess()

		s.log.Info("payment processed successfully",
			"id", id,
			"amount", amount,
		)
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	s.log.Info("payment stored", "id", payment.ID, "status", payment.Status)

	return payment, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (*Payment, error) {

	s.log.Info("fetching payment",
		"id", id,
	)

	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {

		s.log.Error("failed to fetch payment",
			"id", id,
			"error", err,
		)

		return nil, err
	}
	return payment, nil
}

func (s *Service) Health(ctx context.Context) error {

	err := s.repo.Health(ctx)

	if err != nil {
		s.log.Error("health check failed", "error", err)
		return err
	}

	return nil
}
