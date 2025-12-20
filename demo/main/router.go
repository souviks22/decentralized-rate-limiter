package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	drl "github.com/souviks22/decentralized-rate-limiter"
)

func RateLimiter(capacity float64, refillRate float64) gin.HandlerFunc {
	rateLimiter := drl.NewRateLimiter(capacity, refillRate)
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

func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(RateLimiter(10, 1))
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/ping", PingHandler)
	}
	return r
}
