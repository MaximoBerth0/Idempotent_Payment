package product

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id int) (*Product, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*Product, error)
}
