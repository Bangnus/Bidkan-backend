package tenant_list

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/shared"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	if page, limit, ok := shared.ParsePageQuery(c); ok {
		tenants, total, err := h.service.ExecutePaged(c.Context(), limit, (page-1)*limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": tenants,
			"meta": entity.NewPageMeta(page, limit, total),
		})
	}

	tenants, err := h.service.Execute(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": tenants,
	})
}

