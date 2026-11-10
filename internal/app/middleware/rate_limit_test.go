package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type stubUserRepo struct {
	users map[uuid.UUID]*entity.User
}

func (s *stubUserRepo) Create(context.Context, *entity.User) error { return nil }
func (s *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return u, nil
}
func (s *stubUserRepo) GetByPhone(context.Context, string) (*entity.User, error) { return nil, nil }
func (s *stubUserRepo) GetByUsername(context.Context, string) (*entity.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetByIdentifier(context.Context, string) (*entity.User, error) {
	return nil, nil
}
func (s *stubUserRepo) UpdateStatus(context.Context, uuid.UUID, string) error { return nil }
func (s *stubUserRepo) UpdateBalance(context.Context, uuid.UUID, string) error {
	return nil
}
func (s *stubUserRepo) AddBalance(context.Context, uuid.UUID, string) error    { return nil }
func (s *stubUserRepo) DeductBalance(context.Context, uuid.UUID, string) error { return nil }
func (s *stubUserRepo) UpdateProfileImage(context.Context, uuid.UUID, string) error {
	return nil
}
func (s *stubUserRepo) UpsertGoogleUser(context.Context, string, string, string, string, *uuid.UUID) (*entity.User, error) {
	return nil, nil
}

func TestRateLimiterAndBlacklistAndUserStatus(t *testing.T) {
	cfg := &security.Config{
		JWTSecret:       []byte("0123456789abcdef0123456789abcdef"),
		AccessTokenTTL:  30 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	store := security.NewMemorySessionStore()
	activeID := uuid.New()
	suspendedID := uuid.New()
	repo := &stubUserRepo{
		users: map[uuid.UUID]*entity.User{
			activeID:    {ID: activeID, Role: "super_admin", Status: "active"},
			suspendedID: {ID: suspendedID, Role: "super_admin", Status: "suspended"},
		},
	}

	app := fiber.New()
	deps := AuthDeps{Config: cfg, Sessions: store, UserRepo: repo}
	app.Post("/login", AuthRateLimiter(store, "test_rl", 2, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/protected", AuthMiddleware(deps), RequireScope(security.ScopeBackoffice), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Rate limit 2 allowed, 3rd returns 429
	for i := 0; i < 3; i++ {
		resp, err := app.Test(httptest.NewRequest("POST", "/login", nil))
		if err != nil {
			t.Fatalf("rate limit request err: %v", err)
		}
		want := fiber.StatusOK
		if i == 2 {
			want = fiber.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Fatalf("attempt %d: got %d, want %d", i+1, resp.StatusCode, want)
		}
	}

	// Valid backoffice token -> 200, then blacklisted -> 401
	boPair, _ := security.IssueTokenPair(cfg, security.IssueInput{
		UserID: activeID.String(), Role: "super_admin", AppScope: security.ScopeBackoffice,
	})
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+boPair.AccessToken)
	if resp, _ := app.Test(req); resp.StatusCode != fiber.StatusOK {
		t.Fatalf("valid token got %d, want 200", resp.StatusCode)
	}
	_ = store.BlacklistAccessToken(context.Background(), boPair.AccessJTI, time.Minute)
	if resp, _ := app.Test(req); resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("blacklisted token got %d, want 401", resp.StatusCode)
	}

	// Rider token on backoffice route -> 403
	riderPair, _ := security.IssueTokenPair(cfg, security.IssueInput{
		UserID: activeID.String(), Role: "super_admin", AppScope: security.ScopeRider,
	})
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+riderPair.AccessToken)
	if resp, _ := app.Test(req2); resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("wrong scope got %d, want 403", resp.StatusCode)
	}

	// Suspended user -> 403
	suspPair, _ := security.IssueTokenPair(cfg, security.IssueInput{
		UserID: suspendedID.String(), Role: "super_admin", AppScope: security.ScopeBackoffice,
	})
	req3 := httptest.NewRequest("GET", "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+suspPair.AccessToken)
	if resp, _ := app.Test(req3); resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("suspended user got %d, want 403", resp.StatusCode)
	}
}
