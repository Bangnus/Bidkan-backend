-- name: CreateTransaction :exec
INSERT INTO transactions (
    id, user_id, amount, type, reference_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
);

-- name: GetTransaction :one
SELECT * FROM transactions WHERE id = $1;

-- name: ListTransactionsByUser :many
SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC;
