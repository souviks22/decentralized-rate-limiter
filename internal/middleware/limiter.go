package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souviks22/decentralized-rate-limiter/internal/limiter"
)

func RateLimiter(capacity float64, refillRate float64) gin.HandlerFunc {
	rateLimiter := limiter.New(capacity, refillRate)
	return func(c *gin.Context) {
		userId := c.GetHeader("Ping-User-Id")
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing Used ID"})
			return
		}
		if !rateLimiter.AllowRequest(userId) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Tokens Exhaused"})
			return
		}
		c.Next()
	}
}
