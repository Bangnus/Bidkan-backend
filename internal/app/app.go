package app

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
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
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/refresh"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/stripe_webhook"
	walletStripeTopup "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/stripe_topup"
	walletTopup "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	walletTransfer "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	walletVerify "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"
	walletWebhook "github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"
	zoneCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/create"
	zoneList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/list"
)

type Container struct {
	UserRepo   domainRepo.UserRepository
	TenantRepo domainRepo.TenantRepository

	// User
	CreateUserHandler     create.Handler
	VerifyFirebaseHandler verify_firebase.Handler
	LoginHandler          login.Handler
	GoogleLoginHandler    google_login.Handler
	RefreshUserHandler    refresh.Handler
	MeHandler             me.Handler
	RankHandler           rank.Handler
	UpdateProfileHandler  update_profile.Handler
	LogoutHandler         logout.Handler

	// Wallet
	TopupHandler         walletTopup.Handler
	StripeTopupHandler   walletStripeTopup.Handler
	TransferHandler      walletTransfer.Handler
	VerifyReceiverHandler walletVerify.Handler
	WebhookHandler       walletWebhook.Handler
	StripeWebhookHandler stripe_webhook.Handler

	// Bike
	CreateBikeHandler       bikeCreate.Handler
	ListBikeHandler         bikeList.Handler
	UpdateBikeStatusHandler bikeUpdateStatus.Handler

	// Ride
	StartRideHandler rideStart.Handler
	EndRideHandler   rideEnd.Handler

	// Zone
	ListZoneHandler   zoneList.Handler
	CreateZoneHandler zoneCreate.Handler

	// Notification
	NotiBroadcastHandler   notiBroadcast.Handler
	NotiListHandler        notiList.Handler
	NotiUpdateTokenHandler notiUpdateToken.Handler

	// Coupon
	RedeemCouponHandler couponRedeem.Handler
	CreateCouponHandler couponCreate.Handler
	ListMyCouponHandler couponListMy.Handler

	// Report
	ReportSummaryHandler reportSummary.Handler

	// Config
	ConfigHandler configUsecase.Handler

	// External
	SmsProvider   domainService.SmsProvider
	NotiProvider  domainService.NotificationProvider
	MqttPublisher mqtt.Publisher
	StripeClient  infraStripe.Client
}
