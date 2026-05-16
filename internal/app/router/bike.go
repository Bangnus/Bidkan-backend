package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/update_status"
	"github.com/gofiber/fiber/v2"
)

func SetupBikeRoutes(
	app *fiber.App,
	createHandler create.Handler,
	listHandler list.Handler,
	updateStatusHandler update_status.Handler,
) {
	v1 := app.Group("/api/v1/bikes", middleware.AuthMiddleware())
	
	v1.Get("/", listHandler.HandleList)
	v1.Post("/", createHandler.Handle) // Admin Only 
	v1.Patch("/:id/status", updateStatusHandler.Handle) // Staff/Admin Only

}
