package http

import (
	"context"
	"idempotent-payment/internal/product"
	"log/slog"
)

type ProductService interface {
	Create(ctx context.Context, name string, price int, currency string) (*product.Product, error)
	GetByID(ctx context.Context, id int64) (*product.Product, error)
	Delete(ctx context.Context, id int64) error
}

type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
}

func NewProductHandler(s ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{service: s, logger: logger}
}
