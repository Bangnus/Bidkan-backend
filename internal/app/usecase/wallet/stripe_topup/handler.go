package stripe_topup

import (
	"errors"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(
	txRepo domainRepo.TransactionRepository,
	userRepo domainRepo.UserRepository,
	stripeClient infraStripe.Client,
) Handler {
	return Handler{svc: NewService(txRepo, userRepo, stripeClient)}
}

func (h *Handler) HandleConfig(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(h.svc.GetConfig())
}

func (h *Handler) HandleCheckout(c *fiber.Ctx) error {
	userID, ok := extractUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.UserID = userID

	res, err := h.svc.CreateCheckout(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *Handler) HandleVerify(c *fiber.Ctx) error {
	userID, ok := extractUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	res, err := h.svc.VerifySession(c.Context(), userID, c.Query("session_id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func extractUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	raw, ok := c.Locals("user_id").(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	return id, err == nil
}

func writeError(c *fiber.Ctx, err error) error {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		payload := fiber.Map{"error": svcErr.Msg}
		if svcErr.Code != "" {
			payload["code"] = svcErr.Code
		}
		return c.Status(svcErr.Status).JSON(payload)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}

