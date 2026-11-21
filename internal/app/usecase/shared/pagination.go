package shared

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
)

func ParsePageQuery(c *fiber.Ctx) (int, int, bool) {
	if c.Query("page") == "" && c.Query("limit") == "" && c.Query("page_size") == "" && c.Query("pageSize") == "" {
		return 0, 0, false
	}
	limit := c.QueryInt("limit", 0)
	if limit <= 0 {
		limit = c.QueryInt("page_size", 0)
	}
	if limit <= 0 {
		limit = c.QueryInt("pageSize", 10)
	}
	p, l := entity.NormalizePagination(c.QueryInt("page", 1), limit)
	return p, l, true
}

