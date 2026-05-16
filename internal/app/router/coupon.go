package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/list_my"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/redeem"
	"github.com/gofiber/fiber/v2"
)

func SetupCouponRoutes(
	app *fiber.App,
	redeemHandler redeem.Handler,
	createHandler create.Handler,
	listMyHandler list_my.Handler,
) {
	v1 := app.Group("/api/v1/coupons", middleware.AuthMiddleware())
	
	v1.Post("/redeem", redeemHandler.Handle)
	v1.Get("/my", listMyHandler.Handle)
	v1.Post("/", createHandler.Handle) // Admin Only
}
