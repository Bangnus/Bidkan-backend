package app

import (
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	
	// User
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/login"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/logout"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/me"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/update_profile"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/verify_firebase"
	
	// Bike
	bikeCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	bikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	
	// Ride
	rideStart "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/start"
	rideEnd "github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/end"

	// Zone
	zoneList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/list"
	zoneCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/zone/create"

	// Config
	configUsecase "github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
)

type Container struct {
	// User
	CreateUserHandler     create.Handler
	VerifyFirebaseHandler verify_firebase.Handler
	LoginHandler          login.Handler
	MeHandler             me.Handler
	UpdateProfileHandler  update_profile.Handler
	LogoutHandler         logout.Handler

	// Bike
	CreateBikeHandler     bikeCreate.Handler
	ListBikeHandler       bikeList.Handler

	// Ride
	StartRideHandler      rideStart.Handler
	EndRideHandler        rideEnd.Handler

	// Zone
	ListZoneHandler       zoneList.Handler
	CreateZoneHandler     zoneCreate.Handler

	// Config
	ConfigHandler         configUsecase.Handler
}

func NewContainer(db *sql.DB, firebaseProvider service.SmsProvider) *Container {
	// --- Repositories ---
	userRepo := repository.NewUserPostgresRepository(db)
	bikeRepo := repository.NewBikePostgresRepository(db)
	rideRepo := repository.NewRidePostgresRepository(db)
	zoneRepo := repository.NewZoneRepository(db) // ตรวจสอบชื่อฟังก์ชันใน repository/zone_postgres.go
	configRepo := repository.NewConfigRepository(db)

	// --- Services ---
	// User
	createService := create.NewService(userRepo)
	verifyFirebaseService := verify_firebase.NewService(userRepo, firebaseProvider)
	loginService := login.NewService(userRepo)
	meService := me.NewService(userRepo)
	updateProfileService := update_profile.NewService(userRepo)

	// Bike
	bikeCreateService := bikeCreate.NewService(bikeRepo)
	bikeListService := bikeList.NewService(bikeRepo)

	// Ride
	rideStartService := rideStart.NewService(rideRepo, userRepo, bikeRepo)
	rideEndService := rideEnd.NewService(rideRepo, userRepo, bikeRepo, zoneRepo, configRepo)

	// Zone
	zoneListService := zoneList.NewService(zoneRepo)
	zoneCreateService := zoneCreate.NewService(zoneRepo)

	// Config
	configService := configUsecase.NewService(configRepo)

	// --- Handlers ---
	return &Container{
		CreateUserHandler:     create.NewHandler(createService),
		VerifyFirebaseHandler: verify_firebase.NewHandler(verifyFirebaseService),
		LoginHandler:          login.NewHandler(loginService),
		MeHandler:             me.NewHandler(meService),
		UpdateProfileHandler:  update_profile.NewHandler(updateProfileService),
		LogoutHandler:         logout.NewHandler(),

		CreateBikeHandler:     bikeCreate.NewHandler(bikeCreateService),

		ListBikeHandler:       bikeList.NewHandler(bikeListService),

		StartRideHandler:      rideStart.NewHandler(rideStartService),
		EndRideHandler:        rideEnd.NewHandler(rideEndService),

		ListZoneHandler:       zoneList.NewHandler(zoneListService),
		CreateZoneHandler:     zoneCreate.NewHandler(zoneCreateService),

		ConfigHandler:         configUsecase.NewHandler(configService),
	}
}
