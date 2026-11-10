package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/broadcast"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/update_token"
	"github.com/gofiber/fiber/v2"
)

func SetupNotificationRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	broadcastHandler broadcast.Handler,
	listHandler list.Handler,
	updateTokenHandler update_token.Handler,
) {
	v1 := app.Group("/api/v1")

	// User Routes (Rider Scope)
	riderNotifs := v1.Group("/notifications", authMW, middleware.RequireScope(security.ScopeRider))
	riderNotifs.Get("/", listHandler.Handle)
	riderNotifs.Post("/fcm-token", updateTokenHandler.Handle)

	// Admin Routes (Backoffice Scope)
	admin := v1.Group("/admin", authMW, middleware.RequireScope(security.ScopeBackoffice), middleware.BackofficeRoleGuard("super_admin", "org_admin", "tenant_admin"))
	admin.Post("/notifications/broadcast", broadcastHandler.Handle)
}
