package main

import (
	"log"
	"os"

	"bidkan/internal/app/infrastructure/database"
	"bidkan/internal/app/infrastructure/mqtt"
	"bidkan/internal/app/infrastructure/repository"
	"bidkan/internal/app/router"
	"bidkan/internal/app/usecase/bike/tracking"
	"bidkan/internal/app/usecase/user/create"

	_ "bidkan/docs" // ให้โหลดไฟล์ docs ที่จะถูกสร้างขึ้น

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // fiber-swagger middleware
)

// @title Bidkan Backend API
// @version 1.0
// @description This is a sample server for Bidkan Clean Architecture.
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// 1. ต่อ Database (อ่าน DSN จาก Environment Variable)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=yourpassword dbname=bidkan_db sslmode=disable"
	}
	db := database.NewPostgresDB(dsn)
	defer db.Close() // ปิด connection เมื่อโปรแกรมดับ

	// 2. Dependency Injection (ต่อจิ๊กซอว์จากล่างขึ้นบน)
	// สร้าง Repo -> ส่งให้ Service -> ส่งให้ Handler
	userRepo := repository.NewUserPostgresRepository(db)
	createService := create.NewService(userRepo)
	createHandler := create.NewHandler(createService)

	// --- MQTT Setup ---
	trackingService := tracking.NewService()
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "tcp://localhost:1883"
	}
	mqttSub := mqtt.NewSubscriber(mqttBroker, "bidkan_backend_main", trackingService)
	if err := mqttSub.Start(); err != nil {
		log.Fatalf("Failed to start MQTT: %v", err)
	}
	// ------------------

	// 3. สร้าง Fiber App
	app := fiber.New()
	app.Use(logger.New()) // แสดง Log ใน Terminal

	// 4. ตั้งค่า Routes
	app.Get("/swagger/*", swagger.HandlerDefault) // default: http://localhost:8080/swagger/index.html
	router.SetupUserRoutes(app, createHandler)

	// 5. เปิด Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Server is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
