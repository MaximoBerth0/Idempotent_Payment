package payment

import "errors"

var (
	ErrInvalidPaymentID = errors.New("payment ID cannot be empty")
	ErrNotFound         = errors.New("payment not found")
	ErrInvalidProductID = errors.New("product ID must be greater than zero")
	ErrInactiveProduct  = errors.New("product is inactive")
)
