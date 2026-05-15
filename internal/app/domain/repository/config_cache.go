package repository

import (
	"context"
	"time"
)

type ConfigCacheRepository interface {
	SetConfigCache(ctx context.Context, key string, value string, ttl time.Duration) error
	GetConfigCache(ctx context.Context, key string) (string, error)
	DeleteConfigCache(ctx context.Context, key string) error
}
