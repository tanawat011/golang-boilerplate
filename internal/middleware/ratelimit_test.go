package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterPerIPAndRefill(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl := NewRateLimiter(2)
	rl.now = func() time.Time { return now }

	r := gin.New()
	r.GET("/", rl.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	hit := func(ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	for i := range 2 {
		if code := hit("10.0.0.1").Code; code != http.StatusNoContent {
			t.Fatalf("request %d: %d", i+1, code)
		}
	}
	rec := hit("10.0.0.1")
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") != "30" {
		t.Fatalf("over the limit: %d, Retry-After %q", rec.Code, rec.Header().Get("Retry-After"))
	}
	if code := hit("10.0.0.2").Code; code != http.StatusNoContent {
		t.Errorf("another IP was limited: %d", code)
	}

	// A refused request does not use up a token: after 30s one is back.
	now = now.Add(30 * time.Second)
	if code := hit("10.0.0.1").Code; code != http.StatusNoContent {
		t.Errorf("after refill: %d", code)
	}

	// Idle clients are forgotten.
	now = now.Add(time.Hour)
	hit("10.0.0.3")
	if _, ok := rl.clients["10.0.0.1"]; ok {
		t.Error("idle client not cleaned up")
	}
}
