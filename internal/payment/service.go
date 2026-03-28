package payment

import (
	"context"
	"idempotent-payment/internal/product"
	"log/slog"
)

type Service struct {
	repo        PaymentRepository
	log         *slog.Logger
	productRepo product.ProductRepository
}

func NewService(repo PaymentRepository, logger *slog.Logger, productRepo product.ProductRepository) *Service {
	return &Service{repo: repo, log: logger, productRepo: productRepo}
}

func (s *Service) Create(
	ctx context.Context,
	id string,
	productID int64,
) (*Payment, error) {

	log := s.log.With(
		"id", id,
		"product_id", productID,
	)

	log.Info("creating payment")

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		log.Error("product not found", "error", err)
		return nil, err
	}

	payment, err := NewPayment(id, product.ID)
	if err != nil {
		log.Error("failed to create payment", "error", err)
		return nil, err
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		log.Error("failed to save payment", "error", err)
		return nil, err
	}

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
