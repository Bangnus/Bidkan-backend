package manage

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CampusSummary struct {
	Tenant         *entity.Tenant `json:"tenant"`
	TotalBikes     int            `json:"total_bikes"`
	AvailableBikes int            `json:"available_bikes"`
	TodayTrips     int            `json:"today_trips"`
	TodayGross     float64        `json:"today_gross"`
	TodayNet       float64        `json:"today_net"`
	CompletedGross float64        `json:"completed_gross"`
	UnsettledNet   float64        `json:"unsettled_net"`
}

type Handler struct {
	orgRepo    domainRepo.OrganizationRepository
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(
	orgRepo domainRepo.OrganizationRepository,
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
) Handler {
	return Handler{orgRepo: orgRepo, tenantRepo: tenantRepo, campusRepo: campusRepo}
}

func (h *Handler) HandleOverview(c *fiber.Ctx) error {
	org, err := h.orgRepo.GetBySlug(c.Context(), c.Params("orgSlug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "organization not found"})
	}
	if !h.canAccessOrg(c, org.ID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: You do not have access to this organization"})
	}

	campuses, err := h.tenantRepo.ListByOrganization(c.Context(), org.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	summaries := make([]CampusSummary, 0, len(campuses))
	var totalBikes, availableBikes, todayTrips int
	var todayGross, todayNet, completedGross, unsettledNet float64

	for _, t := range campuses {
		bikes, _ := h.tenantRepo.ListBikesByTenant(c.Context(), t.ID)
		avail := 0
		for _, b := range bikes {
			if b.Status == "available" {
				avail++
			}
		}
		today, _ := h.campusRepo.GetTodayStats(c.Context(), t.ID)
		completed, _ := h.campusRepo.GetCompletedStats(c.Context(), t.ID)
		unsettled, _ := h.campusRepo.GetUnsettledStats(c.Context(), t.ID)
		if today == nil {
			today = &entity.RideStatsSummary{}
		}
		if completed == nil {
			completed = &entity.RideStatsSummary{}
		}
		if unsettled == nil {
			unsettled = &entity.RideStatsSummary{}
		}

		totalBikes += len(bikes)
		availableBikes += avail
		todayTrips += today.TotalRides
		todayGross += today.GrossAmount
		todayNet += today.NetAmount
		completedGross += completed.GrossAmount
		unsettledNet += unsettled.NetAmount

		summaries = append(summaries, CampusSummary{
			Tenant:         t,
			TotalBikes:     len(bikes),
			AvailableBikes: avail,
			TodayTrips:     today.TotalRides,
			TodayGross:     today.GrossAmount,
			TodayNet:       today.NetAmount,
			CompletedGross: completed.GrossAmount,
			UnsettledNet:   unsettled.NetAmount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"organization": org,
			"gp_tier":      entity.CalculateTieredGP(totalBikes),
			"totals": fiber.Map{
				"total_campuses":  len(campuses),
				"total_bikes":     totalBikes,
				"available_bikes": availableBikes,
				"today_trips":     todayTrips,
				"today_gross":     todayGross,
				"today_net":       todayNet,
				"completed_gross": completedGross,
				"unsettled_net":   unsettledNet,
			},
			"campuses": summaries,
		},
	})
}

func (h *Handler) canAccessOrg(c *fiber.Ctx, orgID uuid.UUID) bool {
	role, _ := c.Locals("role").(string)
	userOrgID, _ := c.Locals("organization_id").(string)
	return role == "org_admin" && userOrgID != "" && userOrgID == orgID.String()
}
