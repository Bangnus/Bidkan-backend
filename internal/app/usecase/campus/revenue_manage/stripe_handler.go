package revenue_manage

import (
	"fmt"

	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
)

type StripeStatusResponse struct {
	StripeConfigured    bool     `json:"stripe_configured"`
	AccountType         string   `json:"account_type"`
	StripeAccountID     *string  `json:"stripe_account_id"`
	OnboardingCompleted bool     `json:"onboarding_completed"`
	ChargesEnabled      bool     `json:"charges_enabled"`
	PayoutsEnabled      bool     `json:"payouts_enabled"`
	DetailsSubmitted    bool     `json:"details_submitted"`
	CurrentlyDue        []string `json:"currently_due"`
	DisabledReason      string   `json:"disabled_reason"`
}

func (h *Handler) HandleGetStripeStatus(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}

	res := StripeStatusResponse{
		StripeConfigured: h.stripe.IsConfigured(),
		AccountType:      h.stripe.Config().ConnectAccountType,
		StripeAccountID:  tenant.StripeAccountID,
		CurrentlyDue:     []string{},
	}
	if !h.stripe.IsConfigured() || tenant.StripeAccountID == nil || *tenant.StripeAccountID == "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res})
	}

	acct, err := h.stripe.GetAccount(c.Context(), *tenant.StripeAccountID)
	if err != nil {
		if infraStripe.IsMissingResource(err) {
			_ = h.tenantRepo.UpdateStripeStatus(c.Context(), tenant.ID, "", false)
			res.StripeAccountID = nil
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	completed := acct.PayoutsEnabled
	if completed != tenant.StripeOnboardingCompleted {
		_ = h.tenantRepo.UpdateStripeStatus(c.Context(), tenant.ID, acct.ID, completed)
	}
	res.OnboardingCompleted = completed
	res.ChargesEnabled = acct.ChargesEnabled
	res.PayoutsEnabled = acct.PayoutsEnabled
	res.DetailsSubmitted = acct.DetailsSubmitted
	res.CurrentlyDue = acct.CurrentlyDue
	res.DisabledReason = acct.DisabledReason
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res})
}

func (h *Handler) HandleStripeOnboard(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	if !h.stripe.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY) กรุณาติดต่อผู้ดูแลส่วนกลาง",
			"code":  "STRIPE_NOT_CONFIGURED",
		})
	}

	accountID := ""
	if tenant.StripeAccountID != nil && *tenant.StripeAccountID != "" {
		if _, err := h.stripe.GetAccount(c.Context(), *tenant.StripeAccountID); err == nil {
			accountID = *tenant.StripeAccountID
		} else if !infraStripe.IsMissingResource(err) {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
	}

	if accountID == "" {
		createdID, err := h.stripe.CreateConnectedAccount(c.Context(), tenant.Slug, tenant.Name, tenant.ContactEmail)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		accountID = createdID
		if err := h.tenantRepo.UpdateStripeStatus(c.Context(), tenant.ID, accountID, false); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	baseURL := h.stripe.Config().BackofficeURL
	refreshURL := fmt.Sprintf("%s/campus/%s/revenue?stripe_onboarding=refresh", baseURL, tenant.Slug)
	returnURL := fmt.Sprintf("%s/campus/%s/revenue?stripe_onboarding=return", baseURL, tenant.Slug)
	linkURL, err := h.stripe.CreateAccountLink(c.Context(), accountID, refreshURL, returnURL)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":           "Stripe Connect onboarding link created",
		"stripe_account_id": accountID,
		"onboarding_url":    linkURL,
	})
}

func (h *Handler) HandleStripeLoginLink(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	if !h.stripe.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY)",
			"code":  "STRIPE_NOT_CONFIGURED",
		})
	}
	if tenant.StripeAccountID == nil || *tenant.StripeAccountID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ยังไม่ได้เชื่อมต่อบัญชี Stripe Connect"})
	}
	if h.stripe.Config().ConnectAccountType == "standard" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"login_url": "https://dashboard.stripe.com/"})
	}
	loginURL, err := h.stripe.CreateExpressLoginLink(c.Context(), *tenant.StripeAccountID)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"login_url": loginURL})
}

