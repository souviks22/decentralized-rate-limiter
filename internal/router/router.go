package router

import (
	"github.com/gin-gonic/gin"
	"github.com/souviks22/decentralized-rate-limiter/internal/api"
	"github.com/souviks22/decentralized-rate-limiter/internal/middleware"
)

func Setup() *gin.Engine {
	r := gin.Default()
	// r.Use(middleware.Logger())
	r.Use(middleware.RateLimiter(10, 1))
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/ping", api.PingHandler)
	}
	return r
}