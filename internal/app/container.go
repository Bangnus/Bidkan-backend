package app

import (
	"database/sql"

	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	bikeCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	bikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	bikeUpdateStatus "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/update_status"
	configUsecase "github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
	couponCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/create"
	couponListMy "github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/list_my"
	couponRedeem "github.com/Bangnus/Bidkan-backend/internal/app/usecase/coupon/redeem"
	notiBroadcast "github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/broadcast"
	notiList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/list"
	notiUpdateToken "github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/update_token"
	reportSummary "github.com/Bangnus/Bidkan-backend/internal/app/usecase/report/summary"
	rideEnd "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/end"
	rideStart "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/start"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/google_login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/logout"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/rank"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"
	walletTopup "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	walletTransfer "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	walletVerify "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"
	walletWebhook "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"
	zoneCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/create"
	zoneList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/list"
	"github.com/Bangnus/Bidkan-backend/pkg/payment"
	"github.com/redis/go-redis/v9"
)

func NewContainer(
	db *sql.DB,
	rdb *redis.Client,
	smsProvider domainService.SmsProvider,
	notiProvider domainService.NotificationProvider,
	mqttPub mqtt.Publisher,
) *Container {
	userRepo := repository.NewUserPostgresRepository(db)
	bikeRepo := repository.NewBikePostgresRepository(db)
	rideRepo := repository.NewRidePostgresRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	tenantRepo := repository.NewTenantPostgresRepository(db)
	configRepo := repository.NewConfigRepository(db)
	txRepo := repository.NewTransactionPostgresRepository(db)
	spendingRepo := repository.NewUserMonthlySpendingRepository(db)
	reportRepo := repository.NewReportPostgresRepository(db)
	couponRepo := repository.NewCouponPostgresRepository(db)

	rankCacheRepo := repository.NewUserRankRedisRepository(rdb)
	configCacheRepo := repository.NewConfigRedisRepository(rdb)

	createService := create.NewService(userRepo)
	verifyFirebaseService := verify_firebase.NewService(userRepo, smsProvider)
	loginService := login.NewService(userRepo)
	googleLoginService := google_login.NewService(userRepo)
	meService := me.NewService(userRepo)
	rankService := rank.NewService(spendingRepo, rankCacheRepo)
	updateProfileService := update_profile.NewService(userRepo)

	paySvc := payment.NewPaySolutionsService()
	topupService := walletTopup.NewService(txRepo, userRepo, paySvc)
	transferService := walletTransfer.NewService(db, userRepo, txRepo, mqttPub, notiProvider)
	verifyService := walletVerify.NewService(userRepo)
	webhookService := walletWebhook.NewService(txRepo, userRepo)

	bikeCreateService := bikeCreate.NewService(bikeRepo)
	bikeListService := bikeList.NewService(bikeRepo)
	bikeUpdateStatusService := bikeUpdateStatus.NewService(bikeRepo)

	rideStartService := rideStart.NewService(rideRepo, userRepo, bikeRepo, tenantRepo, zoneRepo)
	rideEndService := rideEnd.NewService(rideRepo, userRepo, bikeRepo, zoneRepo, tenantRepo, configRepo, spendingRepo, rankCacheRepo, configCacheRepo, couponRepo, mqttPub, notiProvider, db)

	zoneListService := zoneList.NewService(zoneRepo)
	zoneCreateService := zoneCreate.NewService(zoneRepo)

	notiBroadcastService := notiBroadcast.NewService(db, mqttPub, notiProvider)
	notiListService := notiList.NewService(db)
	updateTokenService := notiUpdateToken.NewService(db)

	couponRedeemService := couponRedeem.NewService(db, couponRepo, userRepo, mqttPub)
	couponCreateService := couponCreate.NewService(db)
	couponListMyService := couponListMy.NewService(db)

	reportSummaryService := reportSummary.NewService(reportRepo)
	configService := configUsecase.NewService(configRepo)

	return &Container{
		CreateUserHandler:       create.NewHandler(createService),
		VerifyFirebaseHandler:   verify_firebase.NewHandler(verifyFirebaseService),
		LoginHandler:            login.NewHandler(loginService),
		GoogleLoginHandler:      google_login.NewHandler(googleLoginService),
		MeHandler:               me.NewHandler(meService),
		RankHandler:             rank.NewHandler(rankService),
		UpdateProfileHandler:    update_profile.NewHandler(updateProfileService),
		LogoutHandler:           logout.NewHandler(),
		TopupHandler:            walletTopup.NewHandler(topupService),
		TransferHandler:         walletTransfer.NewHandler(transferService),
		VerifyReceiverHandler:   walletVerify.NewHandler(verifyService),
		WebhookHandler:          walletWebhook.NewHandler(webhookService),
		CreateBikeHandler:       bikeCreate.NewHandler(bikeCreateService),
		ListBikeHandler:         bikeList.NewHandler(bikeListService),
		UpdateBikeStatusHandler: bikeUpdateStatus.NewHandler(bikeUpdateStatusService),
		StartRideHandler:        rideStart.NewHandler(rideStartService),
		EndRideHandler:          rideEnd.NewHandler(rideEndService),
		ListZoneHandler:         zoneList.NewHandler(zoneListService),
		CreateZoneHandler:       zoneCreate.NewHandler(zoneCreateService),
		NotiBroadcastHandler:    notiBroadcast.NewHandler(notiBroadcastService),
		NotiListHandler:         notiList.NewHandler(notiListService),
		NotiUpdateTokenHandler:  notiUpdateToken.NewHandler(updateTokenService),
		ReportSummaryHandler:    reportSummary.NewHandler(reportSummaryService),
		RedeemCouponHandler:     couponRedeem.NewHandler(couponRedeemService),
		CreateCouponHandler:     couponCreate.NewHandler(couponCreateService),
		ListMyCouponHandler:     couponListMy.NewHandler(couponListMyService),
		ConfigHandler:           configUsecase.NewHandler(configService),
	}
}
