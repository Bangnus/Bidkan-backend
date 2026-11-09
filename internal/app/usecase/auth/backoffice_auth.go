package auth

import (
	"os"
	"time"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Handler struct {
	userRepo   domainRepo.UserRepository
	tenantRepo domainRepo.TenantRepository
}

func NewHandler(userRepo domainRepo.UserRepository, tenantRepo domainRepo.TenantRepository) Handler {
	return Handler{userRepo: userRepo, tenantRepo: tenantRepo}
}

func (h *Handler) HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password are required"})
	}

	user, err := h.userRepo.GetByUsername(c.Context(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	if user.Role == "user" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "backoffice access denied for end-users"})
	}

	var tenantIDStr, tenantSlug, tenantName string
	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
		if tenant, err := h.tenantRepo.GetByID(c.Context(), *user.TenantID); err == nil {
			tenantSlug = tenant.Slug
			tenantName = tenant.Name
		}
	}

	token, err := generateToken(user.ID.String(), user.Role, tenantIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":       token,
		"username":    user.Username,
		"role":        user.Role,
		"tenant_id":   tenantIDStr,
		"tenant_slug": tenantSlug,
		"tenant_name": tenantName,
	})
}

func (h *Handler) HandleMe(c *fiber.Ctx) error {
	uidStr, _ := c.Locals("user_id").(string)
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user, err := h.userRepo.GetByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	var tenantIDStr, tenantSlug, tenantName string
	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
		if tenant, err := h.tenantRepo.GetByID(c.Context(), *user.TenantID); err == nil {
			tenantSlug = tenant.Slug
			tenantName = tenant.Name
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"tenant_id":   tenantIDStr,
		"tenant_slug": tenantSlug,
		"tenant_name": tenantName,
	})
}

func generateToken(userID, role, tenantID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "bidkan_secret_key_2024"
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	}
	if tenantID != "" {
		claims["tenant_id"] = tenantID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
