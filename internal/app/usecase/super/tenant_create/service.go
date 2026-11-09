package tenant_create

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	OrganizationID     *uuid.UUID `json:"organization_id"`
	OrganizationSlug   string     `json:"organization_slug"`
	OrganizationName   string     `json:"organization_name"`
	Slug               string     `json:"slug"`
	Name               string     `json:"name"`
	PlatformFeePercent float64    `json:"platform_fee_percent"`
	AutoTieredGP       *bool      `json:"auto_tiered_gp"`
	UnlockFee          float64    `json:"unlock_fee"`
	PerMinuteRate      float64    `json:"per_minute_rate"`
	ParkingPenaltyFee  float64    `json:"parking_penalty_fee"`
	SimFeePerBike      float64    `json:"sim_fee_per_bike"`
	AdminEmail         string     `json:"admin_email"`
	AdminUsername      string     `json:"admin_username"`
	AdminPassword      string     `json:"admin_password"`
	AdminRole          string     `json:"admin_role"`
}

type Service struct {
	repo       domainRepo.TenantRepository
	orgRepo    domainRepo.OrganizationRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewService(
	repo domainRepo.TenantRepository,
	orgRepo domainRepo.OrganizationRepository,
	campusRepo domainRepo.CampusOpsRepository,
) *Service {
	return &Service{repo: repo, orgRepo: orgRepo, campusRepo: campusRepo}
}

func (s *Service) Execute(ctx context.Context, req Request) (*entity.Tenant, error) {
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.Name = strings.TrimSpace(req.Name)
	if req.Slug == "" || req.Name == "" {
		return nil, errors.New("slug and name are required")
	}

	autoTiered := true
	if req.AutoTieredGP != nil {
		autoTiered = *req.AutoTieredGP
	}
	if autoTiered {
		req.PlatformFeePercent = entity.CalculateTieredGP(0).CurrentGPPercent
	} else if req.PlatformFeePercent <= 0 {
		req.PlatformFeePercent = 18.00
	}
	if req.PerMinuteRate <= 0 {
		req.PerMinuteRate = 1.50
	}
	if req.ParkingPenaltyFee <= 0 {
		req.ParkingPenaltyFee = 50.00
	}

	orgID := req.OrganizationID
	if orgID == nil && strings.TrimSpace(req.OrganizationSlug) != "" && s.orgRepo != nil {
		orgSlug := strings.ToLower(strings.TrimSpace(req.OrganizationSlug))
		if existing, err := s.orgRepo.GetBySlug(ctx, orgSlug); err == nil {
			orgID = &existing.ID
		} else {
			orgName := strings.TrimSpace(req.OrganizationName)
			if orgName == "" {
				orgName = req.Name
			}
			newOrgID := uuid.New()
			if err := s.orgRepo.Create(ctx, &entity.Organization{
				ID:     newOrgID,
				Slug:   orgSlug,
				Name:   orgName,
				Status: "active",
			}); err == nil {
				orgID = &newOrgID
			}
		}
	}

	tenant := &entity.Tenant{
		ID:                        uuid.New(),
		OrganizationID:            orgID,
		Slug:                      req.Slug,
		Name:                      req.Name,
		PlatformFeePercent:        req.PlatformFeePercent,
		AutoTieredGP:              autoTiered,
		UnlockFee:                 req.UnlockFee,
		PerMinuteRate:             req.PerMinuteRate,
		ParkingPenaltyFee:         req.ParkingPenaltyFee,
		SimFeePerBike:             req.SimFeePerBike,
		Status:                    "active",
		StripeOnboardingCompleted: false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	if req.AdminEmail != "" && req.AdminPassword != "" && s.campusRepo != nil {
		if hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost); err == nil {
			username := strings.TrimSpace(req.AdminUsername)
			if username == "" {
				username = strings.Split(strings.TrimSpace(req.AdminEmail), "@")[0]
			}
			role := strings.TrimSpace(req.AdminRole)
			if role != "org_admin" && role != "tenant_admin" {
				if orgID != nil {
					role = "org_admin"
				} else {
					role = "tenant_admin"
				}
			}
			now := time.Now()
			_ = s.campusRepo.CreateBackofficeUser(ctx, &entity.User{
				ID:             uuid.New(),
				TenantID:       &tenant.ID,
				OrganizationID: orgID,
				Email:          strings.TrimSpace(req.AdminEmail),
				Username:       username,
				Password:       string(hash),
				WalletBalance:  "0.00",
				Role:           role,
				Status:         "active",
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}

	return tenant, nil
}
