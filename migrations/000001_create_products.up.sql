CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price INT NOT NULL,
    active BOOLEAN NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);