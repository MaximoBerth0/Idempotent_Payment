package idempotency

import "time"

type IdempotencyStatus string

const (
	StatusInProgress IdempotencyStatus = "in_progress"
	StatusCompleted  IdempotencyStatus = "completed"
	StatusFailed     IdempotencyStatus = "failed"
)

type IdempotencyRecord struct {
	Key         string            `db:"key"`          // UNIQUE (idealmente junto a tenant_id)
	TenantID    string            `db:"tenant_id"`    // opcional pero recomendado si es multi-tenant
	RequestHash string            `db:"request_hash"` // hash normalizado del body
	Status      IdempotencyStatus `db:"status"`

	ResourceType string `db:"resource_type"` // "payment", "refund", etc
	ResourceID   string `db:"resource_id"`

	ResponseStatus int    `db:"response_status"` // HTTP status code
	ResponseBody   []byte `db:"response_body"`   // JSON serializado

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	ExpiresAt *time.Time `db:"expires_at"` // nil si no expira
}
