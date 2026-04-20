package idempotency

import "errors"

var (
	ErrHashMismatch = errors.New("request hash does not match idempotency key")
)

var ErrRequestInProgress = errors.New("a request with this idempotency key is already in progress")

var ErrNotFound = errors.New("idempotency not found")
