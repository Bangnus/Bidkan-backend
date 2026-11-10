package middleware

import (
	"fmt"
	"math"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/gofiber/fiber/v2"
)

func AuthRateLimiter(store security.SessionStore, scope string, maxReqs int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if store == nil {
			return c.Next()
		}
		key := fmt.Sprintf("%s:%s", scope, c.IP())
		allowed, retryAfter, err := store.AllowRateLimit(c.Context(), key, maxReqs, window)
		if err != nil {
			return c.Next()
		}
		if !allowed {
			secs := int(math.Ceil(retryAfter.Seconds()))
			if secs < 1 {
				secs = 1
			}
			c.Set("Retry-After", fmt.Sprintf("%d", secs))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":               "Too many requests. Please try again later.",
				"retry_after_seconds": secs,
			})
		}
		return c.Next()
	}
}

