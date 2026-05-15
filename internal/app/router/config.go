package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
	"github.com/gofiber/fiber/v2"
)

func SetupConfigRoutes(
	app *fiber.App,
	handler config.Handler,
) {
	v1 := app.Group("/api/v1/configs")

	v1.Post("/", handler.HandleSet)
	v1.Get("/:key", handler.HandleGet)
}
