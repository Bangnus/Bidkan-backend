package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	"github.com/gofiber/fiber/v2"
)

func SetupCampusRoutes(
	app *fiber.App,
	dashboardHandler dashboard.Handler,
	bikeListHandler bike_list.Handler,
	fieldOpsHandler field_ops.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")
	campus := v1.Group("/campus/:slug")

	campus.Get("/dashboard", dashboardHandler.Handle)
	campus.Get("/bikes", bikeListHandler.Handle)
	campus.Get("/field/queue", fieldOpsHandler.HandleGetQueue)
	campus.Post("/field/swap-battery", fieldOpsHandler.HandleSwapBattery)
}

