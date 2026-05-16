-- name: CreateRide :exec
INSERT INTO rides (
    id, user_id, bike_id, start_time, start_lat, start_lon, status, type, created_at, updated_at
) VALUES (
    $1, $2, $3, NOW(), $4, $5, 'ongoing', $6, NOW(), NOW()
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
    status = 'completed',
    updated_at = NOW()
WHERE id = $1;

-- name: ListRidesByUser :many
SELECT * FROM rides WHERE user_id = $1 ORDER BY start_time DESC;

-- name: GetActiveRideByUser :one
SELECT * FROM rides WHERE user_id = $1 AND status = 'ongoing' LIMIT 1;
