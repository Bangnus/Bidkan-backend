package security

import (
	"context"
	"testing"
	"time"
)

func TestJWTTokenPairAndVerification(t *testing.T) {
	cfg := &Config{
		JWTSecret:       []byte("0123456789abcdef0123456789abcdef0123456789abcdef"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		AppEnv:          "test",
	}

	pair, err := IssueTokenPair(cfg, IssueInput{
		UserID:   "user-123",
		Role:     "super_admin",
		AppScope: ScopeBackoffice,
	})
	if err != nil {
		t.Fatalf("IssueTokenPair failed: %v", err)
	}

	claims, err := ParseAndVerifyToken(cfg, pair.AccessToken, TokenTypeAccess, ScopeBackoffice)
	if err != nil {
		t.Fatalf("ParseAndVerifyToken failed: %v", err)
	}
	if claims.UserID != "user-123" || claims.Role != "super_admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if _, err := ParseAndVerifyToken(cfg, pair.AccessToken, TokenTypeAccess, ScopeRider); err == nil {
		t.Fatal("expected scope mismatch to fail")
	}
	if _, err := ParseAndVerifyToken(cfg, pair.RefreshToken, TokenTypeAccess, ScopeBackoffice); err == nil {
		t.Fatal("expected refresh token used as access token to fail")
	}
	if _, err := ParseAndVerifyToken(cfg, pair.AccessToken+"tampered", TokenTypeAccess, ScopeBackoffice); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestRefreshReuseDetectionAndBlacklist(t *testing.T) {
	ctx := context.Background()
	store := NewMemorySessionStore()

	rec := RefreshRecord{
		UserID:    "u1",
		FamilyID:  "fam-1",
		AppScope:  ScopeRider,
		Role:      "user",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := store.CreateRefreshSession(ctx, "r1", rec, time.Hour); err != nil {
		t.Fatalf("CreateRefreshSession failed: %v", err)
	}
	if _, err := store.RotateRefreshSession(ctx, "r1"); err != nil {
		t.Fatalf("first RotateRefreshSession failed: %v", err)
	}
	rec2 := rec
	if err := store.CreateRefreshSession(ctx, "r2", rec2, time.Hour); err != nil {
		t.Fatalf("CreateRefreshSession r2 failed: %v", err)
	}
	if _, err := store.RotateRefreshSession(ctx, "r1"); err == nil {
		t.Fatal("expected replay of r1 to fail")
	}
	if _, err := store.RotateRefreshSession(ctx, "r2"); err == nil {
		t.Fatal("expected r2 in compromised family to be revoked")
	}

	_ = store.BlacklistAccessToken(ctx, "a1", time.Minute)
	blacklisted, err := store.IsAccessBlacklisted(ctx, "a1")
	if err != nil || !blacklisted {
		t.Fatal("expected access token to be blacklisted")
	}
}

func TestLoadConfigRejectsWeakOrMissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected empty JWT_SECRET to fail")
	}
	t.Setenv("JWT_SECRET", "bidkan_secret_key_2024")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected weak legacy JWT_SECRET to fail")
	}
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if _, err := LoadConfig(); err != nil {
		t.Fatalf("expected valid 32-char secret to pass: %v", err)
	}
}

