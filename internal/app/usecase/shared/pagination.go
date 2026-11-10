package shared

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
)

func ParsePageQuery(c *fiber.Ctx) (int, int, bool) {
	if c.Query("page") == "" && c.Query("limit") == "" {
		return 0, 0, false
	}
	p, l := entity.NormalizePagination(c.QueryInt("page", 1), c.QueryInt("limit", 10))
	return p, l, true
}

