package manage

import (
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AddCampusRequest struct {
	Slug              string  `json:"slug"`
	Name              string  `json:"name"`
	UnlockFee         float64 `json:"unlock_fee"`
	PerMinuteRate     float64 `json:"per_minute_rate"`
	ParkingPenaltyFee float64 `json:"parking_penalty_fee"`
	AdminEmail        string  `json:"admin_email"`
	AdminUsername     string  `json:"admin_username"`
	AdminPassword     string  `json:"admin_password"`
}

func (h *Handler) HandleAddCampus(c *fiber.Ctx) error {
	org, err := h.orgRepo.GetBySlug(c.Context(), c.Params("orgSlug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "organization not found"})
	}
	if !h.canAccessOrg(c, org.ID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
	var req AddCampusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.Name = strings.TrimSpace(req.Name)
	if req.Slug == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and name are required"})
	}
	if req.PerMinuteRate <= 0 {
		req.PerMinuteRate = 1.50
	}
	if req.ParkingPenaltyFee <= 0 {
		req.ParkingPenaltyFee = 50.00
	}

	campuses, _ := h.tenantRepo.ListByOrganization(c.Context(), org.ID)
	totalFleet := 0
	for _, ct := range campuses {
		if bikes, err := h.tenantRepo.ListBikesByTenant(c.Context(), ct.ID); err == nil {
			totalFleet += len(bikes)
		}
	}
	tier := entity.CalculateTieredGP(totalFleet)

	tenant := &entity.Tenant{
		ID:                 uuid.New(),
		OrganizationID:     &org.ID,
		Slug:               req.Slug,
		Name:               req.Name,
		PlatformFeePercent: tier.CurrentGPPercent,
		AutoTieredGP:       true,
		UnlockFee:          req.UnlockFee,
		PerMinuteRate:      req.PerMinuteRate,
		ParkingPenaltyFee:  req.ParkingPenaltyFee,
		Status:             "active",
	}
	if err := h.tenantRepo.Create(c.Context(), tenant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if req.AdminEmail != "" && req.AdminPassword != "" {
		if hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost); err == nil {
			username := strings.TrimSpace(req.AdminUsername)
			if username == "" {
				username = strings.Split(req.AdminEmail, "@")[0]
			}
			now := time.Now()
			_ = h.campusRepo.CreateBackofficeUser(c.Context(), &entity.User{
				ID:             uuid.New(),
				TenantID:       &tenant.ID,
				OrganizationID: &org.ID,
				Email:          strings.TrimSpace(req.AdminEmail),
				Username:       username,
				Password:       string(hash),
				WalletBalance:  "0.00",
				Role:           "tenant_admin",
				Status:         "active",
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "เพิ่มวิทยาเขตใหม่ภายใต้สังกัดสำเร็จ",
		"data":    tenant,
	})
}
