-- name: GetConfig :one
SELECT value FROM configs WHERE key = $1;

-- name: SetConfig :exec
INSERT INTO configs (key, value, updated_at) 
VALUES ($1, $2, NOW()) 
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
