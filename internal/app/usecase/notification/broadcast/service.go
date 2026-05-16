package broadcast

import (
	"context"
	"database/sql"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type Request struct {
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
	Type    string `json:"type" validate:"required"` // promo, news, system
}

type Service interface {
	Execute(ctx context.Context, req Request) error
}

type service struct {
	queries *sqlc.Queries
	mqttPub mqtt.Publisher
	notiPub domainService.NotificationProvider
}

func NewService(db *sql.DB, mqttPub mqtt.Publisher, notiPub domainService.NotificationProvider) Service {
	return &service{
		queries: sqlc.New(db),
		mqttPub:  mqttPub,
		notiPub:  notiPub,
	}
}

func (s *service) Execute(ctx context.Context, req Request) error {
	id := uuid.New()
	
	// 1. บันทึกลงฐานข้อมูล (user_id เป็น NULL หมายถึงส่งหาทุกคน)
	err := s.queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		ID:      id,
		UserID:  uuid.NullUUID{Valid: false}, // Broadcast
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
	})
	if err != nil {
		return err
	}

	// 2. ส่งผ่าน MQTT (Topic กลางสำหรับ Broadcast)
	if s.mqttPub != nil {
		s.mqttPub.Publish("bidkan/broadcast", map[string]interface{}{
			"id":      id,
			"title":   req.Title,
			"message": req.Message,
			"type":    req.Type,
		})
	}

	// 3. ส่งผ่าน Push Notification (FCM) ไปยัง Topic 'broadcast'
	if s.notiPub != nil {
		_ = s.notiPub.SendToTopic(ctx, "broadcast", req.Title, req.Message, map[string]string{
			"type": req.Type,
			"id":   id.String(),
		})
	}

	return nil
}
