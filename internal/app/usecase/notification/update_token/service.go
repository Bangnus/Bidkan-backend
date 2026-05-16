package update_token

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type Request struct {
	FCMToken string `json:"fcm_token" validate:"required"`
}

type Service interface {
	Execute(ctx context.Context, userID uuid.UUID, token string) error
}

type service struct {
	queries *sqlc.Queries
}

func NewService(db *sql.DB) Service {
	return &service{
		queries: sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, userID uuid.UUID, token string) error {
	return s.queries.UpdateUserFCMToken(ctx, sqlc.UpdateUserFCMTokenParams{
		ID:       userID,
		FcmToken: sql.NullString{String: token, Valid: true},
	})
}
