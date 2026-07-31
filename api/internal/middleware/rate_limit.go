package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/api/pkg/response"
)

type rateLimitEntry struct {
	count     int
	firstSeen time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, entry := range rl.entries {
			if time.Since(entry.firstSeen) > rl.window {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.entries[ip]
	if !exists || time.Since(entry.firstSeen) > rl.window {
		rl.entries[ip] = &rateLimitEntry{count: 1, firstSeen: time.Now()}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

// RateLimitMiddleware limits requests per IP within a sliding window
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newIPRateLimiter(limit, window)

	return func(c *gin.Context) {
		if !limiter.allow(c.ClientIP()) {
			response.Error(c, http.StatusTooManyRequests,
				"Terlalu banyak percobaan login. Coba lagi dalam beberapa menit.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
