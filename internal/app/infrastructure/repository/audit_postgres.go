package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type AuditPostgres struct {
	db *sql.DB
}

func NewAuditPostgres(db *sql.DB) *AuditPostgres {
	return &AuditPostgres{db: db}
}

func (r *AuditPostgres) LogAction(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	query := `
		INSERT INTO audit_logs (id, tenant_id, actor_id, actor_name, actor_role, action, target_type, target_id, details, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.ActorID, log.ActorName, log.ActorRole,
		log.Action, log.TargetType, log.TargetID, log.Details, log.IPAddress,
	)
	return err
}

func (r *AuditPostgres) ListAuditLogs(ctx context.Context, tenantID uuid.UUID, action string, limit int) ([]*entity.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, tenant_id, actor_id, actor_name, actor_role, action, target_type, target_id, COALESCE(details, '-'), COALESCE(ip_address, '-'), created_at
		FROM audit_logs
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	if action != "" && action != "all" {
		args = append(args, action)
		query += fmt.Sprintf(" AND action = $%d", len(args))
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		l := &entity.AuditLog{}
		var actorID *uuid.UUID
		if err := rows.Scan(
			&l.ID, &l.TenantID, &actorID, &l.ActorName, &l.ActorRole,
			&l.Action, &l.TargetType, &l.TargetID, &l.Details, &l.IPAddress, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.ActorID = actorID
		logs = append(logs, l)
	}
	return logs, nil
}
