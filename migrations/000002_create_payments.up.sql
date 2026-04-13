CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    amount INT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);