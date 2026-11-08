-- name: CreateTenantPayout :exec
INSERT INTO tenant_payouts (
    id, tenant_id, stripe_transfer_id, total_rides, gross_amount, platform_fee_amount, net_payout_amount, status, period_start, period_end, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW()
);

-- name: GetTenantPayoutByID :one
SELECT * FROM tenant_payouts WHERE id = $1;

-- name: ListPayoutsByTenant :many
SELECT * FROM tenant_payouts
WHERE tenant_id = $1
ORDER BY period_end DESC;

-- name: ListAllPayouts :many
SELECT * FROM tenant_payouts
ORDER BY period_end DESC;

-- name: UpdatePayoutStatus :exec
UPDATE tenant_payouts
SET status = $2,
    stripe_transfer_id = $3,
    error_message = $4,
    processed_at = NOW()
WHERE id = $1;

