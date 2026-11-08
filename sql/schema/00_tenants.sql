CREATE TABLE tenants (
    id                          UUID PRIMARY KEY,
    slug                        VARCHAR(50) UNIQUE NOT NULL,
    name                        VARCHAR(255) NOT NULL,
    stripe_account_id           VARCHAR(100) UNIQUE,
    stripe_onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    platform_fee_percent        DECIMAL(5,2) NOT NULL DEFAULT 15.00,
    unlock_fee                  DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    per_minute_rate             DECIMAL(10,2) NOT NULL DEFAULT 1.00,
    status                      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT NOW()
);

