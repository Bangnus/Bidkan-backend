package repository

import (
	"database/sql"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func mapTenantModelToEntity(row sqlc.Tenant) *entity.Tenant {
	gp, _ := strconv.ParseFloat(row.PlatformFeePercent, 64)
	unlock, _ := strconv.ParseFloat(row.UnlockFee, 64)
	perMin, _ := strconv.ParseFloat(row.PerMinuteRate, 64)
	penalty, _ := strconv.ParseFloat(row.ParkingPenaltyFee, 64)
	if penalty <= 0 {
		penalty = 50.0
	}
	simFee, _ := strconv.ParseFloat(row.SimFeePerBike, 64)

	var stripeID *string
	if row.StripeAccountID.Valid {
		stripeID = &row.StripeAccountID.String
	}
	var orgID *uuid.UUID
	if row.OrganizationID.Valid {
		id := row.OrganizationID.UUID
		orgID = &id
	}

	return &entity.Tenant{
		ID:                        row.ID,
		OrganizationID:            orgID,
		Slug:                      row.Slug,
		Name:                      row.Name,
		StripeAccountID:           stripeID,
		StripeOnboardingCompleted: row.StripeOnboardingCompleted,
		PlatformFeePercent:        gp,
		AutoTieredGP:              row.AutoTieredGp,
		UnlockFee:                 unlock,
		PerMinuteRate:             perMin,
		ParkingPenaltyFee:         penalty,
		SimFeePerBike:             simFee,
		ContactName:               row.ContactName.String,
		ContactEmail:              row.ContactEmail.String,
		ContactPhone:              row.ContactPhone.String,
		Notes:                     row.Notes.String,
		Status:                    row.Status,
		CreatedAt:                 row.CreatedAt,
		UpdatedAt:                 row.UpdatedAt,
	}
}

func mapBikeModelToEntity(b sqlc.Bike) *entity.BikeData {
	var hb sql.NullTime
	if b.LastHeartbeat.Valid {
		hb = b.LastHeartbeat
	}
	var rideID *string
	if b.CurrentRideID.Valid {
		str := b.CurrentRideID.UUID.String()
		rideID = &str
	}
	var tenantID *uuid.UUID
	if b.TenantID.Valid {
		tid := b.TenantID.UUID
		tenantID = &tid
	}
	return &entity.BikeData{
		BikeID:        b.ID,
		HardwareID:    b.HardwareID,
		TenantID:      tenantID,
		Lat:           b.Lat,
		Lon:           b.Lon,
		Battery:       int(b.BatteryLevel),
		Status:        b.Status,
		ImageURL:      b.ImageUrl.String,
		CurrentRideID: rideID,
		LastHeartbeat: hb.Time,
	}
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}
