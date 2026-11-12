package rider_manage

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type RefundPenaltyRequest struct {
	RideID       uuid.UUID `json:"ride_id"`
	UserID       uuid.UUID `json:"user_id"`
	RefundAmount float64   `json:"refund_amount"`
	Reason       string    `json:"reason"`
}

type Service struct {
	tenantRepo domainRepo.TenantRepository
	riderRepo  domainRepo.CampusRiderRepository
}

func NewService(tenantRepo domainRepo.TenantRepository, riderRepo domainRepo.CampusRiderRepository) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		riderRepo:  riderRepo,
	}
}

func (s *Service) ListRiders(ctx context.Context, slug, search, filter string) ([]*entity.CampusRider, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.riderRepo.ListCampusRiders(ctx, tenant.ID, search, filter)
}

func (s *Service) VerifyStudent(ctx context.Context, slug string, userID uuid.UUID, verified bool) error {
	_, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return errors.New("campus tenant not found")
	}
	return s.riderRepo.UpdateStudentVerification(ctx, userID, verified)
}

func (s *Service) UpdateUserStatus(ctx context.Context, slug string, userID uuid.UUID, status string) error {
	_, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return errors.New("campus tenant not found")
	}
	if status != "active" && status != "suspended" {
		return errors.New("invalid status: must be active or suspended")
	}
	return s.riderRepo.UpdateUserStatus(ctx, userID, status)
}

func (s *Service) ListDisputes(ctx context.Context, slug, status string) ([]*entity.DisputedRide, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.riderRepo.ListDisputedRides(ctx, tenant.ID, status)
}

func (s *Service) ProcessRefund(ctx context.Context, slug string, req RefundPenaltyRequest) error {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return errors.New("campus tenant not found")
	}
	if req.RefundAmount <= 0 {
		return errors.New("refund amount must be greater than zero")
	}
	if req.Reason == "" {
		req.Reason = "อนุมัติคำร้องขอคืนเงินค่าปรับตามดุลยพินิจผู้ดูแลระบบ"
	}
	return s.riderRepo.RefundPenalty(ctx, tenant.ID, req.RideID, req.UserID, req.RefundAmount, req.Reason)
}
