package revenue_manage

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
)

type UpdateRatesRequest struct {
	UnlockFee         float64  `json:"unlock_fee"`
	PerMinuteRate     float64  `json:"per_minute_rate"`
	ParkingPenaltyFee *float64 `json:"parking_penalty_fee"`
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
	stripe     infraStripe.Client
}

func NewHandler(
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	stripeClient infraStripe.Client,
) Handler {
	if stripeClient == nil {
		stripeClient = infraStripe.NewClient(nil)
	}
	return Handler{
		tenantRepo: tenantRepo,
		campusRepo: campusRepo,
		stripe:     stripeClient,
	}
}

func (h *Handler) HandleGetRevenue(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	payouts, err := h.campusRepo.ListPayoutsByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	completedStats, _ := h.campusRepo.GetCompletedStats(c.Context(), tenant.ID)
	unsettledStats, _ := h.campusRepo.GetUnsettledStats(c.Context(), tenant.ID)
	if completedStats == nil {
		completedStats = &entity.RideStatsSummary{}
	}
	if unsettledStats == nil {
		unsettledStats = &entity.RideStatsSummary{}
	}

	fleetCount, _ := h.tenantRepo.CountFleetForTenant(c.Context(), tenant.ID)
	campusBikes, _ := h.tenantRepo.ListBikesByTenant(c.Context(), tenant.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"tenant":                      tenant,
			"gp_tier":                     entity.CalculateTieredGP(fleetCount),
			"campus_bike_count":           len(campusBikes),
			"monthly_sim_fee":             float64(len(campusBikes)) * tenant.SimFeePerBike,
			"stripe_configured":           h.stripe.IsConfigured(),
			"stripe_connect_account_type": h.stripe.Config().ConnectAccountType,
			"total_gross":                 completedStats.GrossAmount,
			"total_gp_fee":                completedStats.GPFeeAmount,
			"total_net":                   completedStats.NetAmount,
			"unsettled_rides":             unsettledStats.TotalRides,
			"unsettled_gross":             unsettledStats.GrossAmount,
			"unsettled_gp_fee":            unsettledStats.GPFeeAmount,
			"unsettled_net":               unsettledStats.NetAmount,
			"payouts":                     payouts,
		},
	})
}

func (h *Handler) HandleUpdateRates(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	var req UpdateRatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	penalty := tenant.ParkingPenaltyFee
	if req.ParkingPenaltyFee != nil {
		penalty = *req.ParkingPenaltyFee
	}
	if err := h.tenantRepo.UpdateRatesFull(
		c.Context(),
		tenant.ID,
		req.UnlockFee,
		req.PerMinuteRate,
		tenant.PlatformFeePercent,
		penalty,
		tenant.SimFeePerBike,
		tenant.AutoTieredGP,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "campus pricing rates updated"})
}
