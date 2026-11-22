package middleware

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/gofiber/fiber/v2"
)

func clientIPApp(secret string) *fiber.App {
	app := fiber.New()
	app.Use(ResolveClientIP(secret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString(ClientIP(c)) })
	return app
}

func getClientIP(t *testing.T, app *fiber.App, headers map[string]string) string {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func TestResolveClientIP(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	app := clientIPApp(secret)
	socketIP := getClientIP(t, app, nil)

	if got := getClientIP(t, app, map[string]string{
		HeaderProxySecret: secret, HeaderForwardedClientIP: "203.0.113.7",
	}); got != "203.0.113.7" {
		t.Fatalf("matching secret: got %q, want forwarded IP", got)
	}
	if got := getClientIP(t, app, map[string]string{
		HeaderProxySecret: "wrong", HeaderForwardedClientIP: "203.0.113.7",
	}); got != socketIP {
		t.Fatalf("wrong secret: got %q, want socket IP %q", got, socketIP)
	}
	if got := getClientIP(t, app, map[string]string{
		HeaderProxySecret: secret, HeaderForwardedClientIP: "not-an-ip",
	}); got != socketIP {
		t.Fatalf("invalid forwarded IP: got %q, want socket IP %q", got, socketIP)
	}
	// ไม่ได้ตั้งรหัสลับ = ไม่เชื่อ header จากหน้าเว็บเลย แม้ส่งรหัสว่างมา
	if got := getClientIP(t, clientIPApp(""), map[string]string{
		HeaderProxySecret: "", HeaderForwardedClientIP: "203.0.113.7",
	}); got != socketIP {
		t.Fatalf("no secret configured: got %q, want socket IP %q", got, socketIP)
	}
}

// ผู้ใช้สองคนที่มาจาก server ของ Vercel ตัวเดียวกัน ต้องได้โควตาแยกกัน
func TestRateLimiterSeparatesForwardedClients(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	app := fiber.New()
	app.Use(ResolveClientIP(secret))
	app.Post("/login", AuthRateLimiter(security.NewMemorySessionStore(), "test_fwd", 1, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	post := func(ip string) int {
		req := httptest.NewRequest("POST", "/login", nil)
		req.Header.Set(HeaderProxySecret, secret)
		req.Header.Set(HeaderForwardedClientIP, ip)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode
	}
	if got := post("203.0.113.1"); got != fiber.StatusOK {
		t.Fatalf("user A first: got %d", got)
	}
	if got := post("203.0.113.2"); got != fiber.StatusOK {
		t.Fatalf("user B first: got %d, want 200 (own bucket)", got)
	}
	if got := post("203.0.113.1"); got != fiber.StatusTooManyRequests {
		t.Fatalf("user A second: got %d, want 429", got)
	}
}
