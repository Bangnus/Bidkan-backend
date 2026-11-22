package repository

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/redis/go-redis/v9"
)

type emailVerifyRecord struct {
	CodeHash string `json:"code_hash"`
	Attempts int    `json:"attempts"`
}

type emailVerificationRedisRepository struct {
	rdb *redis.Client
}

// NewEmailVerificationRedisRepository stores hashed email verification OTPs in Redis.
func NewEmailVerificationRedisRepository(rdb *redis.Client) domainRepo.EmailVerificationRepository {
	return &emailVerificationRedisRepository{rdb: rdb}
}

func (r *emailVerificationRedisRepository) SaveCode(ctx context.Context, email, code string, ttl time.Duration) error {
	if r.rdb == nil {
		return errors.New("verification store unavailable")
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	rec := emailVerifyRecord{CodeHash: hashEmailCode(norm, code), Attempts: 0}
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, emailVerifyKey(norm), raw, ttl).Err()
}

func (r *emailVerificationRedisRepository) VerifyCode(ctx context.Context, email, code string, maxAttempts int) (bool, error) {
	if r.rdb == nil {
		return false, errors.New("verification store unavailable")
	}
	norm := strings.ToLower(strings.TrimSpace(email))
	key := emailVerifyKey(norm)
	raw, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var rec emailVerifyRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		_ = r.rdb.Del(ctx, key).Err()
		return false, nil
	}

	expected := hashEmailCode(norm, strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(rec.CodeHash), []byte(expected)) == 1 {
		_ = r.rdb.Del(ctx, key).Err()
		return true, nil
	}

	rec.Attempts++
	if maxAttempts > 0 && rec.Attempts >= maxAttempts {
		_ = r.rdb.Del(ctx, key).Err()
		return false, errors.New("too many failed verification attempts, please request a new code")
	}

	ttl, _ := r.rdb.TTL(ctx, key).Result()
	if ttl < time.Second {
		ttl = 5 * time.Minute
	}
	updated, _ := json.Marshal(rec)
	_ = r.rdb.Set(ctx, key, updated, ttl).Err()
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
func emailCooldownKey(email string) string { return "auth:email_verify_cd:" + email }

func hashEmailCode(email, code string) string {
	sum := sha256.Sum256([]byte(email + ":" + code))
	return hex.EncodeToString(sum[:])
}
