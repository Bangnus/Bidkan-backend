package tenant_manage

import (
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateAdminRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	Role        string `json:"role"` // tenant_admin | org_admin
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) HandleListAdmins(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	admins, err := h.campusRepo.ListAdminsByTenant(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": admins})
}

func (h *Handler) HandleCreateAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tenant not found"})
	}
	var req CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password are required"})
	}
	if req.Username == "" {
		req.Username = strings.Split(req.Email, "@")[0]
	}
	role := strings.TrimSpace(req.Role)
	if role != "tenant_admin" && role != "org_admin" {
		role = "tenant_admin"
	}
	if role == "org_admin" && tenant.OrganizationID == nil {
		role = "tenant_admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}
	now := time.Now()
	u := &entity.User{
		ID:             uuid.New(),
		TenantID:       &tenant.ID,
		OrganizationID: tenant.OrganizationID,
		Email:          req.Email,
		PhoneNumber:    strings.TrimSpace(req.PhoneNumber),
		Username:       req.Username,
		Password:       string(hash),
		WalletBalance:  "0.00",
		Role:           role,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.campusRepo.CreateBackofficeUser(c.Context(), u); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "university admin account created",
		"data": entity.CampusStaff{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
			Role:        u.Role,
			Status:      u.Status,
			CreatedAt:   u.CreatedAt,
		},
	})
}

func (h *Handler) HandleListOrders(c *fiber.Ctx) error {
	orders, err := h.orderRepo.ListAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": orders})
}

func (h *Handler) HandleUpdateOrderStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil || req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status is required"})
	}
	switch req.Status {
	case "pending", "approved", "shipped", "completed", "cancelled":
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status value"})
	}
	if err := h.orderRepo.UpdateStatus(c.Context(), id, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "hardware order status updated"})
}
