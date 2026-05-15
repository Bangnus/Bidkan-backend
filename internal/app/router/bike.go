package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	"github.com/gofiber/fiber/v2"
)

func SetupBikeRoutes(
	app *fiber.App,
	createHandler create.Handler,
	listHandler list.Handler,
) {
	v1 := app.Group("/api/v1/bikes")

	v1.Get("/", listHandler.HandleListAll)
	v1.Get("/available", listHandler.HandleListAvailable)
	v1.Post("/", createHandler.Handle) // ควรมี Middleware Admin คุม
}
