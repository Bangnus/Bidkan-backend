-- name: CreateTenant :exec
INSERT INTO tenants (
    id, slug, name, platform_fee_percent, unlock_fee, per_minute_rate, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
);

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: GetTenantByStripeAccountID :one
SELECT * FROM tenants WHERE stripe_account_id = $1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC;

-- name: UpdateTenantRates :exec
UPDATE tenants
SET unlock_fee = $2,
    per_minute_rate = $3,
    platform_fee_percent = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateTenantStripeAccount :exec
UPDATE tenants
SET stripe_account_id = $2,
    stripe_onboarding_completed = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateTenantStatus :exec
UPDATE tenants
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

