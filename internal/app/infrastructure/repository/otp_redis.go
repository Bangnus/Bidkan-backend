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
	key := "otp:" + phone
	return r.client.Set(ctx, key, otp, expiration).Err()
}

func (r *otpRedisRepository) VerifyOTP(ctx context.Context, phone string, otp string) (bool, error) {
	key := "otp:" + phone
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == otp, nil
}

func (r *otpRedisRepository) DeleteOTP(ctx context.Context, phone string) error {
	key := "otp:" + phone
	return r.client.Del(ctx, key).Err()
}
