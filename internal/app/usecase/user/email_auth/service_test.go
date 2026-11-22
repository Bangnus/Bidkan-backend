package email_auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/google/uuid"
)

type memUserRepo struct{ byEmail map[string]*entity.User }

func (m *memUserRepo) Create(context.Context, *entity.User) error                  { return nil }
func (m *memUserRepo) GetByPhone(context.Context, string) (*entity.User, error)    { return nil, nil }
func (m *memUserRepo) GetByID(context.Context, uuid.UUID) (*entity.User, error)    { return nil, nil }
func (m *memUserRepo) GetByUsername(context.Context, string) (*entity.User, error) { return nil, nil }
func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*entity.User, error) {
	u := m.byEmail[strings.ToLower(email)]
	if u == nil {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (m *memUserRepo) GetByIdentifier(ctx context.Context, id string) (*entity.User, error) {
	return m.GetByEmail(ctx, id)
}
func (m *memUserRepo) CreateEmailUser(_ context.Context, user *entity.User) (*entity.User, error) {
	u := *user
	u.Role, u.Status, u.WalletBalance = "user", "pending", "0.00"
	m.byEmail[strings.ToLower(u.Email)] = &u
	return &u, nil
}
func (m *memUserRepo) RefreshPendingEmailUser(_ context.Context, user *entity.User) (*entity.User, error) {
	m.byEmail[strings.ToLower(user.Email)] = user
	return user, nil
}
func (m *memUserRepo) ActivateVerifiedEmailUser(_ context.Context, email string) (*entity.User, error) {
	u := m.byEmail[strings.ToLower(email)]
	u.Status, u.EmailVerified = "active", true
	return u, nil
}
func (m *memUserRepo) UpdateStatus(context.Context, uuid.UUID, string) error         { return nil }
func (m *memUserRepo) UpdateBalance(context.Context, uuid.UUID, string) error        { return nil }
func (m *memUserRepo) AddBalance(context.Context, uuid.UUID, string) error           { return nil }
func (m *memUserRepo) DeductBalance(context.Context, uuid.UUID, string) error        { return nil }
func (m *memUserRepo) UpdateProfileImage(context.Context, uuid.UUID, string) error   { return nil }
func (m *memUserRepo) SubmitStudentKYC(context.Context, uuid.UUID, string, string) error {
	return nil
}
func (m *memUserRepo) UpsertGoogleUser(context.Context, string, string, string, string, *uuid.UUID) (*entity.User, error) {
	return nil, nil
}

type memVerifyRepo struct {
	codes    map[string]string
	attempts map[string]int
	cooldown map[string]time.Time
}

func (m *memVerifyRepo) SaveCode(_ context.Context, email, code string, _ time.Duration) error {
	m.codes[strings.ToLower(email)] = code
	m.attempts[strings.ToLower(email)] = 0
	return nil
}
func (m *memVerifyRepo) VerifyCode(_ context.Context, email, code string, maxAttempts int) (bool, error) {
	key := strings.ToLower(email)
	if m.codes[key] == "" {
		return false, nil
	}
	if m.codes[key] == code {
		delete(m.codes, key)
		return true, nil
	}
	m.attempts[key]++
	if m.attempts[key] >= maxAttempts {
		delete(m.codes, key)
		return false, errors.New("too many attempts")
	}
	return false, nil
}
func (m *memVerifyRepo) ResendCooldownRemaining(_ context.Context, email string) (time.Duration, error) {
	until, ok := m.cooldown[strings.ToLower(email)]
	if !ok || time.Now().After(until) {
		return 0, nil
	}
	return time.Until(until), nil
}
func (m *memVerifyRepo) SetResendCooldown(_ context.Context, email string, ttl time.Duration) error {
	m.cooldown[strings.ToLower(email)] = time.Now().Add(ttl)
	return nil
}

type spyEmailProvider struct{ sent []domainService.EmailMessage }

func (s *spyEmailProvider) Send(_ context.Context, msg domainService.EmailMessage) error {
	s.sent = append(s.sent, msg)
	return nil
}
func (s *spyEmailProvider) Mode() string { return "console" }

func TestRegisterVerifyAndCooldownFlow(t *testing.T) {
	users := &memUserRepo{byEmail: map[string]*entity.User{}}
	verifyRepo := &memVerifyRepo{
		codes:    map[string]string{},
		attempts: map[string]int{},
		cooldown: map[string]time.Time{},
	}
	mailer := &spyEmailProvider{}
	secCfg := &security.Config{
		JWTSecret:       []byte("0123456789abcdef0123456789abcdef"),
		AccessTokenTTL:  30 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		AppEnv:          "development",
	}
	svc := NewService(users, nil, verifyRepo, mailer, secCfg, security.NewMemorySessionStore())

	regRes, err := svc.Register(context.Background(), RegisterRequest{
		Email: "Rider@ku.th", Username: "Somchai", Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if regRes.DevOTP == "" || len(mailer.sent) != 1 {
		t.Fatalf("expected dev OTP and 1 email sent, got otp=%q sent=%d", regRes.DevOTP, len(mailer.sent))
	}

	// Wrong code must fail
	if _, err := svc.VerifyEmail(context.Background(), VerifyEmailRequest{
		Email: "rider@ku.th", Code: "000000",
	}); err == nil && regRes.DevOTP != "000000" {
		t.Fatal("expected wrong OTP to fail")
	}

	// Immediate resend must hit cooldown
	if _, err := svc.ResendVerification(context.Background(), ResendRequest{Email: "rider@ku.th"}); !errors.Is(err, ErrResendTooSoon) {
		t.Fatalf("expected ErrResendTooSoon, got %v", err)
	}

	// Valid OTP activates user & issues tokens
	verifyRes, err := svc.VerifyEmail(context.Background(), VerifyEmailRequest{
		Email: "rider@ku.th", Code: regRes.DevOTP,
	})
	if err != nil {
		t.Fatalf("verify email failed: %v", err)
	}
	if verifyRes.AccessToken == "" || verifyRes.User == nil || !verifyRes.User.EmailVerified || verifyRes.User.Status != "active" {
		t.Fatalf("unexpected verify response: %+v", verifyRes)
	}

	// Duplicate active registration rejected
	if _, err := svc.Register(context.Background(), RegisterRequest{
		Email: "rider@ku.th", Username: "Somchai2", Password: "password123",
	}); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}
