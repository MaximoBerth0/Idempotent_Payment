package product

import (
	"context"
	"log/slog"
)

type Service struct {
	repo ProductRepository
	log  *slog.Logger
}

func NewService(repo ProductRepository, logger *slog.Logger) *Service {
	return &Service{repo: repo, log: logger}
}

func (s *Service) Create(ctx context.Context, name string, price int, currency string) (*Product, error) {
	s.log.Info("creating product", "name", name)

	product, err := NewProduct(name, price, currency)
	if err != nil {
		s.log.Error("invalid product data", "error", err)
		return nil, err
	}

	if err := s.repo.Create(ctx, product); err != nil {
		s.log.Error("failed to save product", "error", err)
		return nil, err
	}

	s.log.Info("product created successfully", "product_id", product.ID)
	return product, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Product, error) {
	s.log.Info("fetching product", "product_id", id)

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to fetch product", "product_id", id, "error", err)
		return nil, err
	}

	return product, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	s.log.Info("deleting product", "product_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete product", "product_id", id, "error", err)
		return err
	}

	s.log.Info("product deleted successfully", "product_id", id)
	return nil
}
