package repository

import (
	"context"
	"time"
)

type UserRankCacheRepository interface {
	SetRankCache(ctx context.Context, userID string, data interface{}, ttl time.Duration) error
	GetRankCache(ctx context.Context, userID string) (string, error)
	DeleteRankCache(ctx context.Context, userID string) error
}
