package products

import (
	"errors"
	"time"
)

type Product struct {
	ID        int
	Name      string
	Price     int
	Active    bool
	CreatedAt time.Time
}

func NewProduct(id int, name string, price int) (*Product, error) {
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	return &Product{
		ID:        id,
		Name:      name,
		Price:     price,
		Active:    true,
		CreatedAt: time.Now(),
	}, nil
}

func (p *Product) Deactivate() {
	p.Active = false
}
