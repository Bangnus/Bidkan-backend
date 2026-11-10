-- name: CreateUser :exec
INSERT INTO users (id, phone_number, username, password, wallet_balance, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CreateTenantUser :exec
INSERT INTO users (id, tenant_id, organization_id, phone_number, email, username, password, wallet_balance, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: ListTenantAdminsByTenant :many
SELECT * FROM users
WHERE tenant_id = $1 AND role IN ('tenant_admin', 'org_admin')
ORDER BY created_at DESC;

-- name: ListUsersByOrganization :many
SELECT * FROM users
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUserBalance :exec
UPDATE users SET wallet_balance = $1, updated_at = NOW() WHERE id = $2;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByIdentifier :one
SELECT * FROM users WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($1) LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsersByTenant :many
SELECT * FROM users WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: ListStaffByTenant :many
SELECT * FROM users 
WHERE tenant_id = $1 AND role = 'tenant_staff'
ORDER BY created_at DESC;

-- name: CountStaffByTenant :one
SELECT COUNT(*)::INT FROM users
WHERE tenant_id = $1 AND role = 'tenant_staff';

-- name: ListStaffByTenantPaged :many
SELECT * FROM users
WHERE tenant_id = $1 AND role = 'tenant_staff'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

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

-- name: GetUserByEmail :one
SELECT * FROM users WHERE LOWER(email) = LOWER($1) LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1 LIMIT 1;

-- name: UpsertGoogleUser :one
INSERT INTO users (id, tenant_id, google_id, email, username, phone_number, password, wallet_balance, role, status, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, 'google_oauth', 0.00, 'user', 'active', $7, NOW(), NOW())
ON CONFLICT (email) DO UPDATE SET
    google_id = COALESCE(users.google_id, EXCLUDED.google_id),
    username = EXCLUDED.username,
    image_url = COALESCE(EXCLUDED.image_url, users.image_url),
    tenant_id = COALESCE(users.tenant_id, EXCLUDED.tenant_id),
    updated_at = NOW()
RETURNING *;
