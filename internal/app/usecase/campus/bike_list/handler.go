package bike_list

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
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug is required",
		})
	}

	if page, limit, ok := shared.ParsePageQuery(c); ok {
		bikes, total, err := h.service.ExecutePaged(c.Context(), slug, c.Query("status"), limit, (page-1)*limit)
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if err.Error() == "campus tenant not found" {
				statusCode = fiber.StatusNotFound
			}
			return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": bikes,
			"meta": entity.NewPageMeta(page, limit, total),
		})
	}

	bikes, err := h.service.Execute(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": bikes,
	})
}

