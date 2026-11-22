package audit_manage

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h *Handler) HandleListAuditLogs(c *fiber.Ctx) error {
	slug := c.Params("slug")
	action := c.Query("action")

	logs, err := h.service.ListAuditLogs(c.Context(), slug, action)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": logs})
}

func (h *Handler) HandleExportRidesCSV(c *fiber.Ctx) error {
	slug := c.Params("slug")
	data, err := h.service.ExportRidesCSV(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	filename := fmt.Sprintf("bidkan-rides-%s-%s.csv", slug, time.Now().Format("20060102"))
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(data)
}

func (h *Handler) HandleExportRevenueCSV(c *fiber.Ctx) error {
	slug := c.Params("slug")
	data, err := h.service.ExportRevenueCSV(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	filename := fmt.Sprintf("bidkan-revenue-%s-%s.csv", slug, time.Now().Format("20060102"))
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(data)
}

func (h *Handler) HandleExportStaffPayoutsCSV(c *fiber.Ctx) error {
	slug := c.Params("slug")
	data, err := h.service.ExportStaffPayoutsCSV(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	filename := fmt.Sprintf("bidkan-staff-payouts-%s-%s.csv", slug, time.Now().Format("20060102"))
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(data)
}

