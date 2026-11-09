package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/auth"
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App, authHandler auth.Handler) {
	v1 := app.Group("/api/v1/auth")

	v1.Post("/login", authHandler.HandleLogin)
	v1.Get("/me", middleware.AuthMiddleware(), authHandler.HandleMe)
}
