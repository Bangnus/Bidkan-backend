package zone_manage

import (
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateZoneRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"` // 'parking' | 'riding_limit' | 'restricted'
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Boundary string  `json:"boundary"`
	Radius   float64 `json:"radius"`
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(tenantRepo domainRepo.TenantRepository, campusRepo domainRepo.CampusOpsRepository) Handler {
	return Handler{tenantRepo: tenantRepo, campusRepo: campusRepo}
}

func (h *Handler) HandleList(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	zones, err := h.campusRepo.ListZonesByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": zones})
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	var req CreateZoneRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and type are required"})
	}
	boundary := req.Boundary
	if boundary == "" {
		boundary = fmt.Sprintf(`{"lat":%.6f,"lon":%.6f}`, req.Lat, req.Lon)
	}
	if req.Radius <= 0 {
		req.Radius = 30
	}

	zone := &entity.TenantZone{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		Name:      req.Name,
		Type:      req.Type,
		Boundary:  boundary,
		Radius:    req.Radius,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.campusRepo.CreateZoneForTenant(c.Context(), zone); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "zone created",
		"data":    zone,
	})
}

func (h *Handler) HandleDelete(c *fiber.Ctx) error {
	zoneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid zone id"})
	}
	if err := h.campusRepo.DeleteZone(c.Context(), zoneID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "zone deleted"})
}
