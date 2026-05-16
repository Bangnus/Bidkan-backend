package list

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
	"database/sql"
)

type NotificationInfo struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	Execute(ctx context.Context, userID uuid.UUID) ([]NotificationInfo, error)
}

type service struct {
	queries *sqlc.Queries
}

func NewService(db *sql.DB) Service {
	return &service{
		queries: sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, userID uuid.UUID) ([]NotificationInfo, error) {
	rows, err := s.queries.GetMyNotifications(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	var results []NotificationInfo
	for _, r := range rows {
		results = append(results, NotificationInfo{
			ID:        r.ID,
			Title:     r.Title,
			Message:   r.Message,
			Type:      r.Type,
			IsRead:    r.IsRead,
			CreatedAt: r.CreatedAt,
		})
	}
	return results, nil
}
