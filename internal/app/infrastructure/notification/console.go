package notification

import (
	"context"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type consoleNotificationProvider struct{}

func NewConsoleNotificationProvider() service.NotificationProvider {
	return &consoleNotificationProvider{}
}

func (p *consoleNotificationProvider) SendToToken(ctx context.Context, token string, title string, body string, data map[string]string) error {
	fmt.Printf("[PUSH NOTIFICATION] Send to Token: %s\nTitle: %s\nBody: %s\nData: %v\n", token, title, body, data)
	return nil
}

func (p *consoleNotificationProvider) SendToTopic(ctx context.Context, topic string, title string, body string, data map[string]string) error {
	fmt.Printf("[PUSH BROADCAST] Send to Topic: %s\nTitle: %s\nBody: %s\nData: %v\n", topic, title, body, data)
	return nil
}
