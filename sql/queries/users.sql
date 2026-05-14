-- name: CreateUser :exec
INSERT INTO users (id, phone_number, full_name, password, wallet_balance, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2;
