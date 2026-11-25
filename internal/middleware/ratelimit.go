package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter limits requests per client IP with a token bucket.
//
// It keeps state in memory, so each instance of the app counts on its own.
// Behind several replicas, use a shared store (e.g. Redis) or limit at the
// load balancer instead. The client IP comes from gin's ClientIP, which only
// trusts X-Forwarded-For from the proxies set in TRUSTED_PROXIES.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*client
	limit    rate.Limit
	burst    int
	ttl      time.Duration
	lastScan time.Time
	now      func() time.Time
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter allows perMinute requests per minute per IP, with bursts of
// up to perMinute requests.
func NewRateLimiter(perMinute int) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*client),
		limit:   rate.Limit(float64(perMinute) / 60),
		burst:   perMinute,
		ttl:     10 * time.Minute,
		now:     time.Now,
	}
}

// Middleware returns the Gin handler.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reservation := rl.reserve(c.ClientIP())
		if delay := reservation.DelayFrom(rl.now()); delay > 0 {
			reservation.CancelAt(rl.now())
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(delay.Seconds()))))
			abort(c, http.StatusTooManyRequests, "Too many requests, please try again later")
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter) reserve(ip string) *rate.Reservation {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	if now.Sub(rl.lastScan) > rl.ttl {
		for k, v := range rl.clients {
			if now.Sub(v.lastSeen) > rl.ttl {
				delete(rl.clients, k)
			}
		}
		rl.lastScan = now
	}

	cl, ok := rl.clients[ip]
	if !ok {
		cl = &client{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.clients[ip] = cl
	}
	cl.lastSeen = now
	return cl.limiter.ReserveN(now, 1)
}
