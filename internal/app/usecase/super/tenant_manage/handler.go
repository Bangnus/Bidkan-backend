package tenant_manage

import (
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StatusRequest struct {
	Status string `json:"status"`
}

type RatesRequest struct {
	UnlockFee          float64  `json:"unlock_fee"`
	PerMinuteRate      float64  `json:"per_minute_rate"`
	PlatformFeePercent float64  `json:"platform_fee_percent"`
	ParkingPenaltyFee  *float64 `json:"parking_penalty_fee"`
	SimFeePerBike      *float64 `json:"sim_fee_per_bike"`
	AutoTieredGP       *bool    `json:"auto_tiered_gp"`
}

type CreateOrgRequest struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	superRepo  domainRepo.SuperAdminRepository
	orgRepo    domainRepo.OrganizationRepository
	campusRepo domainRepo.CampusOpsRepository
	orderRepo  domainRepo.HardwareOrderRepository
}

func NewHandler(
	tenantRepo domainRepo.TenantRepository,
	superRepo domainRepo.SuperAdminRepository,
	orgRepo domainRepo.OrganizationRepository,
	campusRepo domainRepo.CampusOpsRepository,
	orderRepo domainRepo.HardwareOrderRepository,
) Handler {
	return Handler{
		tenantRepo: tenantRepo,
		superRepo:  superRepo,
		orgRepo:    orgRepo,
		campusRepo: campusRepo,
		orderRepo:  orderRepo,
	}
}

func (h *Handler) HandleUpdateStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	var req StatusRequest
	if err := c.BodyParser(&req); err != nil || (req.Status != "active" && req.Status != "suspended" && req.Status != "pending") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be pending, active, or suspended"})
	}
	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tenant not found"})
	}
	if err := h.superRepo.UpdateTenantStatus(c.Context(), id, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Status == "active" {
		if tenant.OrganizationID != nil {
			_ = h.orgRepo.UpdateStatus(c.Context(), *tenant.OrganizationID, "active")
		}
		_ = h.superRepo.ActivatePendingAdmins(c.Context(), tenant.ID, tenant.OrganizationID)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "tenant status updated"})
}

func (h *Handler) HandleUpdateRates(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tenant not found"})
	}
	var req RatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	autoTiered := tenant.AutoTieredGP
	if req.AutoTieredGP != nil {
		autoTiered = *req.AutoTieredGP
	}
	penalty := tenant.ParkingPenaltyFee
	if req.ParkingPenaltyFee != nil {
		penalty = *req.ParkingPenaltyFee
	}
	simFee := tenant.SimFeePerBike
	if req.SimFeePerBike != nil {
		simFee = *req.SimFeePerBike
	}

	gp := req.PlatformFeePercent
	if autoTiered {
		cnt, _ := h.tenantRepo.CountFleetForTenant(c.Context(), id)
		gp = entity.CalculateTieredGP(cnt).CurrentGPPercent
	}

	if err := h.tenantRepo.UpdateRatesFull(c.Context(), id, req.UnlockFee, req.PerMinuteRate, gp, penalty, simFee, autoTiered); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":              "tenant rates updated",
		"platform_fee_percent": gp,
		"auto_tiered_gp":       autoTiered,
	})
}

func (h *Handler) HandleListOrganizations(c *fiber.Ctx) error {
	orgs, err := h.orgRepo.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": orgs})
}

func (h *Handler) HandleCreateOrganization(c *fiber.Ctx) error {
	var req CreateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.Name = strings.TrimSpace(req.Name)
	if req.Slug == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and name are required"})
	}
	org := &entity.Organization{
		ID:           uuid.New(),
		Slug:         req.Slug,
		Name:         req.Name,
		ContactEmail: strings.TrimSpace(req.ContactEmail),
		ContactPhone: strings.TrimSpace(req.ContactPhone),
		Status:       "active",
	}
	if err := h.orgRepo.Create(c.Context(), org); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "organization created", "data": org})
}
