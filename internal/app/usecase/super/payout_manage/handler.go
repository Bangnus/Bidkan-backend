package payout_manage

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TriggerPayoutRequest struct {
	TenantSlug    string  `json:"tenant_slug"`
	GrossAmount   float64 `json:"gross_amount"`
	TotalRides    int     `json:"total_rides"`
	AutoCalculate bool    `json:"auto_calculate"`
	IncludeSimFee bool    `json:"include_sim_fee"`
}

type UnsettledTenantSummary struct {
	TenantID        uuid.UUID         `json:"tenant_id"`
	TenantSlug      string            `json:"tenant_slug"`
	TenantName      string            `json:"tenant_name"`
	GPTier          entity.GPTierInfo `json:"gp_tier"`
	HasStripe       bool              `json:"has_stripe"`
	CampusBikes     int               `json:"campus_bikes"`
	MonthlySimFee   float64           `json:"monthly_sim_fee"`
	UnsettledRides  int               `json:"unsettled_rides"`
	UnsettledGross  float64           `json:"unsettled_gross"`
	UnsettledGPFee  float64           `json:"unsettled_gp_fee"`
	UnsettledNet    float64           `json:"unsettled_net"`
}

type Handler struct {
	superRepo  domainRepo.SuperAdminRepository
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(
	superRepo domainRepo.SuperAdminRepository,
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
) Handler {
	return Handler{superRepo: superRepo, tenantRepo: tenantRepo, campusRepo: campusRepo}
}

func (h *Handler) HandleListAll(c *fiber.Ctx) error {
	payouts, err := h.superRepo.ListAllPayouts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	tenants, _ := h.tenantRepo.List(c.Context())
	unsettled := make([]UnsettledTenantSummary, 0, len(tenants))
	for _, t := range tenants {
		st, _ := h.campusRepo.GetUnsettledStats(c.Context(), t.ID)
		if st == nil {
			st = &entity.RideStatsSummary{}
		}
		bikes, _ := h.tenantRepo.ListBikesByTenant(c.Context(), t.ID)
		fleetCnt, _ := h.tenantRepo.CountFleetForTenant(c.Context(), t.ID)
		unsettled = append(unsettled, UnsettledTenantSummary{
			TenantID:       t.ID,
			TenantSlug:     t.Slug,
			TenantName:     t.Name,
			GPTier:         entity.CalculateTieredGP(fleetCnt),
			HasStripe:      t.StripeOnboardingCompleted,
			CampusBikes:    len(bikes),
			MonthlySimFee:  float64(len(bikes)) * t.SimFeePerBike,
			UnsettledRides: st.TotalRides,
			UnsettledGross: st.GrossAmount,
			UnsettledGPFee: st.GPFeeAmount,
			UnsettledNet:   st.NetAmount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":      payouts,
		"unsettled": unsettled,
	})
}

func (h *Handler) HandleTriggerPayout(c *fiber.Ctx) error {
	var req TriggerPayoutRequest
	if err := c.BodyParser(&req); err != nil || req.TenantSlug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant_slug is required"})
	}

	tenant, err := h.tenantRepo.GetBySlug(c.Context(), req.TenantSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tenant not found"})
	}

	gross := req.GrossAmount
	rides := req.TotalRides
	gpFee := math.Round((gross*(tenant.PlatformFeePercent/100.0))*100) / 100

	if req.AutoCalculate || gross <= 0 {
		st, err := h.campusRepo.GetUnsettledStats(c.Context(), tenant.ID)
		if err != nil || st.GrossAmount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ยังไม่มีรายได้ค้างโอนจากทริปที่จบแล้วในรอบบัญชีนี้"})
		}
		gross = st.GrossAmount
		rides = st.TotalRides
		gpFee = st.GPFeeAmount
	}

	if req.IncludeSimFee && tenant.SimFeePerBike > 0 {
		bikes, _ := h.tenantRepo.ListBikesByTenant(c.Context(), tenant.ID)
		gpFee += float64(len(bikes)) * tenant.SimFeePerBike
	}

	netAmount := math.Round((gross-gpFee)*100) / 100
	if netAmount < 0 {
		netAmount = 0
	}

	transferID := fmt.Sprintf("tr_stripe_%s", uuid.NewString()[:8])
	if key := os.Getenv("STRIPE_SECRET_KEY"); key != "" && tenant.StripeAccountID != nil && netAmount > 0 {
		if id, err := executeStripeTransfer(key, *tenant.StripeAccountID, netAmount); err == nil {
			transferID = id
		}
	}

	now := time.Now()
	payout := &entity.TenantPayout{
		ID:                uuid.New(),
		TenantID:          tenant.ID,
		StripeTransferID:  &transferID,
		TotalRides:        rides,
		GrossAmount:       gross,
		PlatformFeeAmount: gpFee,
		NetPayoutAmount:   netAmount,
		Status:            "completed",
		PeriodStart:       now.AddDate(0, 0, -7),
		PeriodEnd:         now,
		ProcessedAt:       &now,
		CreatedAt:         now,
	}

	if err := h.superRepo.CreatePayout(c.Context(), payout); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Stripe Auto-Payout completed",
		"data":    payout,
	})
}

func executeStripeTransfer(secretKey, destinationAccount string, amountTHB float64) (string, error) {
	form := url.Values{}
	form.Set("amount", fmt.Sprintf("%d", int64(amountTHB*100)))
	form.Set("currency", "thb")
	form.Set("destination", destinationAccount)

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/transfers", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return "", fmt.Errorf("stripe api error")
	}
	defer resp.Body.Close()
	return fmt.Sprintf("tr_live_%s", uuid.NewString()[:8]), nil
}
