-- name: CreateUser :exec
INSERT INTO users (id, phone_number, username, password, wallet_balance, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CreateTenantUser :exec
INSERT INTO users (id, tenant_id, phone_number, username, password, wallet_balance, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUserBalance :exec
UPDATE users SET wallet_balance = $1, updated_at = NOW() WHERE id = $2;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsersByTenant :many
SELECT * FROM users WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: ListStaffByTenant :many
SELECT * FROM users 
WHERE tenant_id = $1 AND role = 'tenant_staff'
ORDER BY created_at DESC;

-- name: AssignUserToTenant :exec
UPDATE users 
SET tenant_id = $2, role = $3, updated_at = NOW()
WHERE id = $1;

-- name: AddBalance :exec
UPDATE users SET wallet_balance = wallet_balance + $1, updated_at = NOW() WHERE id = $2;

-- name: DeductBalance :exec
UPDATE users SET wallet_balance = wallet_balance - $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUserProfileImage :exec
UPDATE users SET image_url = $1, updated_at = NOW() WHERE id = $2;
