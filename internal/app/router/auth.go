package router

import (
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/auth"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/public/partner_apply"
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	sessions security.SessionStore,
	authHandler auth.Handler,
	partnerApplyHandler partner_apply.Handler,
) {
	v1 := app.Group("/api/v1/auth")
	v1.Post("/login", middleware.AuthRateLimiter(sessions, "bo_login", 30, time.Minute), authHandler.HandleLogin)
	v1.Post("/refresh", middleware.AuthRateLimiter(sessions, "bo_refresh", 40, time.Minute), authHandler.HandleRefresh)
	v1.Post("/logout", authMW, middleware.RequireScope(security.ScopeBackoffice), authHandler.HandleLogout)
	v1.Get("/me", authMW, middleware.RequireScope(security.ScopeBackoffice), authHandler.HandleMe)

	pub := app.Group("/api/v1/public")
	pub.Post("/partner-apply", middleware.AuthRateLimiter(sessions, "partner_apply", 10, time.Hour), partnerApplyHandler.HandleApply)
}
