-- name: CreateHardwareOrder :exec
INSERT INTO hardware_orders (
    id, tenant_id, order_type, quantity, unit_price, total_amount, note, status, created_by, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
);

-- name: ListHardwareOrdersByTenant :many
SELECT o.*, t.slug AS tenant_slug, t.name AS tenant_name
FROM hardware_orders o
JOIN tenants t ON o.tenant_id = t.id
WHERE o.tenant_id = $1
ORDER BY o.created_at DESC;

-- name: ListAllHardwareOrders :many
SELECT o.*, t.slug AS tenant_slug, t.name AS tenant_name
FROM hardware_orders o
JOIN tenants t ON o.tenant_id = t.id
ORDER BY o.created_at DESC;

-- name: UpdateHardwareOrderStatus :exec
UPDATE hardware_orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CountAllHardwareOrders :one
SELECT COUNT(*)::INT FROM hardware_orders;

-- name: ListAllHardwareOrdersPaged :many
SELECT o.*, t.slug AS tenant_slug, t.name AS tenant_name
FROM hardware_orders o
JOIN tenants t ON o.tenant_id = t.id
ORDER BY o.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountHardwareOrdersByTenant :one
SELECT COUNT(*)::INT FROM hardware_orders WHERE tenant_id = $1;

-- name: ListHardwareOrdersByTenantPaged :many
SELECT o.*, t.slug AS tenant_slug, t.name AS tenant_name
FROM hardware_orders o
JOIN tenants t ON o.tenant_id = t.id
WHERE o.tenant_id = $1
ORDER BY o.created_at DESC
LIMIT $2 OFFSET $3;
