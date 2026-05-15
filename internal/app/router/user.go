package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/logout"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/rank"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(
	app *fiber.App,
	createHandler create.Handler,
	loginHandler login.Handler,
	meHandler me.Handler,
	verifyFirebaseHandler verify_firebase.Handler,
	updateProfileHandler update_profile.Handler,
	logoutHandler logout.Handler,
	rankHandler rank.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Public Routes
	v1.Post("/users", createHandler.Handle)
	v1.Post("/users/verify/firebase", verifyFirebaseHandler.Handle)
	v1.Post("/users/login", loginHandler.Handle)

	// Protected Routes (ต้องใช้ Token)
	userGroup := v1.Group("/users", middleware.AuthMiddleware())
	userGroup.Get("/me", meHandler.Handle)
	userGroup.Get("/rank", rankHandler.Handle)
	userGroup.Patch("/profile/image", updateProfileHandler.Handle)
	userGroup.Post("/logout", logoutHandler.Handle)
}