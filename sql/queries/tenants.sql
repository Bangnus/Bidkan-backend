-- name: CreateTenant :exec
INSERT INTO tenants (
    id, organization_id, slug, name, platform_fee_percent, auto_tiered_gp, unlock_fee, per_minute_rate,
    parking_penalty_fee, sim_fee_per_bike, contact_name, contact_email, contact_phone, notes, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()
);

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: GetTenantByStripeAccountID :one
SELECT * FROM tenants WHERE stripe_account_id = $1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC;

-- name: ListTenantsByOrganization :many
SELECT * FROM tenants WHERE organization_id = $1 ORDER BY name ASC;

-- name: UpdateTenantRates :exec
UPDATE tenants
SET unlock_fee = $2,
    per_minute_rate = $3,
    platform_fee_percent = $4,
    auto_tiered_gp = $5,
    parking_penalty_fee = $6,
    sim_fee_per_bike = $7,
    updated_at = NOW()
WHERE id = $1;

-- name: CountBikesByTenant :one
SELECT COUNT(*)::INT FROM bikes WHERE tenant_id = $1;

-- name: CountBikesByOrganization :one
SELECT COUNT(*)::INT
FROM bikes b
JOIN tenants t ON b.tenant_id = t.id
WHERE t.organization_id = $1;

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

-- name: AssignTenantOrganization :exec
UPDATE tenants
SET organization_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CountTenants :one
SELECT COUNT(*)::INT FROM tenants;

-- name: ListTenantsPaged :many
SELECT * FROM tenants
ORDER BY CASE WHEN status = 'pending' THEN 0 ELSE 1 END, created_at DESC
LIMIT $1 OFFSET $2;
