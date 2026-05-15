package logout

import (
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct{}

func NewHandler() Handler {
	return &handler{}
}

// @Summary Logout User
// @Description Logout the current user by instructing the client to clear the JWT token.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /v1/users/logout [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	// ในระบบ JWT แบบ Stateless การ Logout คือการให้ฝั่ง Client (แอปมือถือ) ลบ Token ทิ้ง
	// ฝั่ง Backend แค่ส่งสถานะ 200 กลับไปยืนยัน
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
