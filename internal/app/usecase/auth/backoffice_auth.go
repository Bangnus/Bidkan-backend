package auth

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
}

func NewHandler(userRepo domainRepo.UserRepository, tenantRepo domainRepo.TenantRepository, orgRepo domainRepo.OrganizationRepository) Handler {
	return Handler{userRepo: userRepo, tenantRepo: tenantRepo, orgRepo: orgRepo}
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

	user, err := h.userRepo.GetByIdentifier(c.Context(), identifier)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
	}
	if user.Role == "user" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "ไม่มีสิทธิ์เข้าใช้งานระบบ Backoffice"})
	}
	if user.Status != "active" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "บัญชีอยู่ระหว่างรอการอนุมัติหรือถูกระงับใช้งาน"})
	}

	profile := h.buildAuthProfile(c.Context(), user)
	orgIDStr, _ := profile["organization_id"].(string)
	tenantIDStr, _ := profile["tenant_id"].(string)

	token, err := generateToken(user.ID.String(), user.Role, tenantIDStr, orgIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}
	profile["token"] = token
	return c.Status(fiber.StatusOK).JSON(profile)
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
	return c.Status(fiber.StatusOK).JSON(h.buildAuthProfile(c.Context(), user))
}

func (h *Handler) buildAuthProfile(ctx context.Context, user *entity.User) fiber.Map {
	var tenantIDStr, tenantSlug, tenantName string
	var orgIDStr, orgSlug, orgName string
	campuses := make([]CampusRef, 0)

	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
		if tenant, err := h.tenantRepo.GetByID(ctx, *user.TenantID); err == nil {
			tenantSlug = tenant.Slug
			tenantName = tenant.Name
			if user.OrganizationID == nil && tenant.OrganizationID != nil {
				user.OrganizationID = tenant.OrganizationID
			}
		}
	}

	if user.OrganizationID != nil && h.orgRepo != nil {
		orgIDStr = user.OrganizationID.String()
		if org, err := h.orgRepo.GetByID(ctx, *user.OrganizationID); err == nil {
			orgSlug = org.Slug
			orgName = org.Name
		}
		if list, err := h.tenantRepo.ListByOrganization(ctx, *user.OrganizationID); err == nil {
			for _, t := range list {
				campuses = append(campuses, CampusRef{ID: t.ID, Slug: t.Slug, Name: t.Name, Status: t.Status})
			}
		}
	}
	if len(campuses) == 0 && (user.Role == "super_admin" || user.Role == "admin") {
		if list, err := h.tenantRepo.List(ctx); err == nil {
			for _, t := range list {
				campuses = append(campuses, CampusRef{ID: t.ID, Slug: t.Slug, Name: t.Name, Status: t.Status})
			}
		}
	}
	if len(campuses) > 0 && tenantSlug == "" {
		tenantIDStr = campuses[0].ID.String()
		tenantSlug = campuses[0].Slug
		tenantName = campuses[0].Name
	}

	return fiber.Map{
		"id":              user.ID,
		"email":           user.Email,
		"username":        user.Username,
		"role":            user.Role,
		"tenant_id":       tenantIDStr,
		"tenant_slug":     tenantSlug,
		"tenant_name":     tenantName,
		"organization_id": orgIDStr,
		"org_slug":        orgSlug,
		"org_name":        orgName,
		"campuses":        campuses,
	}
}

func generateToken(userID, role, tenantID, orgID string) (string, error) {
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
	if orgID != "" {
		claims["organization_id"] = orgID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
