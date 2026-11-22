package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/redis/go-redis/v9"
)

type emailVerificationRedisRepository struct {
	rdb    *redis.Client
	pepper []byte
}

// NewEmailVerificationRedisRepository stores HMAC-SHA256 hashed OTPs with atomic attempt tracking in Redis.
func NewEmailVerificationRedisRepository(rdb *redis.Client, pepper []byte) domainRepo.EmailVerificationRepository {
	return &emailVerificationRedisRepository{rdb: rdb, pepper: pepper}
}

func (r *emailVerificationRedisRepository) SaveCode(ctx context.Context, email, code string, ttl time.Duration) error {
	if r.rdb == nil {
		return errors.New("verification store unavailable")
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, emailVerifyKey(norm), r.hashEmailCode(norm, code), ttl)
	pipe.Del(ctx, emailTriesKey(norm))
	_, err := pipe.Exec(ctx)
	return err
}

func (r *emailVerificationRedisRepository) VerifyCode(ctx context.Context, email, code string, maxAttempts int) (bool, error) {
	if r.rdb == nil {
		return false, errors.New("verification store unavailable")
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	key := emailVerifyKey(norm)
	triesKey := emailTriesKey(norm)

	storedHash, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Atomic INCR prevents concurrent brute-force requests from racing the attempt counter.
	tries, err := r.rdb.Incr(ctx, triesKey).Result()
	if err != nil {
		return false, err
	}
	if tries == 1 {
		if ttl, err := r.rdb.TTL(ctx, key).Result(); err == nil && ttl > 0 {
			_ = r.rdb.Expire(ctx, triesKey, ttl).Err()
		} else {
			_ = r.rdb.Expire(ctx, triesKey, 15*time.Minute).Err()
		}
	}
	if maxAttempts > 0 && int(tries) > maxAttempts {
		_ = r.rdb.Del(ctx, key, triesKey).Err()
		return false, errors.New("too many failed verification attempts, please request a new code")
	}

	expected := r.hashEmailCode(norm, strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(expected)) == 1 {
		_ = r.rdb.Del(ctx, key, triesKey).Err()
		return true, nil
	}

	if maxAttempts > 0 && int(tries) >= maxAttempts {
		_ = r.rdb.Del(ctx, key, triesKey).Err()
		return false, errors.New("too many failed verification attempts, please request a new code")
	}
	return false, nil
}

func (r *emailVerificationRedisRepository) ResendCooldownRemaining(ctx context.Context, email string) (time.Duration, error) {
	if r.rdb == nil {
		return 0, nil
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	ttl, err := r.rdb.TTL(ctx, emailCooldownKey(norm)).Result()
	if err != nil || ttl <= 0 {
		return 0, nil
	}
	return ttl, nil
}

func (r *emailVerificationRedisRepository) SetResendCooldown(ctx context.Context, email string, ttl time.Duration) error {
	if r.rdb == nil || ttl <= 0 {
		return nil
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	return r.rdb.Set(ctx, emailCooldownKey(norm), "1", ttl).Err()
}

func emailVerifyKey(email string) string   { return "auth:email_verify:" + email }
func emailTriesKey(email string) string    { return "auth:email_verify_tries:" + email }
func emailCooldownKey(email string) string { return "auth:email_verify_cd:" + email }

func (r *emailVerificationRedisRepository) hashEmailCode(email, code string) string {
	mac := hmac.New(sha256.New, r.pepper)
	_, _ = mac.Write([]byte(email + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}
