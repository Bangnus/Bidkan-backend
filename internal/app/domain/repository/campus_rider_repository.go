package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type CampusRiderRepository interface {
	ListCampusRiders(ctx context.Context, tenantID uuid.UUID, search, filter string) ([]*entity.CampusRider, error)
	UpdateStudentVerification(ctx context.Context, userID uuid.UUID, verified bool) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	ListDisputedRides(ctx context.Context, tenantID uuid.UUID, status string) ([]*entity.DisputedRide, error)
	RefundPenalty(ctx context.Context, tenantID, rideID, userID uuid.UUID, refundAmount float64, reason string) error
}

