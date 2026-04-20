package postgres

import "errors"

var ErrProductNotFound = errors.New("product not found in database")

var ErrKeyNotFound = errors.New("key not found in database")
