package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	count := 0
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		count++
		log.Printf("[%s] %s %s\n", c.Request.Method, c.Request.URL.Path, duration)
	}
}