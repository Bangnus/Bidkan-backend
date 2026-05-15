package repository

import (
	"context"
)

type ConfigRepository interface {
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key string, value string) error
}
