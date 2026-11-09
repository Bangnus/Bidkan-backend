package partner_apply

import (
	"fmt"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ApplyRequest struct {
	OrganizationName  string  `json:"organization_name"`
	OrganizationSlug  string  `json:"organization_slug"`
	CampusName        string  `json:"campus_name"`
	CampusSlug        string  `json:"campus_slug"`
	ContactName       string  `json:"contact_name"`
	ContactEmail      string  `json:"contact_email"`
	ContactPhone      string  `json:"contact_phone"`
	UnlockFee         float64 `json:"unlock_fee"`
	PerMinuteRate     float64 `json:"per_minute_rate"`
	ParkingPenaltyFee float64 `json:"parking_penalty_fee"`
	InitialBikes      int32   `json:"initial_bikes"`
	AdminEmail        string  `json:"admin_email"`
	AdminUsername     string  `json:"admin_username"`
	AdminPassword     string  `json:"admin_password"`
	Notes             string  `json:"notes"`
}

type Handler struct {
	orgRepo    domainRepo.OrganizationRepository
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
	orderRepo  domainRepo.HardwareOrderRepository
}

func NewHandler(
	orgRepo domainRepo.OrganizationRepository,
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	orderRepo domainRepo.HardwareOrderRepository,
) Handler {
	return Handler{orgRepo: orgRepo, tenantRepo: tenantRepo, campusRepo: campusRepo, orderRepo: orderRepo}
}

func (h *Handler) HandleApply(c *fiber.Ctx) error {
	var req ApplyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.CampusSlug = strings.ToLower(strings.TrimSpace(req.CampusSlug))
	req.CampusName = strings.TrimSpace(req.CampusName)
	req.ContactEmail = strings.TrimSpace(req.ContactEmail)
	if req.CampusSlug == "" || req.CampusName == "" || req.ContactEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "campus_name, campus_slug, and contact_email are required"})
	}
	if _, err := h.tenantRepo.GetBySlug(c.Context(), req.CampusSlug); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "รหัสวิทยาเขต (slug) นี้ถูกใช้งานแล้ว กรุณาใช้รหัสอื่น"})
	}

	orgSlug := strings.ToLower(strings.TrimSpace(req.OrganizationSlug))
	orgName := strings.TrimSpace(req.OrganizationName)
	if orgSlug == "" {
		orgSlug = req.CampusSlug + "-org"
	}
	if orgName == "" {
		orgName = req.CampusName
	}

	var orgID uuid.UUID
	if existingOrg, err := h.orgRepo.GetBySlug(c.Context(), orgSlug); err == nil {
		orgID = existingOrg.ID
	} else {
		orgID = uuid.New()
		org := &entity.Organization{
			ID:           orgID,
			Slug:         orgSlug,
			Name:         orgName,
			ContactEmail: req.ContactEmail,
			ContactPhone: strings.TrimSpace(req.ContactPhone),
			Status:       "pending",
		}
		if err := h.orgRepo.Create(c.Context(), org); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	if req.PerMinuteRate <= 0 {
		req.PerMinuteRate = 1.50
	}
	if req.ParkingPenaltyFee <= 0 {
		req.ParkingPenaltyFee = 50.00
	}
	tier := entity.CalculateTieredGP(int(req.InitialBikes))

	tenant := &entity.Tenant{
		ID:                 uuid.New(),
		OrganizationID:     &orgID,
		Slug:               req.CampusSlug,
		Name:               req.CampusName,
		PlatformFeePercent: tier.CurrentGPPercent,
		AutoTieredGP:       true,
		UnlockFee:          req.UnlockFee,
		PerMinuteRate:      req.PerMinuteRate,
		ParkingPenaltyFee:  req.ParkingPenaltyFee,
		ContactName:        strings.TrimSpace(req.ContactName),
		ContactEmail:       req.ContactEmail,
		ContactPhone:       strings.TrimSpace(req.ContactPhone),
		Notes:              strings.TrimSpace(req.Notes),
		Status:             "pending",
	}
	if err := h.tenantRepo.Create(c.Context(), tenant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if req.AdminEmail != "" && req.AdminPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
		if err == nil {
			username := strings.TrimSpace(req.AdminUsername)
			if username == "" {
				username = strings.Split(req.AdminEmail, "@")[0]
			}
			now := time.Now()
			_ = h.campusRepo.CreateBackofficeUser(c.Context(), &entity.User{
				ID:             uuid.New(),
				TenantID:       &tenant.ID,
				OrganizationID: &orgID,
				Email:          strings.TrimSpace(req.AdminEmail),
				PhoneNumber:    strings.TrimSpace(req.ContactPhone),
				Username:       username,
				Password:       string(hash),
				WalletBalance:  "0.00",
				Role:           "org_admin",
				Status:         "pending",
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}

	if req.InitialBikes > 0 {
		note := fmt.Sprintf("คำสั่งซื้อรถล็อตเริ่มต้นจากการสมัครพาร์ทเนอร์ (%s)", req.CampusName)
		_ = h.orderRepo.Create(c.Context(), &entity.HardwareOrder{
			ID:          uuid.New(),
			TenantID:    tenant.ID,
			OrderType:   "complete_bike",
			Quantity:    req.InitialBikes,
			UnitPrice:   14500,
			TotalAmount: 14500 * float64(req.InitialBikes),
			Note:        note,
			Status:      "pending",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "ส่งคำขอสมัครพาร์ทเนอร์เรียบร้อยแล้ว รอทีมงาน BidKan ตรวจสอบและอนุมัติ",
		"data": fiber.Map{
			"tenant":  tenant,
			"gp_tier": tier,
		},
	})
}
