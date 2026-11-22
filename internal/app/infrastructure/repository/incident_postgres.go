package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type IncidentPostgres struct {
	db *sql.DB
}

func NewIncidentPostgres(db *sql.DB) repository.IncidentRepository {
	return &IncidentPostgres{db: db}
}

func (r *IncidentPostgres) ListByTenant(ctx context.Context, tenantID uuid.UUID, status, severity string) ([]*entity.IncidentReport, error) {
	query := `
		SELECT id, tenant_id, ride_id, user_id, reporter_name, reporter_phone, incident_type,
		       severity, description, COALESCE(location_name, '-'), location_lat, location_lon,
		       COALESCE(police_report_no, '-'), COALESCE(insurance_claim_no, '-'), insurance_status,
		       claim_amount, approved_amount, COALESCE(hospital_name, '-'), COALESCE(notes, '-'),
		       created_at, resolved_at
		FROM incident_reports
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	idx := 2

	if status != "" && status != "all" {
		query += fmtQueryParam(&idx, " AND insurance_status = ")
		args = append(args, status)
	}
	if severity != "" && severity != "all" {
		query += fmtQueryParam(&idx, " AND severity = ")
		args = append(args, severity)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.IncidentReport
	for rows.Next() {
		var i entity.IncidentReport
		var rideID, userID *uuid.UUID
		var resolvedAt *time.Time
		if err := rows.Scan(
			&i.ID, &i.TenantID, &rideID, &userID, &i.ReporterName, &i.ReporterPhone, &i.IncidentType,
			&i.Severity, &i.Description, &i.LocationName, &i.LocationLat, &i.LocationLon,
			&i.PoliceReportNo, &i.InsuranceClaimNo, &i.InsuranceStatus,
			&i.ClaimAmount, &i.ApprovedAmount, &i.HospitalName, &i.Notes,
			&i.CreatedAt, &resolvedAt,
		); err != nil {
			return nil, err
		}
		i.RideID = rideID
		i.UserID = userID
		i.ResolvedAt = resolvedAt
		list = append(list, &i)
	}
	return list, nil
}

func (r *IncidentPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.IncidentReport, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, ride_id, user_id, reporter_name, reporter_phone, incident_type,
		       severity, description, COALESCE(location_name, '-'), location_lat, location_lon,
		       COALESCE(police_report_no, '-'), COALESCE(insurance_claim_no, '-'), insurance_status,
		       claim_amount, approved_amount, COALESCE(hospital_name, '-'), COALESCE(notes, '-'),
		       created_at, resolved_at
		FROM incident_reports WHERE id = $1`, id)

	var i entity.IncidentReport
	var rideID, userID *uuid.UUID
	var resolvedAt *time.Time
	if err := row.Scan(
		&i.ID, &i.TenantID, &rideID, &userID, &i.ReporterName, &i.ReporterPhone, &i.IncidentType,
		&i.Severity, &i.Description, &i.LocationName, &i.LocationLat, &i.LocationLon,
		&i.PoliceReportNo, &i.InsuranceClaimNo, &i.InsuranceStatus,
		&i.ClaimAmount, &i.ApprovedAmount, &i.HospitalName, &i.Notes,
		&i.CreatedAt, &resolvedAt,
	); err != nil {
		return nil, err
	}
	i.RideID = rideID
	i.UserID = userID
	i.ResolvedAt = resolvedAt
	return &i, nil
}

func (r *IncidentPostgres) Create(ctx context.Context, inc *entity.IncidentReport) error {
	if inc.ID == uuid.Nil {
		inc.ID = uuid.New()
	}
	query := `
		INSERT INTO incident_reports (
			id, tenant_id, ride_id, user_id, reporter_name, reporter_phone,
			incident_type, severity, description, location_name, location_lat,
			location_lon, police_report_no, insurance_claim_no, insurance_status,
			claim_amount, approved_amount, hospital_name, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW()
		)`
	_, err := r.db.ExecContext(ctx, query,
		inc.ID, inc.TenantID, inc.RideID, inc.UserID, inc.ReporterName, inc.ReporterPhone,
		inc.IncidentType, inc.Severity, inc.Description, inc.LocationName, inc.LocationLat,
		inc.LocationLon, inc.PoliceReportNo, inc.InsuranceClaimNo, inc.InsuranceStatus,
		inc.ClaimAmount, inc.ApprovedAmount, inc.HospitalName, inc.Notes,
	)
	return err
}

func (r *IncidentPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status string, approvedAmt float64, claimNo, notes string) (*entity.IncidentReport, error) {
	var resolvedAt *time.Time
	if status == "approved" || status == "settled" || status == "rejected" {
		now := time.Now()
		resolvedAt = &now
	}

	query := `
		UPDATE incident_reports
		SET insurance_status = $1, approved_amount = $2, insurance_claim_no = $3, notes = $4, resolved_at = $5
		WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, status, approvedAmt, claimNo, notes, resolvedAt, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func fmtQueryParam(idx *int, prefix string) string {
	res := prefix + fmtPlaceholder(*idx)
	*idx++
	return res
}

func fmtPlaceholder(idx int) string {
	return "$" + string(rune('0'+idx))
}
