package repository

import (
	"context"

	"github.com/google/uuid"
)

func (r *SuperAdminPostgresRepository) ActivatePendingAdmins(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) error {
	if orgID != nil {
		_, err := r.db.ExecContext(ctx, `
			UPDATE users SET status = 'active', updated_at = NOW()
			WHERE (tenant_id = $1 OR organization_id = $2)
			  AND role IN ('tenant_admin', 'org_admin')
			  AND status = 'pending'
		`, tenantID, *orgID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET status = 'active', updated_at = NOW()
		WHERE tenant_id = $1
		  AND role IN ('tenant_admin', 'org_admin')
		  AND status = 'pending'
	`, tenantID)
	return err
}
