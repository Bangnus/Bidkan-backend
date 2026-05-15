package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. ดึง Token จาก Header "Authorization"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		// บังคับตามมาตรฐาน RFC 6750: ต้องขึ้นต้นด้วย "Bearer " เท่านั้น
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization format. Use 'Bearer <token>'",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. ตรวจสอบความถูกต้องของ Token
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "bidkan_secret_key_2024"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// ตรวจสอบ Signing Method เพื่อป้องกันปัญหาระดับต่ำ (Alg: None Attack)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// 3. ดึงข้อมูลจาก Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			c.Locals("user_id", claims["user_id"])
			c.Locals("role", claims["role"])
		}

		return c.Next()
	}
}
