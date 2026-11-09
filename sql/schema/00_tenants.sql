CREATE TABLE organizations (
    id            UUID PRIMARY KEY,
    slug          VARCHAR(50) UNIQUE NOT NULL,
    name          VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(30),
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE tenants (
    id                          UUID PRIMARY KEY,
    organization_id             UUID REFERENCES organizations(id) ON DELETE SET NULL,
    slug                        VARCHAR(50) UNIQUE NOT NULL,
    name                        VARCHAR(255) NOT NULL,
    stripe_account_id           VARCHAR(100) UNIQUE,
    stripe_onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    platform_fee_percent        DECIMAL(5,2) NOT NULL DEFAULT 18.00,
    auto_tiered_gp              BOOLEAN NOT NULL DEFAULT true,
    unlock_fee                  DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    per_minute_rate             DECIMAL(10,2) NOT NULL DEFAULT 1.00,
    parking_penalty_fee         DECIMAL(10,2) NOT NULL DEFAULT 50.00,
    sim_fee_per_bike            DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    contact_name                VARCHAR(255),
    contact_email               VARCHAR(255),
    contact_phone               VARCHAR(30),
    notes                       TEXT,
    status                      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT NOW()
);
