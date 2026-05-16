package main

import (
	"context"
	"log"
	"os"

	"github.com/Bangnus/Bidkan-backend/internal/app"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/database"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/notification"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sms"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/tracking"

	_ "github.com/Bangnus/Bidkan-backend/docs" 

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"google.golang.org/api/option"
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
	var smsProvider domainService.SmsProvider
	var notiProvider domainService.NotificationProvider

	firebaseKey := "configs/bidkan-service-account.json"
	if _, err := os.Stat(firebaseKey); err == nil {
		sP, err := sms.NewFirebaseSmsProvider(firebaseKey)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to init Firebase SMS: %v", err)
			smsProvider = sms.NewConsoleSmsProvider()
			notiProvider = notification.NewConsoleNotificationProvider()
		} else {
			smsProvider = sP
			// ดึง Firebase App จาก SMS Provider เพื่อสร้าง Notification Provider
			// หมายเหตุ: ผมจะไปแก้ NewFirebaseSmsProvider ให้คืนค่า App มาด้วย หรือใช้ท่าอื่น
			// เพื่อความง่าย ผมจะสร้าง App แยกตรงนี้เลยครับ
			opt := option.WithCredentialsFile(firebaseKey)
			app, _ := firebase.NewApp(context.Background(), nil, opt)
			nP, _ := notification.NewFirebaseNotificationProvider(app)
			notiProvider = nP
			log.Println("✅ Firebase Admin SDK initialized successfully")
		}
	} else {
		log.Println("ℹ️ Firebase key not found. Using Console Providers (Dev Mode)")
		smsProvider = sms.NewConsoleSmsProvider()
		notiProvider = notification.NewConsoleNotificationProvider()
	}

	// 3. Setup MQTT
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "tcp://localhost:1883"
	}
	
	mqttPub, err := mqtt.NewPublisher(mqttBroker, "bidkan_backend_pub")
	if err != nil {
		log.Printf("⚠️ Warning: Failed to init MQTT Publisher: %v", err)
	}

	// 4. Dependency Injection Container
	container := app.NewContainer(db, rdb, smsProvider, notiProvider, mqttPub)

	// 5. Setup MQTT Subscriber (สำหรับรับพิกัด)
	bikeRepo := repository.NewBikePostgresRepository(db)
	bikeCache := repository.NewBikeRedisRepository(rdb)
	trackingService := tracking.NewService(bikeRepo, bikeCache)
	mqttSub := mqtt.NewSubscriber(mqttBroker, "bidkan_backend_sub", trackingService)
	if err := mqttSub.Start(); err != nil {
		log.Fatalf("Failed to start MQTT Subscriber: %v", err)
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
		container.RankHandler,
	)
	router.SetupBikeRoutes(server, container.CreateBikeHandler, container.ListBikeHandler, container.UpdateBikeStatusHandler)
	router.SetupRideRoutes(server, container.StartRideHandler, container.EndRideHandler)
	router.SetupZoneRoutes(server, container.ListZoneHandler, container.CreateZoneHandler)
	router.SetupReportRoutes(server, container.ReportSummaryHandler)
	router.SetupCouponRoutes(server, container.RedeemCouponHandler, container.CreateCouponHandler, container.ListMyCouponHandler)
	router.SetupNotificationRoutes(server, container.NotiBroadcastHandler, container.NotiListHandler, container.NotiUpdateTokenHandler)
	router.SetupConfigRoutes(server, container.ConfigHandler)
	router.SetupWalletRoutes(server, container.TopupHandler, container.WebhookHandler, container.TransferHandler, container.VerifyReceiverHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server is running on port %s", port)
	log.Fatal(server.Listen(":" + port))
}
