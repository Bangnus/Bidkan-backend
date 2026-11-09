-- name: UpdateBikeStatus :exec
UPDATE bikes
SET lat = $2,
    lon = $3,
    battery_level = $4,
    last_heartbeat = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CreateBike :exec
INSERT INTO bikes (id, hardware_id, status, image_url, tenant_id)
VALUES ($1, $2, $3, $4, $5);

-- name: AssignBikeToTenant :exec
UPDATE bikes
SET tenant_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetBike :one
SELECT * FROM bikes WHERE id = $1;

-- name: ListAllBikes :many
SELECT * FROM bikes ORDER BY id;

-- name: ListBikesByTenant :many
SELECT * FROM bikes 
WHERE tenant_id = $1 
ORDER BY id;

-- name: ListAvailableBikes :many
SELECT * FROM bikes 
WHERE status = 'available' 
AND battery_level > 20
ORDER BY id;

-- name: ListAvailableBikesByTenant :many
SELECT * FROM bikes 
WHERE tenant_id = $1
AND status = 'available' 
AND battery_level > 20
ORDER BY id;

-- name: CreateBikeLocation :exec
INSERT INTO bike_locations (id, bike_id, lat, lon, battery, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateBikeRideStatus :exec
UPDATE bikes
SET status = $2,
    current_ride_id = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateBikeOperationalStatus :exec
UPDATE bikes
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListAllBikesWithTenant :many
SELECT b.id, b.hardware_id, b.tenant_id, COALESCE(t.slug, '') AS tenant_slug, COALESCE(t.name, '') AS tenant_name,
       b.status, b.lat, b.lon, b.battery_level, b.last_heartbeat, b.created_at
FROM bikes b
LEFT JOIN tenants t ON b.tenant_id = t.id
ORDER BY b.id;
