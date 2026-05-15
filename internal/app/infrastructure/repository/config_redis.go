package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/redis/go-redis/v9"
)

type configRedisRepository struct {
	rdb *redis.Client
}

func NewConfigRedisRepository(rdb *redis.Client) repository.ConfigCacheRepository {
	return &configRedisRepository{rdb: rdb}
}

func (r *configRedisRepository) SetConfigCache(ctx context.Context, key string, value string, ttl time.Duration) error {
	redisKey := fmt.Sprintf("config:%s", key)
	return r.rdb.Set(ctx, redisKey, value, ttl).Err()
}

func (r *configRedisRepository) GetConfigCache(ctx context.Context, key string) (string, error) {
	redisKey := fmt.Sprintf("config:%s", key)
	return r.rdb.Get(ctx, redisKey).Result()
}

func (r *configRedisRepository) DeleteConfigCache(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("config:%s", key)
	return r.rdb.Del(ctx, redisKey).Err()
}
