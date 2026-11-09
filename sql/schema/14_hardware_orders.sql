CREATE TABLE hardware_orders (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_type   VARCHAR(50) NOT NULL,
    quantity     INT NOT NULL DEFAULT 1,
    unit_price   DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    note         TEXT,
    status       VARCHAR(30) NOT NULL DEFAULT 'pending',
    created_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
