package router

import (
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/google_login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/logout"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/rank"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/refresh"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"
	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	sessions security.SessionStore,
	createHandler create.Handler,
	loginHandler login.Handler,
	googleLoginHandler google_login.Handler,
	refreshHandler refresh.Handler,
	meHandler me.Handler,
	verifyFirebaseHandler verify_firebase.Handler,
	updateProfileHandler update_profile.Handler,
	logoutHandler logout.Handler,
	rankHandler rank.Handler,
) {
	v1 := app.Group("/api/v1")

	// Public Routes with rate limits
	v1.Post("/users", middleware.AuthRateLimiter(sessions, "user_register", 15, time.Minute), createHandler.Handle)
	v1.Post("/users/verify/firebase", middleware.AuthRateLimiter(sessions, "user_verify", 15, time.Minute), verifyFirebaseHandler.Handle)
	v1.Post("/users/login", middleware.AuthRateLimiter(sessions, "user_login", 30, time.Minute), loginHandler.Handle)
	v1.Post("/users/google-login", middleware.AuthRateLimiter(sessions, "user_google", 30, time.Minute), googleLoginHandler.Handle)
	v1.Post("/auth/google", middleware.AuthRateLimiter(sessions, "user_google", 30, time.Minute), googleLoginHandler.Handle)
	v1.Post("/users/refresh", middleware.AuthRateLimiter(sessions, "user_refresh", 40, time.Minute), refreshHandler.Handle)

	// Protected Routes (Rider Scope)
	userGroup := v1.Group("/users", authMW, middleware.RequireScope(security.ScopeRider))
	userGroup.Get("/me", meHandler.Handle)
	userGroup.Get("/rank", rankHandler.Handle)
	userGroup.Patch("/profile/image", updateProfileHandler.Handle)
	userGroup.Post("/logout", logoutHandler.Handle)
}
