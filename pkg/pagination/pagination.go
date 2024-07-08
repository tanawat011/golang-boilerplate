package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultLimit is the default number of items per page
	DefaultLimit = 10
	// MaxLimit is the maximum number of items per page
	MaxLimit = 100
)

// GetPaginationParams extracts page and limit from query parameters
func GetPaginationParams(c *gin.Context) (int, int) {
	page := DefaultPage
	limit := DefaultLimit

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	return page, limit
}
