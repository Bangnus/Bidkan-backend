package database

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	var opts *redis.Options
	var err error

	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		opts, err = redis.ParseURL(addr)
		if err != nil {
			log.Fatalf("❌ Invalid REDIS_URL format: %v", err)
		}
	} else {
		opts = &redis.Options{
			Addr: addr,
		}
	}

	rdb := redis.NewClient(opts)

	// ทดสอบการเชื่อมต่อ (Ping) ด้วย Context Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Warning: Failed to connect to Redis at %s: %v", addr, err)
	} else {
		log.Printf("✅ Connected to Redis successfully (%s)", addr)
	}

	return rdb
}
