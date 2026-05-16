package app

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	
	// User
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/logout"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/rank"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"
	
	// Bike
	bikeCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	bikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	bikeUpdateStatus "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/update_status"
	
	// Ride
	rideStart "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/start"
	rideEnd "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/end"

	// Zone
	zoneList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/list"
	zoneCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/create"

	// Report
	reportSummary "github.com/Bangnus/Bidkan-backend/internal/app/usecase/report/summary"

	// Wallet & Payment
	"github.com/Bangnus/Bidkan-backend/pkg/payment"
	walletTopup "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	walletTransfer "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	walletVerify "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"
	walletWebhook "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"

	// Config
	configUsecase "github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
)

type Container struct {
	// User
	CreateUserHandler     create.Handler
	VerifyFirebaseHandler verify_firebase.Handler
	LoginHandler          login.Handler
	MeHandler             me.Handler
	RankHandler           rank.Handler
	UpdateProfileHandler  update_profile.Handler
	LogoutHandler         logout.Handler

	// Wallet
	TopupHandler          walletTopup.Handler
	TransferHandler       walletTransfer.Handler
	VerifyReceiverHandler walletVerify.Handler
	WebhookHandler        walletWebhook.Handler

	// Bike
	CreateBikeHandler     bikeCreate.Handler
	ListBikeHandler       bikeList.Handler
	UpdateBikeStatusHandler bikeUpdateStatus.Handler

	// Ride
	StartRideHandler      rideStart.Handler
	EndRideHandler        rideEnd.Handler

	// Zone
	ListZoneHandler       zoneList.Handler
	CreateZoneHandler     zoneCreate.Handler

	// Report
	ReportSummaryHandler  reportSummary.Handler

	// Config
	ConfigHandler         configUsecase.Handler
}

func NewContainer(db *sql.DB, rdb *redis.Client, firebaseProvider service.SmsProvider) *Container {
	// --- Repositories ---
	userRepo := repository.NewUserPostgresRepository(db)
	bikeRepo := repository.NewBikePostgresRepository(db)
	rideRepo := repository.NewRidePostgresRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	configRepo := repository.NewConfigRepository(db)
	txRepo := repository.NewTransactionPostgresRepository(db)
	spendingRepo := repository.NewUserMonthlySpendingRepository(db)
	reportRepo := repository.NewReportPostgresRepository(db)

	// --- Cache (Redis) ---
	rankCacheRepo := repository.NewUserRankRedisRepository(rdb)
	configCacheRepo := repository.NewConfigRedisRepository(rdb)

	// --- Services ---
	// User
	createService := create.NewService(userRepo)
	verifyFirebaseService := verify_firebase.NewService(userRepo, firebaseProvider)
	loginService := login.NewService(userRepo)
	meService := me.NewService(userRepo)
	rankService := rank.NewService(spendingRepo, rankCacheRepo)
	updateProfileService := update_profile.NewService(userRepo)

	// Payment
	paySvc := payment.NewPaySolutionsService()

	// Wallet
	topupService := walletTopup.NewService(txRepo, userRepo, paySvc)
	transferService := walletTransfer.NewService(db, userRepo, txRepo)
	verifyService := walletVerify.NewService(userRepo)
	webhookService := walletWebhook.NewService(txRepo, userRepo)

	// Bike
	bikeCreateService := bikeCreate.NewService(bikeRepo)
	bikeListService := bikeList.NewService(bikeRepo)
	bikeUpdateStatusService := bikeUpdateStatus.NewService(bikeRepo)

	// Ride
	rideStartService := rideStart.NewService(rideRepo, userRepo, bikeRepo)
	rideEndService := rideEnd.NewService(rideRepo, userRepo, bikeRepo, zoneRepo, configRepo, spendingRepo, rankCacheRepo, configCacheRepo)

	// Zone
	zoneListService := zoneList.NewService(zoneRepo)
	zoneCreateService := zoneCreate.NewService(zoneRepo)

	// Report
	reportSummaryService := reportSummary.NewService(reportRepo)

	// Config
	configService := configUsecase.NewService(configRepo)

	// --- Handlers ---
	return &Container{
		CreateUserHandler:     create.NewHandler(createService),
		VerifyFirebaseHandler: verify_firebase.NewHandler(verifyFirebaseService),
		LoginHandler:          login.NewHandler(loginService),
		MeHandler:             me.NewHandler(meService),
		RankHandler:           rank.NewHandler(rankService),
		UpdateProfileHandler:  update_profile.NewHandler(updateProfileService),
		LogoutHandler:         logout.NewHandler(),

		// Wallet
		TopupHandler:          walletTopup.NewHandler(topupService),
		TransferHandler:       walletTransfer.NewHandler(transferService),
		VerifyReceiverHandler: walletVerify.NewHandler(verifyService),
		WebhookHandler:        walletWebhook.NewHandler(webhookService),

		// Bike
		CreateBikeHandler:     bikeCreate.NewHandler(bikeCreateService),

		ListBikeHandler:       bikeList.NewHandler(bikeListService),
		UpdateBikeStatusHandler: bikeUpdateStatus.NewHandler(bikeUpdateStatusService),

		StartRideHandler:      rideStart.NewHandler(rideStartService),
		EndRideHandler:        rideEnd.NewHandler(rideEndService),

		ListZoneHandler:       zoneList.NewHandler(zoneListService),
		CreateZoneHandler:     zoneCreate.NewHandler(zoneCreateService),

		ReportSummaryHandler:  reportSummary.NewHandler(reportSummaryService),

		ConfigHandler:         configUsecase.NewHandler(configService),
	}
}
