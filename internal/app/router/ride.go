package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/end"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/start"
	"github.com/gofiber/fiber/v2"
)

func SetupRideRoutes(
	app *fiber.App,
	startHandler start.Handler,
	endHandler end.Handler,
) {
	v1 := app.Group("/api/v1/rides", middleware.AuthMiddleware())

	v1.Post("/start", startHandler.Handle)
	v1.Post("/end", endHandler.Handle)
}
