package idempotency

import "time"

type IdempotencyStatus string

const (
	StatusInProgress IdempotencyStatus = "in_progress"
	StatusCompleted  IdempotencyStatus = "completed"
	StatusFailed     IdempotencyStatus = "failed"
)

type IdempotencyRecord struct {
	Key         string
	Status      IdempotencyStatus
	HTTPStatus  int
	Response    []byte
	RequestHash string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewIdempotencyRecord(key string, requestHash string) *IdempotencyRecord {
	return &IdempotencyRecord{
		Key:         key,
		RequestHash: requestHash,
		Status:      StatusInProgress,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (r *IdempotencyRecord) ValidateHash(hash string) error {
	if r.RequestHash != hash {
		return ErrHashMismatch
	}
	return nil
}

func (r *IdempotencyRecord) MarkCompleted(response []byte, httpStatus int) {
	r.Status = StatusCompleted
	r.Response = response
	r.HTTPStatus = httpStatus
	r.UpdatedAt = time.Now()
}

func (r *IdempotencyRecord) MarkFailed(response []byte, httpStatus int) {
	r.Status = StatusFailed
	r.Response = response
	r.HTTPStatus = httpStatus
	r.UpdatedAt = time.Now()
}
