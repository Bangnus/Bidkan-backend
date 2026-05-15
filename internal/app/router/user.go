package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/otp"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App, createHandler create.Handler, otpHandler otp.Handler, verifyHandler verify.Handler) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// V1 Routes
	v1.Post("/users", createHandler.Handle)
	v1.Post("/users/otp", otpHandler.HandleSend)
	v1.Post("/users/verify", verifyHandler.Handle)
}