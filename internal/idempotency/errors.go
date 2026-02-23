package idempotency

import "errors"

var (
	ErrHashMismatch = errors.New("request hash does not match idempotency key")
)
