package router

import (
	"bidkan/internal/app/usecase/user/create"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App, createHandler create.Handler) {
	api := app.Group("/api/v1")

	// ผูก Route เข้ากับ Handler
	api.Post("/users", createHandler.Handle)
}