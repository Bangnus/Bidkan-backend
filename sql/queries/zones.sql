-- name: CreateZone :exec
INSERT INTO zones (
    id, name, type, boundary, radius, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
);

-- name: GetZone :one
SELECT * FROM zones WHERE id = $1;

-- name: ListZones :many
SELECT * FROM zones ORDER BY name;

-- name: ListZonesByType :many
SELECT * FROM zones WHERE type = $1 ORDER BY name;

-- name: UpdateZone :exec
UPDATE zones
SET name = $2,
    type = $3,
    boundary = $4,
    radius = $5,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteZone :exec
DELETE FROM zones WHERE id = $1;
