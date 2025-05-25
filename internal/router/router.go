package router

import (
	"github.com/gin-gonic/gin"
	"github.com/souviks22/decentralized-rate-limiter/internal/api"
)

func Setup() *gin.Engine {
	r := gin.Default()
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/ping", api.PingHandler)
	}
	return r
}