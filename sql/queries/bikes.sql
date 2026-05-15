-- name: UpdateBikeStatus :exec
UPDATE bikes
SET lat = $2,
    lon = $3,
    battery_level = $4,
    last_heartbeat = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CreateBike :exec
INSERT INTO bikes (id, hardware_id, status, image_url)
VALUES ($1, $2, $3, $4);

-- name: GetBike :one
SELECT * FROM bikes WHERE id = $1;

-- name: ListAllBikes :many
SELECT * FROM bikes ORDER BY id;

-- name: ListAvailableBikes :many
SELECT * FROM bikes 
WHERE status = 'available' 
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
