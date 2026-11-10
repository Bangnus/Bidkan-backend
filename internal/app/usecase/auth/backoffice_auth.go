package auth

import (
	"strings"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CampusRef struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

type Handler struct {
	userRepo   domainRepo.UserRepository
	tenantRepo domainRepo.TenantRepository
	orgRepo    domainRepo.OrganizationRepository
	secCfg     *security.Config
	sessions   security.SessionStore
}

func NewHandler(
	userRepo domainRepo.UserRepository,
	tenantRepo domainRepo.TenantRepository,
	orgRepo domainRepo.OrganizationRepository,
	secCfg *security.Config,
	sessions security.SessionStore,
) Handler {
	return Handler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		orgRepo:    orgRepo,
		secCfg:     secCfg,
		sessions:   sessions,
	}
}

func (h *Handler) HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}
	if identifier == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password are required"})
	}

	lockKey := "bo:" + strings.ToLower(identifier)
	if h.sessions != nil {
		if locked, retryAfter, _ := h.sessions.IsLoginLocked(c.Context(), lockKey); locked {
			setRetryHeader(c, retryAfter)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":               "บัญชีถูกล็อกชั่วคราวเนื่องจากพยายามเข้าสู่ระบบผิดหลายครั้ง กรุณารอสักครู่",
				"retry_after_seconds": int(retryAfter.Seconds()) + 1,
			})
		}
	}

	user, err := h.userRepo.GetByIdentifier(c.Context(), identifier)
	if err != nil {
		h.recordFailure(c.Context(), lockKey)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.recordFailure(c.Context(), lockKey)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
	}
	if !isBackofficeRole(user.Role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "ไม่มีสิทธิ์เข้าใช้งานระบบ Backoffice"})
	}
	if user.Status != "active" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "บัญชีอยู่ระหว่างรอการอนุมัติหรือถูกระงับใช้งาน"})
	}

	if h.sessions != nil {
		_ = h.sessions.ClearLoginFailures(c.Context(), lockKey)
	}

	pair, profile, err := h.issueBackofficeSession(c.Context(), user, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}

	profile["token"] = pair.AccessToken
	profile["access_token"] = pair.AccessToken
	profile["refresh_token"] = pair.RefreshToken
	profile["expires_in"] = pair.ExpiresIn
	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h *Handler) HandleMe(c *fiber.Ctx) error {
	uid, ok := middleware.GetAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	user, err := h.userRepo.GetByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	profile, _ := h.buildAuthProfile(c.Context(), user)
	return c.Status(fiber.StatusOK).JSON(profile)
}
