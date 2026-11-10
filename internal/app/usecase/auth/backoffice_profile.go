package auth

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
)

type scopeMeta struct {
	tenantID     string
	tenantSlug   string
	orgID        string
	orgSlug      string
	allowedSlugs []string
}

func (h *Handler) buildAuthProfile(ctx context.Context, user *entity.User) (fiber.Map, scopeMeta) {
	var tenantIDStr, tenantSlug, tenantName string
	var orgIDStr, orgSlug, orgName string
	campuses := make([]CampusRef, 0)
	allowedSlugs := make([]string, 0)

	switch user.Role {
	case "org_admin":
		orgID := user.OrganizationID
		if orgID == nil && user.TenantID != nil {
			if tenant, err := h.tenantRepo.GetByID(ctx, *user.TenantID); err == nil {
				orgID = tenant.OrganizationID
			}
		}
		if orgID != nil && h.orgRepo != nil {
			orgIDStr = orgID.String()
			if org, err := h.orgRepo.GetByID(ctx, *orgID); err == nil {
				orgSlug = org.Slug
				orgName = org.Name
			}
			if list, err := h.tenantRepo.ListByOrganization(ctx, *orgID); err == nil {
				for _, t := range list {
					campuses = append(campuses, CampusRef{ID: t.ID, Slug: t.Slug, Name: t.Name, Status: t.Status})
					allowedSlugs = append(allowedSlugs, t.Slug)
				}
			}
		}
		if len(campuses) > 0 {
			tenantIDStr = campuses[0].ID.String()
			tenantSlug = campuses[0].Slug
			tenantName = campuses[0].Name
		}
	case "tenant_admin", "tenant_staff":
		if user.TenantID != nil {
			tenantIDStr = user.TenantID.String()
			if tenant, err := h.tenantRepo.GetByID(ctx, *user.TenantID); err == nil {
				tenantSlug = tenant.Slug
				tenantName = tenant.Name
				allowedSlugs = []string{tenant.Slug}
			}
		}
	case "super_admin":
		// Platform owner has no campus or organization scope
	}

	resp := fiber.Map{
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
		"allowed_slugs":   allowedSlugs,
	}
	meta := scopeMeta{
		tenantID:     tenantIDStr,
		tenantSlug:   tenantSlug,
		orgID:        orgIDStr,
		orgSlug:      orgSlug,
		allowedSlugs: allowedSlugs,
	}
	return resp, meta
}

func (h *Handler) recordFailure(ctx context.Context, key string) {
	if h.sessions == nil {
		return
	}
	_, _ = h.sessions.RecordLoginFailure(ctx, key, 5, 5*time.Minute, 5*time.Minute)
}

func isBackofficeRole(role string) bool {
	return role == "super_admin" || role == "org_admin" || role == "tenant_admin" || role == "tenant_staff"
}

