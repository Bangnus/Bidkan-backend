package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/otp"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(
	app *fiber.App,
	createHandler create.Handler,
	otpHandler otp.Handler,
	verifyHandler verify.Handler,
	loginHandler login.Handler,
	meHandler me.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Public Routes
	v1.Post("/users", createHandler.Handle)
	v1.Post("/users/otp", otpHandler.HandleSend)
	v1.Post("/users/verify", verifyHandler.Handle)
	v1.Post("/users/login", loginHandler.Handle)

	// Protected Routes (ต้องใช้ Token)
	userGroup := v1.Group("/users", middleware.AuthMiddleware())
	userGroup.Get("/me", meHandler.Handle)
}