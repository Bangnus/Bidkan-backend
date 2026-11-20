package incident_manage

import (
	"fmt"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func resolveActorName(c *fiber.Ctx) string {
	if name, ok := c.Locals("user_name").(string); ok && name != "" {
		return name
	}
	switch middleware.GetAuthRole(c) {
	case "tenant_admin":
		return "Campus Admin"
	case "org_admin":
		return "Organization Admin"
	default:
		return "Campus Staff"
	}
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListIncidents(c *fiber.Ctx) error {
	slug := c.Params("slug")
	status := c.Query("status")
	severity := c.Query("severity")

	summary, err := h.service.ListIncidents(c.Context(), slug, status, severity)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": summary})
}

func (h *Handler) CreateIncident(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var req CreateIncidentReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	actor := resolveActorName(c)

	inc, err := h.service.CreateIncident(c.Context(), slug, req, actor)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": inc})
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	slug := c.Params("slug")
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid incident id"})
	}

	var req UpdateStatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	actor := resolveActorName(c)

	inc, err := h.service.UpdateStatus(c.Context(), slug, id, req, actor)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": inc})
}

func (h *Handler) ExportCSV(c *fiber.Ctx) error {
	slug := c.Params("slug")
	summary, err := h.service.ListIncidents(c.Context(), slug, "", "")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=incidents-%s.csv", slug))

	var b strings.Builder
	b.WriteString("ID,Date,Reporter,Phone,Type,Severity,Status,ClaimAmount,ApprovedAmount,PoliceReport,Hospital\n")
	for _, inc := range summary.Incidents {
		b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%.2f,%.2f,%s,%s\n",
			inc.ID, inc.CreatedAt.Format("2006-01-02 15:04"),
			inc.ReporterName, inc.ReporterPhone, inc.IncidentType, inc.Severity,
			inc.InsuranceStatus, inc.ClaimAmount, inc.ApprovedAmount,
			inc.PoliceReportNo, inc.HospitalName,
		))
	}
	return c.SendString(b.String())
}
