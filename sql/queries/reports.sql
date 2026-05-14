-- name: CreateReport :exec
INSERT INTO reports (id, bike_id, reported_by, issue_type, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW());

-- name: GetReport :one
SELECT * FROM reports WHERE id = $1;

-- name: ListReportsByBike :many
SELECT * FROM reports WHERE bike_id = $1 ORDER BY created_at DESC;

-- name: ListReportsByStatus :many
SELECT * FROM reports WHERE status = $1 ORDER BY created_at DESC;

-- name: UpdateReportStatus :exec
UPDATE reports 
SET status = $2, 
    resolved_by = $3, 
    updated_at = NOW() 
WHERE id = $1;
