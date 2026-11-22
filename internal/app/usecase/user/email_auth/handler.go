package email_auth

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type Handler interface {
	HandleRegister(c *fiber.Ctx) error
	HandleVerifyEmail(c *fiber.Ctx) error
	HandleResendVerification(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) HandleRegister(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณากรอกอีเมล ชื่อผู้ใช้ และรหัสผ่าน (อย่างน้อย 6 ตัวอักษร) ให้ถูกต้อง"})
	}
	res, err := h.service.Register(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *handler) HandleVerifyEmail(c *fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณากรอกอีเมลและรหัสยืนยัน 6 หลักให้ครบถ้วน"})
	}
	res, err := h.service.VerifyEmail(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *handler) HandleResendVerification(c *fiber.Ctx) error {
	var req ResendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณากรอกอีเมลให้ถูกต้อง"})
	}
	res, err := h.service.ResendVerification(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func writeError(c *fiber.Ctx, err error) error {
	var cdErr *CooldownError
	if errors.As(err, &cdErr) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":               err.Error(),
			"retry_after_seconds": cdErr.RetryAfterSeconds,
		})
	}
	switch {
	case errors.Is(err, ErrEmailExists), errors.Is(err, ErrPhoneExists), errors.Is(err, ErrAlreadyVerified):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, ErrAccountRestricted):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, ErrInvalidVerification), errors.Is(err, ErrPendingAccountNotSet):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
}
