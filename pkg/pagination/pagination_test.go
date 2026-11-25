package pagination

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := map[string][2]int{
		"":                   {DefaultPage, DefaultLimit},
		"?page=3&limit=25":   {3, 25},
		"?page=0&limit=-5":   {DefaultPage, DefaultLimit},
		"?page=abc&limit=xy": {DefaultPage, DefaultLimit},
		"?limit=1000":        {DefaultPage, MaxLimit},
	}
	for query, want := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/"+query, nil)
		page, limit := GetPaginationParams(c)
		if page != want[0] || limit != want[1] {
			t.Errorf("%q: got (%d, %d), want (%d, %d)", query, page, limit, want[0], want[1])
		}
	}
}
