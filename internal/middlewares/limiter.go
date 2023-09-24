package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter *rate.Limiter
	lastSeen time.Time
}

var rateLimiters = make(map[string]*clientLimiter)
var mtx sync.Mutex

//better to use redis in case theres horizontal scaling
func RateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mtx.Lock()
		limiter, exists := rateLimiters[ip]
		if !exists {
			limiter = &clientLimiter{
				limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(rps)), burst),
				lastSeen: time.Now(),
			}
			rateLimiters[ip] = limiter
		}
		mtx.Unlock()

		if !limiter.limiter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}