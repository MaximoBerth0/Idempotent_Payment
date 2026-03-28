package payment

import "errors"

var ErrNotFound = errors.New("payment not found")

var ErrInvalidPaymentID = errors.New("invalid payment ID")

var ErrInvalidProductID = errors.New("invalid product ID")

var ErrInactiveProduct = errors.New("product is inactive")
