package coupon_manage

import (
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateCouponRequest struct {
	Code      string  `json:"code"`
	Type      string  `json:"type"` // 'discount' | 'credit'
	Value     float64 `json:"value"`
	MinAmount float64 `json:"min_amount"`
	MaxUses   int32   `json:"max_uses"`
	DaysValid int     `json:"days_valid"`
}

type Handler struct {
	superRepo domainRepo.SuperAdminRepository
}

func NewHandler(superRepo domainRepo.SuperAdminRepository) Handler {
	return Handler{superRepo: superRepo}
}

func (h *Handler) HandleList(c *fiber.Ctx) error {
	if page, limit, ok := shared.ParsePageQuery(c); ok {
		coupons, total, err := h.superRepo.ListGlobalCouponsPaged(c.Context(), limit, (page-1)*limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": coupons,
			"meta": entity.NewPageMeta(page, limit, total),
		})
	}
	coupons, err := h.superRepo.ListGlobalCoupons(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": coupons})
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	var req CreateCouponRequest
	if err := c.BodyParser(&req); err != nil || req.Code == "" || req.Value <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code and positive value are required"})
	}
	if req.Type == "" {
		req.Type = "discount"
	}
	if req.MaxUses <= 0 {
		req.MaxUses = 100
	}
	if req.DaysValid <= 0 {
		req.DaysValid = 30
	}

	now := time.Now()
	coupon := &entity.Coupon{
		ID:        uuid.New(),
		Code:      req.Code,
		Type:      req.Type,
		Value:     fmt.Sprintf("%.2f", req.Value),
		MinAmount: fmt.Sprintf("%.2f", req.MinAmount),
		StartAt:   now,
		MaxUses:   req.MaxUses,
		ExpiredAt: now.AddDate(0, 0, req.DaysValid),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.superRepo.CreateGlobalCoupon(c.Context(), coupon); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "global coupon created",
		"data":    coupon,
	})
}
