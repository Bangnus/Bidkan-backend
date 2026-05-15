-- name: CreateTransaction :exec
INSERT INTO transactions (id, user_id, amount, type, status, reference_id, gateway_ref, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW());

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status = $2, gateway_ref = COALESCE($3, gateway_ref), updated_at = NOW()
WHERE id = $1;

-- name: GetTotalSpendingInWindow :one
SELECT COALESCE(SUM(ABS(amount)), 0)::TEXT
FROM transactions
WHERE user_id = $1 
  AND type = 'fare_deduction' 
  AND status = 'completed'
  AND created_at >= $2;
