-- name: UpdateBikeStatus :exec
UPDATE bikes
SET lat = $2,
    lon = $3,
    battery_level = $4,
    last_heartbeat = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CreateBike :exec
INSERT INTO bikes (id, hardware_id, status)
VALUES ($1, $2, $3);

-- name: GetBike :one
SELECT * FROM bikes WHERE id = $1;

-- name: ListAllBikes :many
SELECT * FROM bikes ORDER BY id;

-- name: CreateBikeLocation :exec
INSERT INTO bike_locations (id, bike_id, lat, lon, battery, created_at)
VALUES ($1, $2, $3, $4, $5, $6);
