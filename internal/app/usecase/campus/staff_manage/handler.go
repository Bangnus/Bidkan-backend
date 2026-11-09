package staff_manage

import (
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateStaffRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
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
	staff, err := h.campusRepo.ListStaffByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": staff})
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	var req CreateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	if req.Username == "" || (req.PhoneNumber == "" && req.Email == "") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and email or phone_number are required"})
	}
	if req.Password == "" {
		req.Password = "password123"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	now := time.Now()
	user := &entity.User{
		ID:             uuid.New(),
		TenantID:       &tenant.ID,
		OrganizationID: tenant.OrganizationID,
		PhoneNumber:    req.PhoneNumber,
		Email:          req.Email,
		Username:       req.Username,
		Password:       string(hash),
		WalletBalance:  "0.00",
		Role:           "tenant_staff",
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.campusRepo.CreateStaffForTenant(c.Context(), user, tenant.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "staff member created",
		"data": entity.CampusStaff{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        user.Role,
			Status:      user.Status,
			CreatedAt:   user.CreatedAt,
		},
	})
}
