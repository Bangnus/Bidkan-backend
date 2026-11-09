package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type OrganizationPostgresRepository struct {
	queries *sqlc.Queries
}

func NewOrganizationPostgresRepository(db *sql.DB) domainRepo.OrganizationRepository {
	return &OrganizationPostgresRepository{queries: sqlc.New(db)}
}

func (r *OrganizationPostgresRepository) Create(ctx context.Context, org *entity.Organization) error {
	return r.queries.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
		ID:           org.ID,
		Slug:         org.Slug,
		Name:         org.Name,
		ContactEmail: toNullString(org.ContactEmail),
		ContactPhone: toNullString(org.ContactPhone),
		Status:       org.Status,
	})
}

func (r *OrganizationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	row, err := r.queries.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapOrg(row), nil
}

func (r *OrganizationPostgresRepository) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	row, err := r.queries.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return mapOrg(row), nil
}

func (r *OrganizationPostgresRepository) List(ctx context.Context) ([]*entity.Organization, error) {
	rows, err := r.queries.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.Organization, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapOrg(row))
	}
	return res, nil
}

func (r *OrganizationPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateOrganizationStatus(ctx, sqlc.UpdateOrganizationStatusParams{
		ID:     id,
		Status: status,
	})
}

func mapOrg(row sqlc.Organization) *entity.Organization {
	return &entity.Organization{
		ID:           row.ID,
		Slug:         row.Slug,
		Name:         row.Name,
		ContactEmail: row.ContactEmail.String,
		ContactPhone: row.ContactPhone.String,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
