package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipRateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.Mutex
	rps      float64
	burst    int
}

func NewIPRateLimiter(rps float64, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(l.rps), l.burst)
		l.visitors[ip] = limiter
	}
	return limiter
}

func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rps, burst)
	return func(c *gin.Context) {
		if !limiter.getLimiter(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
