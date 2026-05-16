package notification

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type firebaseNotificationProvider struct {
	client *messaging.Client
}

func NewFirebaseNotificationProvider(app *firebase.App) (service.NotificationProvider, error) {
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting firebase messaging client: %v", err)
	}
	return &firebaseNotificationProvider{client: client}, nil
}

func (p *firebaseNotificationProvider) SendToToken(ctx context.Context, token string, title string, body string, data map[string]string) error {
	if token == "" {
		return nil // ไม่มี Token ไม่ต้องส่ง
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := p.client.Send(ctx, message)
	return err
}

func (p *firebaseNotificationProvider) SendToTopic(ctx context.Context, topic string, title string, body string, data map[string]string) error {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := p.client.Send(ctx, message)
	return err
}
