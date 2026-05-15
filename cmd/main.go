package main

import (
	"log"
	"os"

	"github.com/Bangnus/Bidkan-backend/internal/app"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/database"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sms"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/tracking"

	_ "github.com/Bangnus/Bidkan-backend/docs" 

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

// @title Bidkan Backend API
// @version 1.0
// @description This is a sample server for Bidkan Clean Architecture.
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type 'Bearer ' followed by your JWT token.

func main() {
	// 1. Setup Infrastructure (Database, Redis)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=bidkan12345 dbname=bidkan_db sslmode=disable"
	}
	db := database.NewPostgresDB(dsn)
	defer db.Close()

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := database.NewRedisClient(redisAddr)
	
	// 2. Setup External Services (Firebase)
	var firebaseProvider domainService.SmsProvider
	firebaseKey := "configs/bidkan-service-account.json"
	if _, err := os.Stat(firebaseKey); err == nil {
		p, err := sms.NewFirebaseSmsProvider(firebaseKey)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to init Firebase: %v", err)
			firebaseProvider = sms.NewConsoleSmsProvider()
		} else {
			firebaseProvider = p
			log.Println("✅ Firebase Admin SDK initialized successfully")
		}
	} else {
		log.Println("ℹ️ Firebase key not found. Using Console Provider (Dev Mode)")
		firebaseProvider = sms.NewConsoleSmsProvider()
	}

	// 3. Dependency Injection Container (ย้าย Logic การสร้าง Service ไปไว้ที่นี่)
	container := app.NewContainer(db, firebaseProvider)

	// 4. Setup MQTT (แยกส่วนการทำงาน)
	bikeRepo := repository.NewBikePostgresRepository(db)
	bikeCache := repository.NewBikeRedisRepository(rdb)
	trackingService := tracking.NewService(bikeRepo, bikeCache)
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "tcp://localhost:1883"
	}
	mqttSub := mqtt.NewSubscriber(mqttBroker, "bidkan_backend_main", trackingService)
	if err := mqttSub.Start(); err != nil {
		log.Fatalf("Failed to start MQTT: %v", err)
	}

	// 5. Start Fiber Server
	server := fiber.New()
	server.Use(logger.New())

	server.Get("/swagger/*", swagger.HandlerDefault)
	
	// Setup Modules
	router.SetupUserRoutes(
		server, 
		container.CreateUserHandler, 
		container.LoginHandler, 
		container.MeHandler, 
		container.VerifyFirebaseHandler,
		container.UpdateProfileHandler,
		container.LogoutHandler,
	)
	router.SetupBikeRoutes(server, container.CreateBikeHandler, container.ListBikeHandler)
	router.SetupRideRoutes(server, container.StartRideHandler, container.EndRideHandler)
	router.SetupZoneRoutes(server, container.ListZoneHandler, container.CreateZoneHandler)
	router.SetupConfigRoutes(server, container.ConfigHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server is running on port %s", port)
	log.Fatal(server.Listen(":" + port))
}
