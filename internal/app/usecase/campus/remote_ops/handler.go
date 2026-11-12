package remote_ops

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h *Handler) HandleGetLiveFleet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	fleet, err := h.service.GetLiveFleet(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": fleet})
}

func (h *Handler) HandleRemoteBeep(c *fiber.Ctx) error {
	slug := c.Params("slug")
	bikeID := c.Params("id")
	if slug == "" || bikeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and bike_id are required"})
	}

	if err := h.service.SendCommand(c.Context(), slug, bikeID, "beep"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ส่งคำสั่งส่งสัญญาณเสียง (Buzzer) สำเร็จ",
		"bike_id": bikeID,
	})
}

func (h *Handler) HandleRemoteUnlock(c *fiber.Ctx) error {
	slug := c.Params("slug")
	bikeID := c.Params("id")
	if slug == "" || bikeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and bike_id are required"})
	}

	if err := h.service.SendCommand(c.Context(), slug, bikeID, "unlock"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ส่งคำสั่งปลดล็อกสลักไฟฉุกเฉิน (Remote Unlock) สำเร็จ",
		"bike_id": bikeID,
	})
}

func (h *Handler) HandleRemoteLock(c *fiber.Ctx) error {
	slug := c.Params("slug")
	bikeID := c.Params("id")
	if slug == "" || bikeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and bike_id are required"})
	}

	if err := h.service.SendCommand(c.Context(), slug, bikeID, "lock"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ส่งคำสั่งล็อกระบบและตัด Relay (Remote Lock) สำเร็จ",
		"bike_id": bikeID,
	})
}

func (h *Handler) HandleForceEnd(c *fiber.Ctx) error {
	slug := c.Params("slug")
	bikeID := c.Params("id")
	if slug == "" || bikeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug and bike_id are required"})
	}

	if err := h.service.ForceEndRide(c.Context(), slug, bikeID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "บังคับจบทริปและคืนสถานะจักรยานพร้อมใช้งานสำเร็จ",
		"bike_id": bikeID,
	})
}
