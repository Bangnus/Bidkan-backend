package repository

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/redis/go-redis/v9"
)

type otpRedisRepository struct {
	client *redis.Client
}

func NewOtpRedisRepository(client *redis.Client) repository.OtpRepository {
	return &otpRedisRepository{client: client}
}

func (r *otpRedisRepository) SaveOTP(ctx context.Context, phone string, otp string, expiration time.Duration) error {
	// ใช้ OTP เป็น Key เพื่อให้ค้นหาเบอร์โทรกลับมาได้
	key := "otp:" + otp
	return r.client.Set(ctx, key, phone, expiration).Err()
}

func (r *otpRedisRepository) VerifyOTP(ctx context.Context, otp string) (string, error) {
	key := "otp:" + otp
	phone, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // ไม่พบรหัสนี้
	}
	if err != nil {
		return "", err
	}
	return phone, nil
}

func (r *otpRedisRepository) DeleteOTP(ctx context.Context, otp string) error {
	key := "otp:" + otp
	return r.client.Del(ctx, key).Err()
}
