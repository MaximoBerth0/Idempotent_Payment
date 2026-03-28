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

var validCurrencies = map[string]struct{}{
	"USD": {},
	"ARS": {},
}

func NewProduct(id int, name string, price int, currency string) (*Product, error) {
	if price <= 0 {
		return nil, errors.New("price cannot be negative or zero")
	}

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if _, ok := validCurrencies[currency]; !ok {
		return nil, errors.New("invalid currency")
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

func (p *Product) Deactivate() error {
	if !p.Active {
		return errors.New("product already inactive")
	}
	p.Active = false
	return nil
}
