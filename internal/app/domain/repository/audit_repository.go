package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type AuditRepository interface {
	LogAction(ctx context.Context, log *entity.AuditLog) error
	ListAuditLogs(ctx context.Context, tenantID uuid.UUID, action string, limit int) ([]*entity.AuditLog, error)
}

