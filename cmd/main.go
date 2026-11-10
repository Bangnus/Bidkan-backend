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
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sms"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/tracking"

	_ "github.com/Bangnus/Bidkan-backend/docs"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
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
	_ = godotenv.Load()
	secCfg := security.MustLoadConfig()

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
	sessions := security.NewRedisSessionStore(rdb)

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
			opt := option.WithCredentialsFile(firebaseKey)
			fbApp, _ := firebase.NewApp(context.Background(), nil, opt)
			nP, _ := notification.NewFirebaseNotificationProvider(fbApp)
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

	// 4. Dependency Injection Container & Auth Middleware
	stripeClient := infraStripe.NewClient(infraStripe.LoadConfigFromEnv())
	container := app.NewContainer(db, rdb, secCfg, sessions, smsProvider, notiProvider, mqttPub, stripeClient)
	authDeps := middleware.AuthDeps{
		Config:   secCfg,
		Sessions: sessions,
		UserRepo: container.UserRepo,
	}
	authMW := middleware.AuthMiddleware(authDeps)
	optAuthMW := middleware.OptionalAuthMiddleware(authDeps)

	// 5. Setup MQTT Subscriber (สำหรับรับพิกัด)
	bikeRepo := repository.NewBikePostgresRepository(db)
	bikeCache := repository.NewBikeRedisRepository(rdb)
	trackingService := tracking.NewService(bikeRepo, bikeCache)
	mqttSub := mqtt.NewSubscriber(mqttBroker, "bidkan_backend_sub", trackingService)
	if err := mqttSub.Start(); err != nil {
		log.Fatalf("Failed to start MQTT Subscriber: %v", err)
	}

	// 6. Start Fiber Server
	server := fiber.New(fiber.Config{
		BodyLimit: 4 * 1024 * 1024,
	})
	server.Use(recover.New())
	server.Use(helmet.New())
	server.Use(cors.New(cors.Config{
		AllowOrigins:     secCfg.CORSAllowedOrigins,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))
	server.Use(logger.New())

	if secCfg.EnableSwagger {
		server.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Setup Multi-Tenant Module
	tenantModule := app.NewTenantModule(db, secCfg, sessions, stripeClient)
	tenantModule.SeedAndRegisterRoutes(server, db, authMW, sessions)
	router.SetupStripeWebhookRoutes(server, container.StripeWebhookHandler)
	app.StartAutoPayoutScheduler(context.Background(), stripeClient, tenantModule.PayoutManageHandler.Service())

	// Setup Modules
	router.SetupUserRoutes(
		server,
		authMW,
		sessions,
		container.CreateUserHandler,
		container.LoginHandler,
		container.GoogleLoginHandler,
		container.RefreshUserHandler,
		container.MeHandler,
		container.VerifyFirebaseHandler,
		container.UpdateProfileHandler,
		container.LogoutHandler,
		container.RankHandler,
	)
	router.SetupBikeRoutes(server, authMW, optAuthMW, container.CreateBikeHandler, container.ListBikeHandler, container.UpdateBikeStatusHandler)
	router.SetupRideRoutes(server, authMW, container.StartRideHandler, container.EndRideHandler)
	router.SetupZoneRoutes(server, authMW, container.ListZoneHandler, container.CreateZoneHandler)
	router.SetupReportRoutes(server, authMW, container.ReportSummaryHandler)
	router.SetupCouponRoutes(server, authMW, container.RedeemCouponHandler, container.CreateCouponHandler, container.ListMyCouponHandler)
	router.SetupNotificationRoutes(server, authMW, container.NotiBroadcastHandler, container.NotiListHandler, container.NotiUpdateTokenHandler)
	router.SetupConfigRoutes(server, authMW, container.ConfigHandler)
	router.SetupWalletRoutes(server, authMW, container.StripeTopupHandler, container.TransferHandler, container.VerifyReceiverHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server is running on port %s", port)
	log.Fatal(server.Listen(":" + port))
}
