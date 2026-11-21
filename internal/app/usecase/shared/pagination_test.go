package shared

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestParsePageQuery(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		p, l, ok := ParsePageQuery(c)
		if !ok {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.JSON(fiber.Map{"page": p, "limit": l})
	})

	// Case 1: No query params
	req1 := httptest.NewRequest("GET", "/test", nil)
	resp1, err := app.Test(req1)
	if err != nil || resp1.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 when no params, got status=%d, err=%v", resp1.StatusCode, err)
	}

	// Case 2: limit and page
	req2 := httptest.NewRequest("GET", "/test?page=2&limit=15", nil)
	resp2, err := app.Test(req2)
	if err != nil || resp2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got status=%d, err=%v", resp2.StatusCode, err)
	}

	// Case 3: page_size parameter
	req3 := httptest.NewRequest("GET", "/test?page=3&page_size=20", nil)
	resp3, err := app.Test(req3)
	if err != nil || resp3.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for page_size, got status=%d, err=%v", resp3.StatusCode, err)
	}

	// Case 4: pageSize camelCase parameter
	req4 := httptest.NewRequest("GET", "/test?page=1&pageSize=5", nil)
	resp4, err := app.Test(req4)
	if err != nil || resp4.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for pageSize, got status=%d, err=%v", resp4.StatusCode, err)
	}
}
