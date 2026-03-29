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
	productID int64,
	id string,
) (*Payment, error) {

	log := s.log.With(
		"id", id,
		"product_id", productID,
	)

	log.Info("creating payment")

	if id == "" {
		return nil, ErrInvalidPaymentID
	}

	if productID <= 0 {
		return nil, ErrInvalidProductID
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		log.Error("failed to fetch product", "error", err)
		return nil, err
	}

	if !product.Active {
		log.Error("product is inactive", "product_id", productID)
		return nil, ErrInactiveProduct
	}

	payment, err := NewPayment(
		id,
		product.ID,
		product.Price,
		product.Currency,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		log.Error("failed to save payment", "error", err)
		return nil, err
	}

	log.Info("payment created successfully")
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
