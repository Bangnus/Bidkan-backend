package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type reportPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewReportPostgresRepository(db *sql.DB) repository.ReportRepository {
	return &reportPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *reportPostgresRepository) Create(ctx context.Context, report *entity.Report) error {
	return r.queries.CreateReport(ctx, sqlc.CreateReportParams{
		ID:         report.ID,
		BikeID:     report.BikeID,
		ReportedBy: report.ReportedBy,
		IssueType:  report.IssueType,
		Status:     report.Status,
	})
}

func (r *reportPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Report, error) {
	res, err := r.queries.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Report{
		ID:         res.ID,
		BikeID:     res.BikeID,
		ReportedBy: res.ReportedBy,
		IssueType:  res.IssueType,
		Status:     res.Status,
		ResolvedBy: res.ResolvedBy,
		CreatedAt:  res.CreatedAt,
		UpdatedAt:  res.UpdatedAt,
	}, nil
}

func (r *reportPostgresRepository) ListByBike(ctx context.Context, bikeID string) ([]entity.Report, error) {
	reports, err := r.queries.ListReportsByBike(ctx, bikeID)
	if err != nil {
		return nil, err
	}
	var res []entity.Report
	for _, report := range reports {
		res = append(res, entity.Report{
			ID:         report.ID,
			BikeID:     report.BikeID,
			ReportedBy: report.ReportedBy,
			IssueType:  report.IssueType,
			Status:     report.Status,
			ResolvedBy: report.ResolvedBy,
			CreatedAt:  report.CreatedAt,
			UpdatedAt:  report.UpdatedAt,
		})
	}
	return res, nil
}

func (r *reportPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, resolvedBy uuid.UUID) error {
	return r.queries.UpdateReportStatus(ctx, sqlc.UpdateReportStatusParams{
		ID:         id,
		Status:     status,
		ResolvedBy: uuid.NullUUID{UUID: resolvedBy, Valid: resolvedBy != uuid.Nil},
	})
}
