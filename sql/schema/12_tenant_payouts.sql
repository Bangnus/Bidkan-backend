CREATE TABLE tenant_payouts (
    id                  UUID PRIMARY KEY,
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    stripe_transfer_id  VARCHAR(100),
    total_rides         INT NOT NULL DEFAULT 0,
    gross_amount        DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    platform_fee_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    net_payout_amount   DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message       TEXT,
    period_start        TIMESTAMP NOT NULL,
    period_end          TIMESTAMP NOT NULL,
    processed_at        TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

