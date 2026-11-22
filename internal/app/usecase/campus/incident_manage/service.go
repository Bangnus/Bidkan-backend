package incident_manage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type CreateIncidentReq struct {
	ReporterName   string  `json:"reporter_name" validate:"required"`
	ReporterPhone  string  `json:"reporter_phone" validate:"required"`
	IncidentType   string  `json:"incident_type" validate:"required"`
	Severity       string  `json:"severity" validate:"required"`
	Description    string  `json:"description" validate:"required"`
	LocationName   string  `json:"location_name"`
	LocationLat    float64 `json:"location_lat"`
	LocationLon    float64 `json:"location_lon"`
	PoliceReportNo string  `json:"police_report_no"`
	ClaimAmount    float64 `json:"claim_amount"`
	HospitalName   string  `json:"hospital_name"`
	Notes          string  `json:"notes"`
	RideID         *string `json:"ride_id,omitempty"`
	UserID         *string `json:"user_id,omitempty"`
}

type UpdateStatusReq struct {
	Status         string  `json:"status" validate:"required"`
	ApprovedAmount float64 `json:"approved_amount"`
	ClaimNo        string  `json:"claim_no"`
	Notes          string  `json:"notes"`
}

type Service struct {
	tenantRepo repository.TenantRepository
	incRepo    repository.IncidentRepository
	auditRepo  repository.AuditRepository
}

func NewService(
	tenantRepo repository.TenantRepository,
	incRepo repository.IncidentRepository,
	auditRepo repository.AuditRepository,
) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		incRepo:    incRepo,
		auditRepo:  auditRepo,
	}
}

func (s *Service) ListIncidents(ctx context.Context, slug, status, severity string) (*entity.IncidentSummary, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	incidents, err := s.incRepo.ListByTenant(ctx, tenant.ID, status, severity)
	if err != nil {
		return nil, err
	}

	summary := &entity.IncidentSummary{
		TotalIncidents: len(incidents),
		Incidents:      incidents,
	}

	for _, inc := range incidents {
		summary.TotalClaimedAmount += inc.ClaimAmount
		summary.TotalSettledAmount += inc.ApprovedAmount
		if inc.InsuranceStatus == "filed" || inc.InsuranceStatus == "under_review" {
			summary.PendingClaims++
		}
	}
	return summary, nil
}

func (s *Service) CreateIncident(ctx context.Context, slug string, req CreateIncidentReq, actor string) (*entity.IncidentReport, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	inc := &entity.IncidentReport{
		ID:              uuid.New(),
		TenantID:        tenant.ID,
		ReporterName:    req.ReporterName,
		ReporterPhone:   req.ReporterPhone,
		IncidentType:    req.IncidentType,
		Severity:        req.Severity,
		Description:     req.Description,
		LocationName:    req.LocationName,
		LocationLat:     req.LocationLat,
		LocationLon:     req.LocationLon,
		PoliceReportNo:  req.PoliceReportNo,
		InsuranceStatus: "unfiled",
		ClaimAmount:     req.ClaimAmount,
		HospitalName:    req.HospitalName,
		Notes:           req.Notes,
	}

	if req.ClaimAmount > 0 {
		inc.InsuranceStatus = "filed"
	}

	if req.RideID != nil {
		if rID, err := uuid.Parse(*req.RideID); err == nil {
			inc.RideID = &rID
		}
	}
	if req.UserID != nil {
		if uID, err := uuid.Parse(*req.UserID); err == nil {
			inc.UserID = &uID
		}
	}

	if err := s.incRepo.Create(ctx, inc); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		_ = s.auditRepo.LogAction(ctx, &entity.AuditLog{
			TenantID:   tenant.ID,
			ActorName:  actor,
			ActorRole:  "tenant_admin",
			Action:     "incident_reported",
			TargetType: "incident",
			TargetID:   inc.ID.String(),
			Details:    fmt.Sprintf("New incident reported: %s (%s). Severity: %s", inc.IncidentType, inc.ReporterName, inc.Severity),
			IPAddress:  "127.0.0.1",
		})
	}

	return s.incRepo.GetByID(ctx, inc.ID)
}

func (s *Service) UpdateStatus(ctx context.Context, slug string, id uuid.UUID, req UpdateStatusReq, actor string) (*entity.IncidentReport, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	inc, err := s.incRepo.UpdateStatus(ctx, id, req.Status, req.ApprovedAmount, req.ClaimNo, req.Notes)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		_ = s.auditRepo.LogAction(ctx, &entity.AuditLog{
			TenantID:   tenant.ID,
			ActorName:  actor,
			ActorRole:  "tenant_admin",
			Action:     "incident_claim_updated",
			TargetType: "incident",
			TargetID:   id.String(),
			Details:    fmt.Sprintf("Claim updated to %s. Approved: ฿%.2f. Ref: %s", req.Status, req.ApprovedAmount, req.ClaimNo),
			IPAddress:  "127.0.0.1",
		})
	}

	return inc, nil
}
