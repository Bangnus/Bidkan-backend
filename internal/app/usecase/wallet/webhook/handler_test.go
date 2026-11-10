package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type stubWebhookService struct {
	called bool
}

func (s *stubWebhookService) ProcessCallback(_ context.Context, _ PaySolutionsWebhookPayload) error {
	s.called = true
	return nil
}

func TestWebhookSignatureVerification(t *testing.T) {
	secret := "super_secret_webhook_key_1234567890"
	svc := &stubWebhookService{}
	h := NewHandler(svc, secret)

	app := fiber.New()
	app.Post("/webhook", h.Handle)

	body := []byte(`{"order_no":"11111111-1111-1111-1111-111111111111","ref_no":"REF1","status":"CP","amount":50}`)

	// Missing signature -> 401
	req1 := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	if resp, _ := app.Test(req1); resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("missing signature got %d, want 401", resp.StatusCode)
	}

	// Bad signature -> 401
	req2 := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-PaySolutions-Signature", "deadbeef")
	if resp, _ := app.Test(req2); resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("bad signature got %d, want 401", resp.StatusCode)
	}

	// Valid HMAC-SHA256 signature -> 200
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req3 := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-PaySolutions-Signature", "sha256="+sig)
	resp3, err := app.Test(req3)
	if err != nil {
		t.Fatalf("valid webhook request err: %v", err)
	}
	if resp3.StatusCode != fiber.StatusOK || !svc.called {
		t.Fatalf("valid signature got %d (called=%v), want 200 (called=true)", resp3.StatusCode, svc.called)
	}
}

