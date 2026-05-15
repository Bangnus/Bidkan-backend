package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/list"
	"github.com/gofiber/fiber/v2"
)

func SetupZoneRoutes(
	app *fiber.App,
	listHandler list.Handler,
	createHandler create.Handler,
) {
	v1 := app.Group("/api/v1/zones")

	v1.Get("/", listHandler.Handle)
	v1.Post("/", createHandler.Handle) // ควรใส่ Middleware Admin
}
