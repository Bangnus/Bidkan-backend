package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshRecord struct {
	UserID         string    `json:"user_id"`
	FamilyID       string    `json:"family_id"`
	AppScope       string    `json:"app_scope"`
	Role           string    `json:"role"`
	TenantID       string    `json:"tenant_id,omitempty"`
	TenantSlug     string    `json:"tenant_slug,omitempty"`
	OrganizationID string    `json:"organization_id,omitempty"`
	OrgSlug        string    `json:"org_slug,omitempty"`
	AllowedSlugs   []string  `json:"allowed_slugs,omitempty"`
	Used           bool      `json:"used"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type SessionStore interface {
	CreateRefreshSession(ctx context.Context, jti string, rec RefreshRecord, ttl time.Duration) error
	RotateRefreshSession(ctx context.Context, jti string) (*RefreshRecord, error)
	RevokeFamily(ctx context.Context, familyID string) error
	BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessBlacklisted(ctx context.Context, jti string) (bool, error)
	AllowRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error)
	IsLoginLocked(ctx context.Context, key string) (bool, time.Duration, error)
	RecordLoginFailure(ctx context.Context, key string, maxAttempts int, window, lockDuration time.Duration) (bool, error)
	ClearLoginFailures(ctx context.Context, key string) error
}

type redisSessionStore struct {
	rdb *redis.Client
}

func NewRedisSessionStore(rdb *redis.Client) SessionStore {
	return &redisSessionStore{rdb: rdb}
}

func (s *redisSessionStore) CreateRefreshSession(ctx context.Context, jti string, rec RefreshRecord, ttl time.Duration) error {
	if s.rdb == nil {
		return errors.New("redis unavailable")
	}
	rec.Used = false
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, familyKey(rec.FamilyID), "active", ttl)
	pipe.Set(ctx, refreshKey(jti), raw, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *redisSessionStore) RotateRefreshSession(ctx context.Context, jti string) (*RefreshRecord, error) {
	if s.rdb == nil {
		return nil, errors.New("redis unavailable")
	}
	raw, err := s.rdb.Get(ctx, refreshKey(jti)).Result()
	if err != nil {
		return nil, errors.New("refresh session not found or expired")
	}
	var rec RefreshRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, errors.New("corrupt refresh session")
	}
	exists, err := s.rdb.Exists(ctx, familyKey(rec.FamilyID)).Result()
	if err != nil || exists == 0 {
		return nil, errors.New("session family revoked")
	}
	if rec.Used {
		_ = s.RevokeFamily(ctx, rec.FamilyID)
		return nil, errors.New("refresh token reuse detected: session revoked")
	}
	rec.Used = true
	updated, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}
	remaining := time.Until(rec.ExpiresAt)
	if remaining < time.Second {
		remaining = time.Minute
	}
	if err := s.rdb.Set(ctx, refreshKey(jti), updated, remaining).Err(); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *redisSessionStore) RevokeFamily(ctx context.Context, familyID string) error {
	if s.rdb == nil || familyID == "" {
		return nil
	}
	return s.rdb.Del(ctx, familyKey(familyID)).Err()
}

func (s *redisSessionStore) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	if s.rdb == nil || jti == "" || ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

func (s *redisSessionStore) IsAccessBlacklisted(ctx context.Context, jti string) (bool, error) {
	if s.rdb == nil || jti == "" {
		return false, nil
	}
	n, err := s.rdb.Exists(ctx, blacklistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *redisSessionStore) AllowRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	if s.rdb == nil {
		return true, 0, nil
	}
	fullKey := fmt.Sprintf("auth:rl:%s", key)
	count, err := s.rdb.Incr(ctx, fullKey).Result()
	if err != nil {
		return true, 0, err
	}
	if count == 1 {
		_ = s.rdb.Expire(ctx, fullKey, window).Err()
	}
	if int(count) > limit {
		ttl, _ := s.rdb.TTL(ctx, fullKey).Result()
		if ttl < 0 {
			ttl = window
		}
		return false, ttl, nil
	}
	return true, 0, nil
}

func (s *redisSessionStore) IsLoginLocked(ctx context.Context, key string) (bool, time.Duration, error) {
	if s.rdb == nil {
		return false, 0, nil
	}
	lKey := lockKey(key)
	exists, err := s.rdb.Exists(ctx, lKey).Result()
	if err != nil || exists == 0 {
		return false, 0, err
	}
	ttl, _ := s.rdb.TTL(ctx, lKey).Result()
	if ttl < 0 {
		ttl = time.Minute
	}
	return true, ttl, nil
}

func (s *redisSessionStore) RecordLoginFailure(ctx context.Context, key string, maxAttempts int, window, lockDuration time.Duration) (bool, error) {
	if s.rdb == nil {
		return false, nil
	}
	fKey := failKey(key)
	cnt, err := s.rdb.Incr(ctx, fKey).Result()
	if err != nil {
		return false, err
	}
	if cnt == 1 {
		_ = s.rdb.Expire(ctx, fKey, window).Err()
	}
	if int(cnt) >= maxAttempts {
		_ = s.rdb.Set(ctx, lockKey(key), "1", lockDuration).Err()
		_ = s.rdb.Del(ctx, fKey).Err()
		return true, nil
	}
	return false, nil
}

func (s *redisSessionStore) ClearLoginFailures(ctx context.Context, key string) error {
	if s.rdb == nil {
		return nil
	}
	return s.rdb.Del(ctx, failKey(key), lockKey(key)).Err()
}

func refreshKey(jti string) string   { return "auth:refresh:" + jti }
func familyKey(id string) string     { return "auth:family:" + id }
func blacklistKey(jti string) string { return "auth:blacklist:" + jti }
func failKey(key string) string      { return "auth:fail:" + key }
func lockKey(key string) string      { return "auth:lock:" + key }

