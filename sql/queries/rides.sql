-- name: CreateRide :exec
INSERT INTO rides (
    id, user_id, bike_id, tenant_id, start_time, start_lat, start_lon, unlock_fee, rate_per_minute, status, type, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), $5, $6, $7, $8, 'ongoing', $9, NOW(), NOW()
);

-- name: GetRide :one
SELECT * FROM rides WHERE id = $1;

-- name: EndRide :exec
UPDATE rides
SET end_time = NOW(),
    end_lat = $2,
    end_lon = $3,
    distance_km = $4,
    total_fare = $5,
    platform_fee = $6,
    partner_earnings = $7,
    status = 'completed',
    updated_at = NOW()
WHERE id = $1;

-- name: ListRidesByUser :many
SELECT * FROM rides WHERE user_id = $1 ORDER BY start_time DESC;

-- name: ListRidesByTenant :many
SELECT * FROM rides WHERE tenant_id = $1 ORDER BY start_time DESC;

-- name: ListRidesByTenantWithUser :many
SELECT r.id, r.bike_id, r.user_id, u.username, u.phone_number, r.distance_km,
       r.total_fare, r.platform_fee, r.partner_earnings, r.status, r.start_time, r.end_time
FROM rides r
JOIN users u ON r.user_id = u.id
WHERE r.tenant_id = $1
ORDER BY r.start_time DESC
LIMIT 100;

-- name: GetActiveRideByUser :one
SELECT * FROM rides WHERE user_id = $1 AND status = 'ongoing' LIMIT 1;
