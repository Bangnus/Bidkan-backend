package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/redis/go-redis/v9"
)

type userRankRedisRepository struct {
	rdb *redis.Client
}

func NewUserRankRedisRepository(rdb *redis.Client) repository.UserRankCacheRepository {
	return &userRankRedisRepository{rdb: rdb}
}

func (r *userRankRedisRepository) SetRankCache(ctx context.Context, userID string, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("user:rank:%s", userID)
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, payload, ttl).Err()
}

func (r *userRankRedisRepository) GetRankCache(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user:rank:%s", userID)
	return r.rdb.Get(ctx, key).Result()
}

func (r *userRankRedisRepository) DeleteRankCache(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:rank:%s", userID)
	return r.rdb.Del(ctx, key).Err()
}
