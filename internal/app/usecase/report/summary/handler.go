package summary

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	svc Service
}

func NewHandler(svc Service) Handler {
	return &handler{svc: svc}
}

// @Summary รายงานสรุปภาพรวมระบบ (Admin Only)
// @Description ดึงข้อมูลรายงานสรุปยอดรวม เช่น รายได้รวม จำนวนผู้ใช้ และสถิติการใช้งานรถจักรยาน
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.FullReport "ข้อมูลรายงานสรุป"
// @Failure 403 {object} map[string]string "ไม่มีสิทธิ์ (เฉพาะแอดมินเท่านั้น)"
// @Failure 500 {object} map[string]string "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/reports/summary [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	role := middleware.GetAuthRole(c)
	if role != "super_admin" && role != "org_admin" && role != "tenant_admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "permission denied: admin only"})
	}

	report, err := h.svc.GetSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}
