package main

import (
	"log"
	"os"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/database"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sms"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/tracking"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/otp"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify"

	_ "github.com/Bangnus/Bidkan-backend/docs" // ให้โหลดไฟล์ docs ที่จะถูกสร้างขึ้น

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // fiber-swagger middleware
)

// @title Bidkan Backend API
// @version 1.0
// @description This is a sample server for Bidkan Clean Architecture.
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.http bearer
// @name BearerAuth
// @description Paste your JWT token ONLY (The 'Bearer ' prefix will be added automatically)

func main() {
	// 1. ต่อ Database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=yourpassword dbname=bidkan_db sslmode=disable"
	}
	db := database.NewPostgresDB(dsn)
	defer db.Close()

	// --- Redis Setup ---
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := database.NewRedisClient(redisAddr)
	bikeCache := repository.NewBikeRedisRepository(rdb)
	otpRepo := repository.NewOtpRedisRepository(rdb)
	// ------------------

	// 2. Dependency Injection
	// --- SMS Setup (เลือกสลับสายตรงนี้ได้เลย) ---
	smsProvider := sms.NewConsoleSmsProvider() // ตอนนี้ใช้แบบ Console (ฟรี)
	// smsProvider := sms.NewFirebaseSmsProvider(os.Getenv("FIREBASE_KEY")) // เปลี่ยนมาใช้ตัวนี้เมื่อพร้อม
	
	// --- User Setup ---
	userRepo := repository.NewUserPostgresRepository(db)
	
	// OTP
	otpService := otp.NewService(otpRepo, smsProvider)
	otpHandler := otp.NewHandler(otpService)
	
	// Create User
	createService := create.NewService(userRepo, otpRepo, smsProvider)
	createHandler := create.NewHandler(createService)
	
	// Verify User
	verifyService := verify.NewService(userRepo, otpRepo)
	verifyHandler := verify.NewHandler(verifyService)

	// Login User
	loginService := login.NewService(userRepo)
	loginHandler := login.NewHandler(loginService)

	// Me Profile (Protected)
	meService := me.NewService(userRepo)
	meHandler := me.NewHandler(meService)
	// ------------------

	// --- MQTT Setup ---
	bikeRepo := repository.NewBikePostgresRepository(db)
	trackingService := tracking.NewService(bikeRepo, bikeCache)
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
	app.Use(logger.New())

	// 4. ตั้งค่า Routes
	app.Get("/swagger/*", swagger.HandlerDefault)
	router.SetupUserRoutes(app, createHandler, otpHandler, verifyHandler, loginHandler, meHandler)

	// 5. เปิด Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Server is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
