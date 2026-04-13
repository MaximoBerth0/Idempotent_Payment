CREATE TABLE idempotency (
    key          TEXT        PRIMARY KEY,
    status       TEXT        NOT NULL,
    http_status  INT         NOT NULL DEFAULT 0,
    response     BYTEA,
    request_hash TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
