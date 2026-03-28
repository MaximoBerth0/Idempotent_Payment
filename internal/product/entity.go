package product

import (
	"errors"
	"time"
)

type Product struct {
	ID        int64
	Name      string
	Price     int
	Active    bool
	Currency  string
	CreatedAt time.Time
}

func NewProduct(id int, name string, price int, currency string) (*Product, error) {
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if currency == "" {
		return nil, errors.New("currency cannot be empty")
	}

	return &Product{
		ID:        int64(id),
		Name:      name,
		Price:     price,
		Currency:  currency,
		Active:    true,
		CreatedAt: time.Now(),
	}, nil
}

func (p *Product) Deactivate() {
	p.Active = false
}
