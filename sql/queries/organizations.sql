-- name: CreateOrganization :exec
INSERT INTO organizations (
    id, slug, name, contact_email, contact_phone, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
);

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations ORDER BY created_at DESC;

-- name: UpdateOrganizationStatus :exec
UPDATE organizations
SET status = $2,
    updated_at = NOW()
WHERE id = $1;
