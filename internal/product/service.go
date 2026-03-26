package product

import (
	"context"
)

type Service struct {
	repo ProductRepository
}

func NewService(repo ProductRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, id int, name string, price int, currency string) (*Product, error) {
	product, err := NewProduct(id, name, price, currency)
	if err != nil {
		return nil, err
	}

	err = s.repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Product, error) {
	return s.repo.List(ctx)
}
