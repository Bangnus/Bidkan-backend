package service

import "context"

type NotificationProvider interface {
	// SendToToken ส่งหาเฉพาะรายบุคคล (ใช้ fcm_token)
	SendToToken(ctx context.Context, token string, title string, body string, data map[string]string) error
	
	// SendToTopic ส่งหาทุกคนที่ Subscribe Topic นี้ (เช่น 'broadcast' สำหรับ Ads/PR)
	SendToTopic(ctx context.Context, topic string, title string, body string, data map[string]string) error
}
