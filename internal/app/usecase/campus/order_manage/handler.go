package order_manage

import (
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateOrderRequest struct {
	OrderType string  `json:"order_type"`
	Quantity  int32   `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Note      string  `json:"note"`
}

var defaultCatalogPrices = map[string]float64{
	"complete_bike":  14500,
	"spare_battery":  4200,
	"iot_controller": 1850,
	"spare_parts":    950,
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	orderRepo  domainRepo.HardwareOrderRepository
}

func NewHandler(tenantRepo domainRepo.TenantRepository, orderRepo domainRepo.HardwareOrderRepository) Handler {
	return Handler{tenantRepo: tenantRepo, orderRepo: orderRepo}
}

func (h *Handler) HandleList(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	fleetCount, _ := h.tenantRepo.CountFleetForTenant(c.Context(), tenant.ID)
	gpTier := entity.CalculateTieredGP(fleetCount)
	if page, limit, ok := shared.ParsePageQuery(c); ok {
		orders, total, err := h.orderRepo.ListByTenantPaged(c.Context(), tenant.ID, limit, (page-1)*limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":    orders,
			"meta":    entity.NewPageMeta(page, limit, total),
			"gp_tier": gpTier,
		})
	}
	orders, err := h.orderRepo.ListByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    orders,
		"gp_tier": gpTier,
	})
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.OrderType = strings.TrimSpace(req.OrderType)
	if req.OrderType == "" || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order_type and positive quantity are required"})
	}
	unitPrice := req.UnitPrice
	if unitPrice <= 0 {
		unitPrice = defaultCatalogPrices[req.OrderType]
	}

	var createdBy *uuid.UUID
	if uidStr, ok := c.Locals("user_id").(string); ok {
		if id, err := uuid.Parse(uidStr); err == nil {
			createdBy = &id
		}
	}

	now := time.Now()
	order := &entity.HardwareOrder{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		TenantSlug:  tenant.Slug,
		TenantName:  tenant.Name,
		OrderType:   req.OrderType,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		TotalAmount: unitPrice * float64(req.Quantity),
		Note:        strings.TrimSpace(req.Note),
		Status:      "pending",
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.orderRepo.Create(c.Context(), order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "hardware order request submitted",
		"data":    order,
	})
}
